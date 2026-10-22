package api

import "testing"

// TestCashAccountKinds: accounts default to CASH; on a loan, interest is
// charged (negative) rather than earned, and it is a cost rather than
// income on the Income page.
func TestCashAccountKinds(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		giro: createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id kind estimated balance { amount } }
		loan: createCashAccount(input: { displayName: "Vorfinanzierung", currency: "EUR", kind: LOAN }) { id kind }
	}`, nil)
	giro := data["giro"].(map[string]any)
	loan := data["loan"].(map[string]any)
	if giro["kind"] != "CASH" || giro["estimated"] != false {
		t.Errorf("giro = %v, want a tracked CASH account", giro)
	}
	if loan["kind"] != "LOAN" {
		t.Errorf("loan kind = %v, want LOAN", loan["kind"])
	}

	// Interest charged on the loan is negative; earned on the Giro positive.
	execExpectError(t, schema, ctx, `mutation ($acct: ID!) {
		createTransaction(input: { type: INTEREST, time: "2026-01-31T00:00:00Z", currency: "EUR", cashAccountID: $acct, cashDelta: 47500 }) { id }
	}`, map[string]any{"acct": loan["id"]})
	execExpectError(t, schema, ctx, `mutation ($acct: ID!) {
		createTransaction(input: { type: INTEREST, time: "2026-01-31T00:00:00Z", currency: "EUR", cashAccountID: $acct, cashDelta: -100 }) { id }
	}`, map[string]any{"acct": giro["id"]})
	exec(t, schema, ctx, `mutation ($loan: ID!, $giro: ID!) {
		charged: createTransaction(input: { type: INTEREST, time: "2026-01-31T00:00:00Z", currency: "EUR", cashAccountID: $loan, cashDelta: -47500 }) { id }
		earned: createTransaction(input: { type: INTEREST, time: "2026-01-31T00:00:00Z", currency: "EUR", cashAccountID: $giro, cashDelta: 1200 }) { id }
	}`, map[string]any{"loan": loan["id"], "giro": giro["id"]})

	data = exec(t, schema, ctx, `query ($loan: ID!) {
		cashAccount(id: $loan) { balance { amount } }
		realizedGains(from: "2026-01-01T00:00:00Z", to: "2027-01-01T00:00:00Z") { interest { amount } }
	}`, map[string]any{"loan": loan["id"]})
	if got := data["cashAccount"].(map[string]any)["balance"].(map[string]any)["amount"]; got != -47500.0 {
		t.Errorf("loan balance = %v, want -47500", got)
	}
	if got := data["realizedGains"].(map[string]any)["interest"].(map[string]any)["amount"]; got != 1200.0 {
		t.Errorf("realized interest = %v, want 1200 (the loan's charge is a cost, not income)", got)
	}

	// Changing the kind later works through updateCashAccount.
	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		updateCashAccount(id: $id, input: { kind: SAVINGS }) { kind }
	}`, map[string]any{"id": giro["id"]})
	if got := data["updateCashAccount"].(map[string]any)["kind"]; got != "SAVINGS" {
		t.Errorf("kind after update = %v, want SAVINGS", got)
	}
}

// TestEstimatedAccount: a declared balance is the balance; transactions on
// the account still exist but do not move it; the ledger stays visible;
// the account can go back to being tracked.
func TestEstimatedAccount(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR", declaredBalance: 150000 }) {
			id estimated declaredBalance { amount } declaredAt balance { amount } ledgerBalance { amount }
		}
		createAsset(input: { displayName: "Vanguard FTSE All-World", listings: [{ ticker: "VWCE", currency: "EUR" }] }) { id }
	}`, nil)
	giro := data["createCashAccount"].(map[string]any)
	assetID := data["createAsset"].(map[string]any)["id"].(string)
	if giro["estimated"] != true || giro["balance"].(map[string]any)["amount"] != 150000.0 || giro["ledgerBalance"].(map[string]any)["amount"] != 0.0 || giro["declaredAt"] == nil {
		t.Fatalf("created = %v, want estimated at 150000 with an empty ledger", giro)
	}

	// A purchase paid from the Giro: real for the portfolio, invisible to
	// the declared balance, visible in the ledger.
	data = exec(t, schema, ctx, `mutation ($acct: ID!) {
		createPortfolio(input: { displayName: "Depot", cashAccountID: $acct }) { id }
	}`, map[string]any{"acct": giro["id"]})
	portfolioID := data["createPortfolio"].(map[string]any)["id"].(string)
	exec(t, schema, ctx, `mutation ($portfolio: ID!, $asset: ID!, $acct: ID!) {
		createTransaction(input: { type: BUY, time: "2026-01-03T10:00:00Z", currency: "EUR", portfolioID: $portfolio, assetID: $asset, cashAccountID: $acct, units: 10, price: "100.00" }) { id }
	}`, map[string]any{"portfolio": portfolioID, "asset": assetID, "acct": giro["id"]})
	account := func() map[string]any {
		data := exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { estimated balance { amount } ledgerBalance { amount } declaredBalance { amount } } }`, map[string]any{"id": giro["id"]})
		return data["cashAccount"].(map[string]any)
	}
	a := account()
	if a["balance"].(map[string]any)["amount"] != 150000.0 || a["ledgerBalance"].(map[string]any)["amount"] != -100000.0 {
		t.Errorf("after a purchase: %v, want balance 150000 and ledger -100000", a)
	}

	// Net worth: the declared 1,500 plus the shares, with the purchase as
	// capital brought in from outside.
	data = exec(t, schema, ctx, `{ netWorthHistory(period: ALL_TIME, points: 2) { netWorth { amount } invested { amount } estimated { amount } } }`, nil)
	last := data["netWorthHistory"].([]any)[1].(map[string]any)
	if last["netWorth"].(map[string]any)["amount"] != 250000.0 || last["invested"].(map[string]any)["amount"] != 100000.0 || last["estimated"].(map[string]any)["amount"] != 150000.0 {
		t.Errorf("net worth point = %v, want 250000 / invested 100000 / estimated 150000", last)
	}

	// Re-declaring moves the declaration; switching off returns to the ledger.
	exec(t, schema, ctx, `mutation ($id: ID!) { updateCashAccount(id: $id, input: { declaredBalance: 120000 }) { id } }`, map[string]any{"id": giro["id"]})
	if a := account(); a["balance"].(map[string]any)["amount"] != 120000.0 {
		t.Errorf("after re-declaring: %v, want 120000", a)
	}
	exec(t, schema, ctx, `mutation ($id: ID!) { updateCashAccount(id: $id, input: { estimated: false }) { id } }`, map[string]any{"id": giro["id"]})
	if a := account(); a["estimated"] != false || a["balance"].(map[string]any)["amount"] != -100000.0 || a["declaredBalance"] != nil {
		t.Errorf("after switching off: %v, want the ledger balance and no declaration", a)
	}
	// Back on without a balance to declare is an error.
	execExpectError(t, schema, ctx, `mutation ($id: ID!) { updateCashAccount(id: $id, input: { estimated: true }) { id } }`, map[string]any{"id": giro["id"]})
}
