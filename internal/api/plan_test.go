package api

import (
	"fmt"
	"math"
	"testing"
	"time"
)

const planFields = `id displayName start horizonMonths annualCost
	startingAmount { amount currency } startingAmountDeclared contribution { amount } contributionDeclared
	assumptions { bucket value { amount } weight weightDeclared annualReturn volatility }
	status { month expected { amount } actual { amount } gap { amount } }`

func createPlan(t *testing.T, h transferHelpers, input string) map[string]any {
	t.Helper()
	data := exec(t, h.schema, h.ctx, fmt.Sprintf(`mutation { createPlan(input: { %s }) { %s } }`, input, planFields), nil)
	return data["createPlan"].(map[string]any)
}

func amount(m map[string]any, field string) float64 {
	return m[field].(map[string]any)["amount"].(float64)
}

func assumption(p map[string]any, bucket string) map[string]any {
	for _, a := range p["assumptions"].([]any) {
		if a.(map[string]any)["bucket"] == bucket {
			return a.(map[string]any)
		}
	}
	return nil
}

// household sets up a ledger Giro with an ETF portfolio, an estimated
// Giro, a savings account, a loan, and the schedules of a salary, a card
// bill and a Sparrate. It returns the net worth now as netWorthHistory
// reports it.
func household(t *testing.T, h transferHelpers) float64 {
	t.Helper()
	giro := h.account("Depot-Giro", "CASH", nil)
	declared := 150000
	estimated := h.account("Giro", "CASH", &declared)
	savings := h.account("Bausparvertrag", "SAVINGS", nil)
	loan := h.account("Vorfinanzierung", "LOAN", nil)

	data := exec(t, h.schema, h.ctx, `mutation ($account: ID!) {
		createPortfolio(input: { displayName: "Depot", cashAccountID: $account }) { id }
	}`, map[string]any{"account": giro})
	portfolio := data["createPortfolio"].(map[string]any)["id"].(string)
	data = exec(t, h.schema, h.ctx, `mutation {
		etf: createAsset(input: { displayName: "World ETF", kind: ETF, listings: [{ ticker: "WRLD", currency: "EUR" }] }) { id }
		stock: createAsset(input: { displayName: "A share", kind: STOCK, listings: [{ ticker: "SHR", currency: "EUR" }] }) { id }
	}`, nil)
	etf := data["etf"].(map[string]any)["id"].(string)
	stock := data["stock"].(map[string]any)["id"].(string)

	exec(t, h.schema, h.ctx, `mutation ($account: ID!) {
		createTransaction(input: { type: DEPOSIT_CASH, time: "2026-01-02T10:00:00Z", currency: "EUR", cashAccountID: $account, cashDelta: 1000000 }) { id }
	}`, map[string]any{"account": giro})
	for _, buy := range []struct {
		asset string
		units int
	}{{etf, 30}, {stock, 10}} {
		exec(t, h.schema, h.ctx, `mutation ($portfolio: ID!, $asset: ID!, $account: ID!, $units: Float!) {
			createTransaction(input: { type: BUY, time: "2026-01-03T10:00:00Z", currency: "EUR", portfolioID: $portfolio, assetID: $asset, cashAccountID: $account, units: $units, price: "100.00" }) { id }
		}`, map[string]any{"portfolio": portfolio, "asset": buy.asset, "account": giro, "units": buy.units})
	}
	// Savings and the loan, as the bank books them.
	h.transfer(estimated, savings, 50000, "2026-02-01")
	exec(t, h.schema, h.ctx, `mutation ($account: ID!) {
		createTransaction(input: { type: WITHDRAW_CASH, time: "2026-02-01T10:00:00Z", currency: "EUR", cashAccountID: $account, cashDelta: -19000000 }) { id }
	}`, map[string]any{"account": loan})

	createSchedule(t, h, fmt.Sprintf(`displayName: "Gehalt", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 28, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 400000 }]`, estimated))
	createSchedule(t, h, fmt.Sprintf(`displayName: "Kreditkarte", type: WITHDRAW_CASH, interval: MONTHLY, dayOfMonth: 3, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 150000 }]`, estimated))
	createSchedule(t, h, fmt.Sprintf(`displayName: "Sparrate", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 50000 }]`, savings, estimated))

	data = exec(t, h.schema, h.ctx, `query { netWorthHistory(period: ONE_MONTH, points: 2) { netWorth { amount } } }`, nil)
	points := data["netWorthHistory"].([]any)
	return amount(points[len(points)-1].(map[string]any), "netWorth")
}

