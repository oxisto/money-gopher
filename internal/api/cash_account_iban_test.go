package api

import "testing"

// TestCashAccountIBAN: the IBAN is stored normalized however it was
// typed, can be changed later, an empty string removes it, and a
// name-only update leaves it alone.
func TestCashAccountIBAN(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR", iban: "de58 5001 0517 5437 6844 18" }) { id iban }
	}`, nil)
	account := data["createCashAccount"].(map[string]any)
	if account["iban"] != "DE58500105175437684418" {
		t.Errorf("created iban = %v, want DE58500105175437684418", account["iban"])
	}

	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		updateCashAccount(id: $id, input: { displayName: "Girokonto" }) { displayName iban }
	}`, map[string]any{"id": account["id"]})
	if got := data["updateCashAccount"].(map[string]any); got["displayName"] != "Girokonto" || got["iban"] != "DE58500105175437684418" {
		t.Errorf("name-only update = %v, want the IBAN kept", got)
	}

	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		updateCashAccount(id: $id, input: { iban: "DE46 5001 0517 5802 2452 63" }) { iban }
	}`, map[string]any{"id": account["id"]})
	if got := data["updateCashAccount"].(map[string]any)["iban"]; got != "DE46500105175802245263" {
		t.Errorf("changed iban = %v, want DE46500105175802245263", got)
	}

	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		updateCashAccount(id: $id, input: { iban: "" }) { iban }
	}`, map[string]any{"id": account["id"]})
	if got := data["updateCashAccount"].(map[string]any)["iban"]; got != nil {
		t.Errorf("cleared iban = %v, want null", got)
	}
}
