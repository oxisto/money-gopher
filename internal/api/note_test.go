package api

import "testing"

// TestTransactionNote: a note is stored trimmed, an empty string removes
// it, and an update without the field leaves it alone.
func TestTransactionNote(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	giro := h.account("Giro", "CASH", nil)

	data := exec(t, schema, ctx, `mutation ($acct: ID!) {
		createTransaction(input: { type: DEPOSIT_CASH, time: "2026-10-28T12:00:00Z", currency: "EUR", cashAccountID: $acct, cashDelta: 500000, note: "  Gehalt Oktober " }) { id note }
	}`, map[string]any{"acct": giro})
	tx := data["createTransaction"].(map[string]any)
	if tx["note"] != "Gehalt Oktober" {
		t.Errorf("note = %v, want it trimmed", tx["note"])
	}
	id := tx["id"].(string)

	data = exec(t, schema, ctx, `mutation ($id: ID!) { updateTransaction(id: $id, input: { cashDelta: 510000 }) { note } }`, map[string]any{"id": id})
	if data["updateTransaction"].(map[string]any)["note"] != "Gehalt Oktober" {
		t.Errorf("an update without the field changed the note: %v", data["updateTransaction"])
	}
	data = exec(t, schema, ctx, `mutation ($id: ID!) { updateTransaction(id: $id, input: { note: "" }) { note } }`, map[string]any{"id": id})
	if data["updateTransaction"].(map[string]any)["note"] != nil {
		t.Errorf("an empty string did not remove the note: %v", data["updateTransaction"])
	}
}