// TestPlanFromTheBooks: a plan without declarations takes its starting
// amount from the net worth, its contribution from the schedules and its
// allocation from what is held, bucket by bucket.
func TestPlanFromTheBooks(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	netWorth := household(t, h)

	p := createPlan(t, h, `displayName: "Haushalt", start: "2026-01-15T00:00:00Z", horizonMonths: 120`)
	if p["start"].(string)[:10] != "2026-01-01" || p["horizonMonths"] != 120.0 || p["annualCost"] != 0.0 {
		t.Errorf("plan = %v", p)
	}
	if p["startingAmountDeclared"] != false || p["contributionDeclared"] != false {
		t.Errorf("declared flags set: %v", p)
	}
	// The plan starts before the first transaction: only the estimated
	// Giro's declared balance, a constant in the net worth, was there.
	if got := amount(p, "startingAmount"); got != 150000 {
		t.Errorf("startingAmount = %v, want the declared 150000 before the first transaction", got)
	}
	// Salary 4,000 − card 1,500; the Sparrate moves money, it does not leave.
	if got := amount(p, "contribution"); got != 250000 {
		t.Errorf("contribution = %v, want 250000", got)
	}

	var sum float64
	for _, a := range p["assumptions"].([]any) {
		sum += amount(a.(map[string]any), "value")
	}
	if sum != netWorth {
		t.Errorf("allocation sums to %v, net worth is %v", sum, netWorth)
	}
	for bucket, want := range map[string]float64{"ETF": 300000, "STOCK": 100000, "CASH": 600000 + 150000, "SAVINGS": 50000, "DEBT": -19000000, "PROPERTY": 0} {
		if got := amount(assumption(p, bucket), "value"); got != want {
			t.Errorf("%s = %v, want %v", bucket, got, want)
		}
	}
	if got := len(p["assumptions"].([]any)); got != len(planBuckets) {
		t.Errorf("assumptions = %d, want every bucket", got)
	}
	etf := assumption(p, "ETF")
	if etf["annualReturn"] != 0.06 || etf["volatility"] != 0.15 || etf["weightDeclared"] != false {
		t.Errorf("ETF defaults = %v", etf)
	}
	// The default split follows the liquid holdings: ETF 3,000 of
	// 3,000 + 1,000 + 7,500 + 500.
	if got := etf["weight"].(float64); math.Abs(got-0.25) > 1e-9 {
		t.Errorf("ETF weight = %v, want 0.25", got)
	}
	if got := assumption(p, "DEBT")["weight"]; got != 0.0 {
		t.Errorf("DEBT weight = %v, want 0", got)
	}

	status := p["status"].(map[string]any)
	if amount(status, "actual") != netWorth || status["month"].(string)[:7] != time.Now().UTC().Format("2006-01") {
		t.Errorf("status = %v, want the net worth now in the current month", status)
	}
	if amount(status, "gap") != amount(status, "actual")-amount(status, "expected") {
		t.Errorf("gap = %v", status)
	}
}

