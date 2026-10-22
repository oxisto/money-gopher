package api

import (
	"fmt"
	"testing"
)

// TestCashFlow: the household's month from the schedules, seen from the
// Giro. Salary and rent come in, the card bill and the loan payment go
// out, the Sparrate is saved; the loan's own interest charge is not a
// cash-account movement and does not count twice.
func TestCashFlow(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	declared := 150000
	giro := h.account("Giro", "CASH", &declared)
	bauspar := h.account("Bausparvertrag", "SAVINGS", nil)
	loan := h.account("Vorfinanzierung", "LOAN", nil)

	mk := func(input string) {
		createSchedule(t, h, input)
	}
	mk(fmt.Sprintf(`displayName: "Gehalt", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 28, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 485000 }]`, giro))
	mk(fmt.Sprintf(`displayName: "Kreditkarte", type: WITHDRAW_CASH, interval: MONTHLY, dayOfMonth: 3, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 120000 }]`, giro))
	mk(fmt.Sprintf(`displayName: "Sparrate", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 53582 }]`, bauspar, giro))
	mk(fmt.Sprintf(`displayName: "Zinszahlung", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 62542 }]`, loan, giro))
	mk(fmt.Sprintf(`displayName: "Zinsen", type: INTEREST, interval: MONTHLY, dayOfMonth: 31, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 62542 }]`, loan))
	mk(fmt.Sprintf(`displayName: "Versicherung", type: WITHDRAW_CASH, interval: YEARLY, dayOfMonth: 15, start: "2026-03-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 45000 }]`, giro))

	data := exec(t, schema, ctx, `{ cashFlow(from: "2026-02-10T00:00:00Z", months: 3) {
		currency from
		months { month expectedIn { amount } expectedOut { amount } expectedSaved { amount } expectedFree { amount } actualIn { amount } flows { schedule { displayName } } }
		schedules { displayName }
	} }`, nil)
	cf := data["cashFlow"].(map[string]any)
	if cf["from"].(string)[:10] != "2026-02-01" || cf["currency"] != "EUR" {
		t.Errorf("window = %v %v, want from the first of February in EUR", cf["from"], cf["currency"])
	}
	months := cf["months"].([]any)
	if len(months) != 3 {
		t.Fatalf("%d months, want 3", len(months))
	}
	for i, raw := range months {
		m := raw.(map[string]any)
		in, out, saved, free := m["expectedIn"].(map[string]any)["amount"], m["expectedOut"].(map[string]any)["amount"], m["expectedSaved"].(map[string]any)["amount"], m["expectedFree"].(map[string]any)["amount"]
		wantOut := 120000.0 + 62542.0
		if i == 1 { // March: the yearly insurance
			wantOut += 45000
		}
		if in != 485000.0 || out != wantOut || saved != 53582.0 || free != 485000.0-wantOut-53582.0 {
			t.Errorf("month %d: in %v out %v saved %v free %v; want 485000 %v 53582 %v", i, in, out, saved, free, wantOut, 485000.0-wantOut-53582.0)
		}
		if m["actualIn"] != nil {
			t.Errorf("month %d has actuals, want none: the only cash account is an estimate", i)
		}
		names := map[string]bool{}
		for _, f := range m["flows"].([]any) {
			names[f.(map[string]any)["schedule"].(map[string]any)["displayName"].(string)] = true
		}
		if names["Zinsen"] || !names["Zinszahlung"] {
			t.Errorf("month %d flows = %v: the loan's charge must not appear, its payment must", i, names)
		}
	}
	if got := len(cf["schedules"].([]any)); got != 6 {
		t.Errorf("%d schedules listed, want all 6", got)
	}
}
