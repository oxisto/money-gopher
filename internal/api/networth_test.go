package api

import "testing"

// TestNetWorthHistory checks the sampling and that the last point, which
// lands on now, agrees with the current balances.
func TestNetWorthHistory(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)

	exec(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: DEPOSIT_CASH, time: "2026-01-02T10:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: 100000
		}) { id }
	}`, map[string]any{"account": accountID})
	exec(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: INTEREST, time: "2026-03-31T10:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: 1250
		}) { id }
	}`, map[string]any{"account": accountID})

	data = exec(t, schema, ctx, `query {
		netWorthHistory(period: ALL_TIME, points: 5) {
			time netWorth { amount currency } invested { amount }
		}
	}`, nil)
	points := data["netWorthHistory"].([]any)
	if len(points) != 5 {
		t.Fatalf("points = %d, want 5", len(points))
	}

	first := points[0].(map[string]any)
	if got := first["netWorth"].(map[string]any)["amount"].(float64); got != 100000 {
		t.Errorf("first netWorth = %v, want 100000 (the deposit, at the first transaction)", got)
	}

	last := points[4].(map[string]any)
	if got := last["netWorth"].(map[string]any)["amount"].(float64); got != 101250 {
		t.Errorf("last netWorth = %v, want 101250", got)
	}
	if got := last["invested"].(map[string]any)["amount"].(float64); got != 100000 {
		t.Errorf("last invested = %v, want 100000 (interest is earned, not invested)", got)
	}

	execExpectError(t, schema, ctx, `query { netWorthHistory(period: ONE_YEAR, points: 1) { time } }`, nil)
}
