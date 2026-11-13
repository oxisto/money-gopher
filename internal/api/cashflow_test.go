package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

// TestCashFlow: the household's month from the schedules, seen from the
// Giro. Salary and rent come in, the card bill and the loan payment go
// out, the Sparrate is saved; the loan's own interest charge is not a
// cash-account movement and does not count twice.
func TestCashFlow(t *testing.T) {
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	user, person, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatalf("EnsureDevUser: %v", err)
	}
	ctx := auth.WithPerson(auth.WithUser(context.Background(), user), person)
	schema := NewSchema(db, &quotes.Updater{DB: db, Registry: quotes.NewRegistry()})
	h := transferHelpers{t, schema, ctx}
	declared := 150000
	giro := h.account("Giro", "CASH", &declared)
	exec(t, schema, ctx, `mutation ($id: ID!) { updateCashAccount(id: $id, input: { iban: "DE02120300000000202051" }) { id } }`, map[string]any{"id": giro})
	bauspar := h.account("Bausparvertrag", "SAVINGS", nil)
	loan := h.account("Vorfinanzierung", "LOAN", nil)

	mk := func(input string) {
		createSchedule(t, h, input)
	}
	mk(fmt.Sprintf(`displayName: "Gehalt", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 28, start: "2026-01-01T00:00:00Z", end: "2026-04-30T00:00:00Z", cashAccountID: %q, books: true, steps: [{ from: "2026-01-01T00:00:00Z", amount: 485000 }]`, giro))
	mk(fmt.Sprintf(`displayName: "Kreditkarte", type: WITHDRAW_CASH, interval: MONTHLY, dayOfMonth: 3, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 120000 }]`, giro))
	mk(fmt.Sprintf(`displayName: "Sparrate", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", end: "2026-04-30T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, books: true, steps: [{ from: "2026-01-01T00:00:00Z", amount: 53582 }]`, bauspar, giro))
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

	// February's statement: the salary (confirming the booking) and the
	// groceries nobody scheduled. February now has actuals; March has none.
	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "feb.txt", "text/plain", []byte("ING-DiBa AG\nKontoauszug 2026\nIBAN DE02 1203 0000 0000 2020 51\n"+
		"Buchung      Buchung / Verwendungszweck                                     Betrag (EUR)\nValuta\n\n"+
		"27.02.2026                 Gehalt/Rente Musterfirma GmbH                                   4.850,00\n27.02.2026                 Lohn/Gehalt 02/2026\n"+
		"02.02.2026                 Lastschrifteinzug LBS SUED                                        -535,82\n02.02.2026                 Bausparvertrag\n"+
		"14.02.2026                 Lastschrift Supermarkt                                            -85,20\n14.02.2026                 Einkauf\n"))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}
	exec(t, schema, ctx, `mutation ($id: ID!, $acct: ID!) { confirmImport(documentID: $id, input: { cashAccountID: $acct, transactions: [] }) { id } }`, map[string]any{"id": doc.ID, "acct": giro})

	data = exec(t, schema, ctx, `{ cashFlow(from: "2026-02-01T00:00:00Z", months: 2) { months {
		actualIn { amount } actualOut { amount } actualSaved { amount } actualFree { amount } unscheduledIn { amount } unscheduledOut { amount } unscheduledSaved { amount } unscheduled { type note }
	} } }`, nil)
	feb := data["cashFlow"].(map[string]any)["months"].([]any)[0].(map[string]any)
	// The salary confirmed its booking, the LBS debit confirmed the
	// Sparrate's Giro leg (saved), the groceries are unscheduled.
	if feb["actualIn"] == nil || feb["actualIn"].(map[string]any)["amount"] != 485000.0 || feb["actualOut"].(map[string]any)["amount"] != 8520.0 || feb["actualSaved"].(map[string]any)["amount"] != 53582.0 {
		t.Errorf("February actuals = %v, want in 485000, out 8520, saved 53582", feb)
	}
	if feb["actualFree"].(map[string]any)["amount"] != 485000.0-8520.0-53582.0 || feb["unscheduledSaved"].(map[string]any)["amount"] != 0.0 {
		t.Errorf("February free/unscheduled saved = %v", feb)
	}
	if feb["unscheduledIn"].(map[string]any)["amount"] != 0.0 || feb["unscheduledOut"].(map[string]any)["amount"] != 8520.0 || len(feb["unscheduled"].([]any)) != 1 {
		t.Errorf("February unscheduled = %v, want only the groceries", feb)
	}
	march := data["cashFlow"].(map[string]any)["months"].([]any)[1].(map[string]any)
	if march["actualIn"] != nil || len(march["unscheduled"].([]any)) != 0 {
		t.Errorf("March has actuals without a statement: %v", march)
	}
}