// TestPlanProjection: the series' shape and a declared plan's figures.
func TestPlanProjection(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	household(t, h)

	p := createPlan(t, h, `displayName: "Fixed", start: "2026-11-01T00:00:00Z", horizonMonths: 24, startingAmount: 1000000, contribution: 100000, annualCost: 0.005,
		assumptions: [{ bucket: ETF, weight: 1, annualReturn: 0.12, volatility: 0 }, { bucket: CASH, annualReturn: 0 }, { bucket: STOCK, annualReturn: 0 }, { bucket: SAVINGS, annualReturn: 0 }, { bucket: DEBT, annualReturn: 0 }]`)
	if amount(p, "startingAmount") != 1000000 || p["startingAmountDeclared"] != true || amount(p, "contribution") != 100000 || p["contributionDeclared"] != true || p["annualCost"] != 0.005 {
		t.Errorf("declared plan = %v", p)
	}
	if got := assumption(p, "ETF"); got["weight"] != 1.0 || got["weightDeclared"] != true || got["annualReturn"] != 0.12 || got["volatility"] != 0.0 {
		t.Errorf("ETF = %v", got)
	}
	if got := assumption(p, "CASH"); got["weight"] != 0.0 || got["annualReturn"] != 0.0 {
		t.Errorf("CASH next to a declared weight = %v, want weight 0 and the declared return", got)
	}

	data := exec(t, schema, ctx, `query ($id: ID!) { plan(id: $id) { projection { month expected { amount } contributed { amount } p10 { amount } p50 { amount } p90 { amount } } } }`, map[string]any{"id": p["id"]})
	points := data["plan"].(map[string]any)["projection"].([]any)
	if len(points) != 25 {
		t.Fatalf("projection has %d points, want 25", len(points))
	}
	first := points[0].(map[string]any)
	if first["month"].(string)[:10] != "2026-11-01" || amount(first, "expected") != 1000000 || amount(first, "contributed") != 0 {
		t.Errorf("first point = %v", first)
	}
	last := points[24].(map[string]any)
	if amount(last, "contributed") != 2400000 {
		t.Errorf("contributed after two years = %v", amount(last, "contributed"))
	}
	// Everything held is re-based onto the declared 10,000 and grows only
	// through the ETF bucket; check the bounds rather than the split:
	// more than the money put in, no more than 11.5 % net on all of it.
	expected := amount(last, "expected")
	if expected <= 3400000 || expected > 1000000*math.Pow(1.115, 2)+2400000*1.115 {
		t.Errorf("expected after two years = %v", expected)
	}
	if amount(last, "p10") != expected || amount(last, "p90") != expected {
		t.Errorf("without volatility the fan is the path: %v", last)
	}

	// A plan starting ahead has no status and no actuals yet.
	ahead := createPlan(t, h, `displayName: "Later", start: "2099-01-01T00:00:00Z", horizonMonths: 12`)
	if ahead["status"] != nil {
		t.Errorf("status ahead of start = %v", ahead["status"])
	}
	data = exec(t, schema, ctx, `query ($id: ID!) { plan(id: $id) { actual { time } } }`, map[string]any{"id": ahead["id"]})
	if got := data["plan"].(map[string]any)["actual"].([]any); len(got) != 0 {
		t.Errorf("actual ahead of start = %v", got)
	}
}

// TestPlanActual: the net worth at start, at each month end since, and now.
func TestPlanActual(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	household(t, h)
	start := time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -3, 0)
	p := createPlan(t, h, fmt.Sprintf(`displayName: "Recent", start: %q, horizonMonths: 12`, start.Format(time.RFC3339)))
	data := exec(t, schema, ctx, `query ($id: ID!) { plan(id: $id) { actual { time netWorth { amount } } } }`, map[string]any{"id": p["id"]})
	points := data["plan"].(map[string]any)["actual"].([]any)
	// start, three month ends, now
	if len(points) != 5 {
		t.Fatalf("actual = %d points, want 5", len(points))
	}
	if got := points[0].(map[string]any)["time"].(string)[:10]; got != start.Format("2006-01-02") {
		t.Errorf("first actual at %s, want the start", got)
	}
	if got := points[1].(map[string]any)["time"].(string)[:10]; got != start.AddDate(0, 1, -1).Format("2006-01-02") {
		t.Errorf("second actual at %s, want the first month's end", got)
	}
}

