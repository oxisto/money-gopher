package api

import (
	"testing"
)

// TestPortfolioReturns: the four parts of a period's return add up, the
// all-time value change is the unrealized gain, the years sum to all
// time, and every period parses.
func TestPortfolioReturns(t *testing.T) {
	schema, ctx := newTestSchema(t)
	giro := exec(t, schema, ctx, `mutation { createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id } }`, nil)["createCashAccount"].(map[string]any)["id"].(string)
	portfolio := exec(t, schema, ctx, `mutation ($acct: ID!) { createPortfolio(input: { displayName: "Depot", cashAccountID: $acct }) { id } }`, map[string]any{"acct": giro})["createPortfolio"].(map[string]any)["id"].(string)
	asset := exec(t, schema, ctx, `mutation { createAsset(input: { displayName: "Muster AG", identifiers: [{ kind: "ISIN", value: "DE0000000001" }], listings: [{ ticker: "MST", exchange: "XETR", currency: "EUR" }] }) { id } }`, nil)["createAsset"].(map[string]any)["id"].(string)

	tx := func(q string) {
		exec(t, schema, ctx, `mutation ($p: ID!, $a: ID!, $c: ID!) { createTransaction(input: { portfolioID: $p, assetID: $a, cashAccountID: $c, currency: "EUR", `+q+` }) { id } }`,
			map[string]any{"p": portfolio, "a": asset, "c": giro})
	}
	// Two years ago: 10 bought at 100 with a fee; last year: a dividend
	// and half sold at 120; this year: nothing. No quotes: the last trade
	// prices the rest.
	tx(`type: BUY, time: "2024-03-01T10:00:00Z", units: 10, price: "100", fees: 1000`)
	tx(`type: DIVIDEND, time: "2025-05-10T10:00:00Z", taxes: 500, cashDelta: 1500`)
	tx(`type: SELL, time: "2025-06-01T10:00:00Z", units: 5, price: "120", fees: 200`)

	const fields = `period from to realized { amount } income { amount } costs { amount } valueChange { amount } total { amount } timeWeightedReturn`
	data := exec(t, schema, ctx, `query ($id: ID!) { portfolio(id: $id) { snapshot {
		totalProfitOrLoss { amount }
		all: returns(period: ALL_TIME) { `+fields+` }
		ytd: returns(period: YEAR_TO_DATE) { `+fields+` }
		week: returns(period: ONE_WEEK) { period }
		month: returns(period: ONE_MONTH) { period }
		six: returns(period: SIX_MONTHS) { period }
		year: returns(period: ONE_YEAR) { period }
		returnsByYear { `+fields+` }
	} } }`, map[string]any{"id": portfolio})
	snap := data["portfolio"].(map[string]any)["snapshot"].(map[string]any)
	amount := func(r map[string]any, k string) float64 { return r[k].(map[string]any)["amount"].(float64) }
	all := snap["all"].(map[string]any)
	if all["period"] != "ALL_TIME" {
		t.Errorf("period = %v", all["period"])
	}
	// Realized 5 × (120 − 100) = 100 gross; income 20 gross; costs 10 + 5 + 2.
	if amount(all, "realized") != 10000 || amount(all, "income") != 2000 || amount(all, "costs") != 1700 {
		t.Errorf("all time = %v", all)
	}
	if parts := amount(all, "realized") + amount(all, "income") - amount(all, "costs") + amount(all, "valueChange"); parts != amount(all, "total") {
		t.Errorf("parts %v != total %v", parts, amount(all, "total"))
	}
	if amount(all, "valueChange") != amount(snap, "totalProfitOrLoss") {
		t.Errorf("all-time value change %v != unrealized gain %v", amount(all, "valueChange"), amount(snap, "totalProfitOrLoss"))
	}
	years := snap["returnsByYear"].([]any)
	if len(years) < 3 {
		t.Fatalf("%d years, want 2024 to this year", len(years))
	}
	var sum float64
	for _, y := range years {
		yr := y.(map[string]any)
		if yr["period"] != nil {
			t.Errorf("a year carries a period: %v", yr["period"])
		}
		sum += amount(yr, "total")
	}
	if sum != amount(all, "total") {
		t.Errorf("years sum to %v, all time is %v", sum, amount(all, "total"))
	}
	if ytd := snap["ytd"].(map[string]any); amount(ytd, "realized") != 0 || amount(ytd, "income") != 0 {
		t.Errorf("this year has flows: %v", ytd)
	}
	for _, k := range []string{"week", "month", "six", "year"} {
		if snap[k] == nil {
			t.Errorf("%s returns nil", k)
		}
	}
}
