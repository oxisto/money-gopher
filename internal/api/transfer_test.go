package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// transferHelpers wraps the mutations a transfer scenario is made of.
type transferHelpers struct {
	t      *testing.T
	schema *graphql.Schema
	ctx    context.Context
}

func (h transferHelpers) account(name, kind string, declared *int) string {
	h.t.Helper()
	q := fmt.Sprintf(`mutation { createCashAccount(input: { displayName: %q, currency: "EUR", kind: %s`, name, kind)
	if declared != nil {
		q += fmt.Sprintf(`, declaredBalance: %d`, *declared)
	}
	q += ` }) { id } }`
	return exec(h.t, h.schema, h.ctx, q, nil)["createCashAccount"].(map[string]any)["id"].(string)
}

func (h transferHelpers) transfer(from, to string, amount int, day string) map[string]any {
	h.t.Helper()
	data := exec(h.t, h.schema, h.ctx, `mutation ($from: ID!, $to: ID!, $amount: Int!, $time: Time!) {
		createTransfer(input: { time: $time, fromCashAccountID: $from, toCashAccountID: $to, amount: $amount, currency: "EUR" }) {
			id amount { amount }
			from { id type cashDelta { amount } transferID counterpart { id } }
			to { id type cashDelta { amount } transferID counterpart { id } }
		}
	}`, map[string]any{"from": from, "to": to, "amount": amount, "time": day + "T12:00:00Z"})
	return data["createTransfer"].(map[string]any)
}

func (h transferHelpers) interest(account string, amount int, day string) {
	h.t.Helper()
	exec(h.t, h.schema, h.ctx, `mutation ($acct: ID!, $amount: Int!, $time: Time!) {
		createTransaction(input: { type: INTEREST, time: $time, currency: "EUR", cashAccountID: $acct, cashDelta: $amount }) { id }
	}`, map[string]any{"acct": account, "amount": amount, "time": day + "T12:00:00Z"})
}

func (h transferHelpers) balance(account string) float64 {
	h.t.Helper()
	data := exec(h.t, h.schema, h.ctx, `query ($id: ID!) { cashAccount(id: $id) { balance { amount } } }`, map[string]any{"id": account})
	return data["cashAccount"].(map[string]any)["balance"].(map[string]any)["amount"].(float64)
}