// TestPlanLifecycle: update, assumptions replaced, validation, isolation,
// delete.
func TestPlanLifecycle(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	household(t, h)
	p := createPlan(t, h, `displayName: "Haushalt"`)
	if p["horizonMonths"] != 240.0 || p["start"].(string)[:7] != time.Now().UTC().Format("2006-01") {
		t.Errorf("defaults = %v", p)
	}

	data := exec(t, schema, ctx, fmt.Sprintf(`mutation ($id: ID!) { updatePlan(id: $id, input: { displayName: "Haus", horizonMonths: 60, startingAmount: 500000, contribution: 0, annualCost: 0.01,
		assumptions: [{ bucket: ETF, annualReturn: 0.08 }] }) { %s } }`, planFields), map[string]any{"id": p["id"]})
	u := data["updatePlan"].(map[string]any)
	if u["displayName"] != "Haus" || u["horizonMonths"] != 60.0 || amount(u, "startingAmount") != 500000 || amount(u, "contribution") != 0 || u["annualCost"] != 0.01 {
		t.Errorf("updated = %v", u)
	}
	if got := assumption(u, "ETF"); got["annualReturn"] != 0.08 || got["volatility"] != 0.15 {
		t.Errorf("ETF after update = %v, want 8 %% with the default volatility", got)
	}
	if got := assumption(u, "STOCK"); got["annualReturn"] != 0.07 {
		t.Errorf("STOCK left out of the update = %v, want the default", got)
	}

	data = exec(t, schema, ctx, fmt.Sprintf(`mutation ($id: ID!) { updatePlan(id: $id, input: { clearStartingAmount: true, clearContribution: true }) { %s } }`, planFields), map[string]any{"id": p["id"]})
	u = data["updatePlan"].(map[string]any)
	if u["startingAmountDeclared"] != false || u["contributionDeclared"] != false || amount(u, "contribution") != 250000 {
		t.Errorf("after clearing = %v", u)
	}

	for _, bad := range []string{
		`displayName: "", horizonMonths: 12`,
		`displayName: "x", horizonMonths: 0`,
		`displayName: "x", horizonMonths: 601`,
		`displayName: "x", annualCost: 0.5`,
		`displayName: "x", assumptions: [{ bucket: ETF, weight: 2 }]`,
		`displayName: "x", assumptions: [{ bucket: ETF, annualReturn: 3 }]`,
		`displayName: "x", assumptions: [{ bucket: ETF, volatility: -1 }]`,
		`displayName: "x", assumptions: [{ bucket: ETF }, { bucket: ETF }]`,
	} {
		execExpectError(t, schema, ctx, fmt.Sprintf(`mutation { createPlan(input: { %s }) { id } }`, bad), nil)
	}

	data = exec(t, schema, ctx, `query { plans { id } }`, nil)
	if got := data["plans"].([]any); len(got) != 1 {
		t.Errorf("plans = %v, want one", got)
	}
	data = exec(t, schema, ctx, `mutation ($id: ID!) { deletePlan(id: $id) }`, map[string]any{"id": p["id"]})
	if data["deletePlan"] != p["id"] {
		t.Errorf("deletePlan = %v", data["deletePlan"])
	}
	execExpectError(t, schema, ctx, `mutation ($id: ID!) { deletePlan(id: $id) }`, map[string]any{"id": p["id"]})
	data = exec(t, schema, ctx, `query ($id: ID!) { plan(id: $id) { id } }`, map[string]any{"id": p["id"]})
	if data["plan"] != nil {
		t.Errorf("deleted plan still readable: %v", data["plan"])
	}
}
