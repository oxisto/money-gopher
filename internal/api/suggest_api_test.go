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

// TestSuggestedSchedules: three months of statements make a salary a
// suggestion on the account, shaped for the schedule form, and covered
// once a schedule of that name exists.
func TestSuggestedSchedules(t *testing.T) {
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
	giro := h.account("Giro", "CASH", nil)
	exec(t, schema, ctx, `mutation ($id: ID!) { updateCashAccount(id: $id, input: { iban: "DE02120300000000202051" }) { id } }`, map[string]any{"id": giro})

	rows := ""
	for _, m := range []string{"28.04.2026", "28.05.2026", "26.06.2026"} {
		rows += m + "                 Gehalt/Rente Musterfirma GmbH                                   4.850,00\n" + m + "                 Lohn/Gehalt\n"
	}
	rows += "02.05.2026                 Ueberweisung Fahrradladen                                       -1.299,00\n02.05.2026                 Rechnung 77\n"
	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "q2.txt", "text/plain", []byte("ING-DiBa AG\nKontoauszug 2026\nIBAN DE02 1203 0000 0000 2020 51\n"+
		"Buchung      Buchung / Verwendungszweck                                     Betrag (EUR)\nValuta\n\n"+rows))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}
	exec(t, schema, ctx, `mutation ($id: ID!, $acct: ID!) { confirmImport(documentID: $id, input: { cashAccountID: $acct, transactions: [] }) { id } }`, map[string]any{"id": doc.ID, "acct": giro})

	const fields = `name incoming interval dayOfMonth start lastSeen count variable covered steps { from amount { amount currency } }`
	data := exec(t, schema, ctx, fmt.Sprintf(`query ($id: ID!) { cashAccount(id: $id) { suggestedSchedules { %s } } }`, fields), map[string]any{"id": giro})
	got := data["cashAccount"].(map[string]any)["suggestedSchedules"].([]any)
	if len(got) != 1 {
		t.Fatalf("suggestions = %v, want the salary only (the bike shop was once)", got)
	}
	s := got[0].(map[string]any)
	if s["name"] != "Musterfirma GmbH" || s["incoming"] != true || s["interval"] != "MONTHLY" || s["dayOfMonth"] != 28.0 || s["count"] != 3.0 || s["variable"] != false || s["covered"] != false {
		t.Errorf("suggestion = %v", s)
	}
	steps := s["steps"].([]any)
	if len(steps) != 1 || steps[0].(map[string]any)["amount"].(map[string]any)["amount"] != 485000.0 || s["start"].(string)[:10] != "2026-04-28" || s["lastSeen"].(string)[:10] != "2026-06-26" {
		t.Errorf("steps/dates = %v %v %v", steps, s["start"], s["lastSeen"])
	}

	createSchedule(t, h, fmt.Sprintf(`displayName: "musterfirma gmbh", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 28,
		start: "2026-04-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-04-01T00:00:00Z", amount: 485000 }]`, giro))
	data = exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { suggestedSchedules { name covered } } }`, map[string]any{"id": giro})
	if got := data["cashAccount"].(map[string]any)["suggestedSchedules"].([]any); len(got) != 1 || got[0].(map[string]any)["covered"] != true {
		t.Errorf("after creating the schedule: %v, want the suggestion marked covered", got)
	}
}