func TestCreateTransfer(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	giro := h.account("Giro", "CASH", nil)
	tagesgeld := h.account("Tagesgeld", "SAVINGS", nil)

	tr := h.transfer(giro, tagesgeld, 50000, "2026-02-01")
	from := tr["from"].(map[string]any)
	to := tr["to"].(map[string]any)
	if from["type"] != "WITHDRAW_CASH" || from["cashDelta"].(map[string]any)["amount"] != -50000.0 {
		t.Errorf("from leg = %v, want a WITHDRAW_CASH of -50000", from)
	}
	if to["type"] != "DEPOSIT_CASH" || to["cashDelta"].(map[string]any)["amount"] != 50000.0 {
		t.Errorf("to leg = %v, want a DEPOSIT_CASH of 50000", to)
	}
	if from["transferID"] != tr["id"] || to["transferID"] != tr["id"] {
		t.Errorf("legs carry %v / %v, want the transfer's id %v", from["transferID"], to["transferID"], tr["id"])
	}
	if from["counterpart"].(map[string]any)["id"] != to["id"] || to["counterpart"].(map[string]any)["id"] != from["id"] {
		t.Errorf("counterparts do not point at each other: %v / %v", from["counterpart"], to["counterpart"])
	}
	if h.balance(giro) != -50000 || h.balance(tagesgeld) != 50000 {
		t.Errorf("balances = %v / %v, want -50000 / 50000", h.balance(giro), h.balance(tagesgeld))
	}

	// Validation: same account, nothing, a foreign account, another currency.
	for name, vars := range map[string]map[string]any{
		"same account": {"from": giro, "to": giro, "amount": 100},
		"zero":         {"from": giro, "to": tagesgeld, "amount": 0},
		"unknown":      {"from": giro, "to": "nope", "amount": 100},
	} {
		t.Run(name, func(t *testing.T) {
			vars["time"] = "2026-02-02T12:00:00Z"
			execExpectError(t, schema, ctx, `mutation ($from: ID!, $to: ID!, $amount: Int!, $time: Time!) {
				createTransfer(input: { time: $time, fromCashAccountID: $from, toCashAccountID: $to, amount: $amount, currency: "EUR" }) { id }
			}`, vars)
		})
	}
	usd := exec(t, schema, ctx, `mutation { createCashAccount(input: { displayName: "USD", currency: "USD" }) { id } }`, nil)["createCashAccount"].(map[string]any)["id"].(string)
	execExpectError(t, schema, ctx, `mutation ($from: ID!, $to: ID!) {
		createTransfer(input: { time: "2026-02-02T12:00:00Z", fromCashAccountID: $from, toCashAccountID: $to, amount: 100, currency: "EUR" }) { id }
	}`, map[string]any{"from": giro, "to": usd})
	other := auth.WithPerson(auth.WithUser(context.Background(), &persistence.User{ID: "other"}), &persistence.Person{ID: "other"})
	execExpectError(t, schema, other, `mutation ($from: ID!, $to: ID!) {
		createTransfer(input: { time: "2026-02-02T12:00:00Z", fromCashAccountID: $from, toCashAccountID: $to, amount: 100, currency: "EUR" }) { id }
	}`, map[string]any{"from": giro, "to": tagesgeld})

	// Editing one leg mirrors onto the other; the type cannot change.
	exec(t, schema, ctx, `mutation ($id: ID!) {
		updateTransaction(id: $id, input: { cashDelta: -60000, time: "2026-02-03T12:00:00Z" }) { id }
	}`, map[string]any{"id": from["id"]})
	data := exec(t, schema, ctx, `query ($id: ID!) { transaction(id: $id) { cashDelta { amount } time } }`, map[string]any{"id": to["id"]})
	if got := data["transaction"].(map[string]any); got["cashDelta"].(map[string]any)["amount"] != 60000.0 || got["time"] != "2026-02-03T12:00:00Z" {
		t.Errorf("counterpart after edit = %v, want 60000 on 2026-02-03", got)
	}
	execExpectError(t, schema, ctx, `mutation ($id: ID!) { updateTransaction(id: $id, input: { type: INTEREST }) { id } }`, map[string]any{"id": from["id"]})

	// Deleting a leg deletes both.
	exec(t, schema, ctx, `mutation ($id: ID!) { deleteTransaction(id: $id) }`, map[string]any{"id": to["id"]})
	data = exec(t, schema, ctx, `query ($id: ID!) { transaction(id: $id) { id } }`, map[string]any{"id": from["id"]})
	if data["transaction"] != nil {
		t.Errorf("the other leg survived the delete: %v", data["transaction"])
	}
	if h.balance(giro) != 0 || h.balance(tagesgeld) != 0 {
		t.Errorf("balances after delete = %v / %v, want 0 / 0", h.balance(giro), h.balance(tagesgeld))
	}
}

// TestTransfer_InvestedUnchanged: moving money between two tracked
// accounts is neither a deposit nor a withdrawal for net worth.
func TestTransfer_InvestedUnchanged(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	giro := h.account("Giro", "CASH", nil)
	tagesgeld := h.account("Tagesgeld", "SAVINGS", nil)
	exec(t, schema, ctx, `mutation ($acct: ID!) {
		createTransaction(input: { type: DEPOSIT_CASH, time: "2026-01-02T12:00:00Z", currency: "EUR", cashAccountID: $acct, cashDelta: 100000 }) { id }
	}`, map[string]any{"acct": giro})
	h.transfer(giro, tagesgeld, 50000, "2026-01-05")

	data := exec(t, schema, ctx, `{ netWorthHistory(period: ALL_TIME, points: 2) { netWorth { amount } invested { amount } } }`, nil)
	last := data["netWorthHistory"].([]any)[1].(map[string]any)
	if last["netWorth"].(map[string]any)["amount"] != 100000.0 || last["invested"].(map[string]any)["amount"] != 100000.0 {
		t.Errorf("after a transfer: %v, want net worth and invested both 100000", last)
	}
}

// TestBausparvorfinanzierung walks the user's house financing: an
// endfälliges Darlehen (interest only), a Bausparvertrag saved into
// monthly, and the Zuteilung that repays the loan partly from the savings
// and partly through a new, amortising Bauspardarlehen. The Giro is an
// estimate throughout: money flows through it without moving it.
func TestBausparvorfinanzierung(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}

	declared := 150000
	giro := h.account("Giro", "CASH", &declared)
	loan := h.account("Vorfinanzierung", "LOAN", nil)
	bauspar := h.account("Bausparvertrag", "SAVINGS", nil)

	// The room, booked at cost into the assets portfolio settling on the Giro.
	exec(t, schema, ctx, `mutation ($acct: ID!) {
		bookAsset(input: { displayName: "Office room", kind: PROPERTY, currency: "EUR", acquiredAt: "2023-01-02T12:00:00Z", cost: "200000", cashAccountID: $acct }) { id }
	}`, map[string]any{"acct": giro})

	// Disbursed: the loan goes negative, the Giro (an estimate) stays put.
	h.transfer(loan, giro, 19000000, "2023-01-02")
	if h.balance(loan) != -19000000 || h.balance(giro) != 150000 {
		t.Fatalf("after disbursement: loan %v, giro %v; want -19000000 and the declared 150000", h.balance(loan), h.balance(giro))
	}

	// Two years: interest charged on the loan and paid from the Giro; the
	// Sparrate into the Bausparvertrag; interest credited there each year.
	for year := 2023; year <= 2024; year++ {
		for month := 1; month <= 12; month++ {
			day := fmt.Sprintf("%d-%02d-28", year, month)
			h.interest(loan, -47500, day)
			h.transfer(giro, loan, 47500, day)
			h.transfer(giro, bauspar, 50000, day)
		}
		h.interest(bauspar, 3000, fmt.Sprintf("%d-12-31", year))
	}
	if h.balance(loan) != -19000000 {
		t.Errorf("an interest-only loan after two years = %v, want still -19000000", h.balance(loan))
	}
	if h.balance(bauspar) != 1206000 {
		t.Errorf("Bausparvertrag = %v, want 1206000 (24 × 500 + 2 × 30)", h.balance(bauspar))
	}

	// Zuteilung: the savings repay part of the loan, a new Bauspardarlehen
	// repays the rest; the Vorfinanzierung and the Bausparvertrag close.
	h.transfer(bauspar, loan, 1206000, "2025-01-10")
	bauspardarlehen := h.account("Bauspardarlehen", "LOAN", nil)
	h.transfer(bauspardarlehen, loan, 17794000, "2025-01-10")
	if h.balance(loan) != 0 || h.balance(bauspar) != 0 || h.balance(bauspardarlehen) != -17794000 {
		t.Errorf("after Zuteilung: loan %v, Bausparvertrag %v, Bauspardarlehen %v; want 0, 0, -17794000", h.balance(loan), h.balance(bauspar), h.balance(bauspardarlehen))
	}

	// The new loan amortises: interest charged, a larger payment made.
	h.interest(bauspardarlehen, -30000, "2025-02-28")
	h.transfer(giro, bauspardarlehen, 90000, "2025-02-28")
	if h.balance(bauspardarlehen) != -17734000 {
		t.Errorf("Bauspardarlehen after the first payment = %v, want -17734000", h.balance(bauspardarlehen))
	}

	// Net worth: the room at cost plus the declared Giro, minus what is owed.
	data := exec(t, schema, ctx, `{
		netWorthHistory(period: ALL_TIME, points: 2) { netWorth { amount } estimated { amount } }
		realizedGains(from: "2024-01-01T00:00:00Z", to: "2025-01-01T00:00:00Z") { interest { amount } }
	}`, nil)
	last := data["netWorthHistory"].([]any)[1].(map[string]any)
	if got := last["netWorth"].(map[string]any)["amount"]; got != 20000000.0+150000.0-17734000.0 {
		t.Errorf("net worth = %v, want 2416000 (room 200,000 + Giro 1,500 − 177,340 owed)", got)
	}
	if got := last["estimated"].(map[string]any)["amount"]; got != 150000.0 {
		t.Errorf("estimated part = %v, want 150000", got)
	}
	if got := data["realizedGains"].(map[string]any)["interest"].(map[string]any)["amount"]; got != 3000.0 {
		t.Errorf("2024 interest income = %v, want 3000 (the savings' credit; the loan's charges are costs)", got)
	}
}
