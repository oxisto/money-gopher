package api

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

const scheduleFields = `id displayName type interval dayOfMonth start end books booked
	cashAccount { id } counterpartAccount { id } steps { from amount { amount } }
	currentAmount { amount } nextDate unconfirmed { id }`

func createSchedule(t *testing.T, h transferHelpers, input string) map[string]any {
	t.Helper()
	data := exec(t, h.schema, h.ctx, fmt.Sprintf(`mutation { createSchedule(input: { %s }) { %s } }`, input, scheduleFields), nil)
	return data["createSchedule"].(map[string]any)
}

func TestSchedule_Lifecycle(t *testing.T) {
	schema, ctx := newTestSchema(t)
	h := transferHelpers{t, schema, ctx}
	giro := h.account("Giro", "CASH", nil)
	bauspar := h.account("Bausparvertrag", "SAVINGS", nil)
	loan := h.account("Vorfinanzierung", "LOAN", nil)

	sparrate := createSchedule(t, h, fmt.Sprintf(`displayName: "Sparrate", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1,
		start: "2026-01-01T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q,
		steps: [{ from: "2026-01-01T00:00:00Z", amount: 53582 }]`, bauspar, giro))
	if sparrate["currentAmount"].(map[string]any)["amount"] != 53582.0 || sparrate["books"] != false || sparrate["end"] != nil {
		t.Errorf("created = %v, want +535.82 on the Bausparvertrag, not booking, no end", sparrate)
	}
	if sparrate["nextDate"] == nil {
		t.Error("nextDate is nil for a schedule without end")
	}

	// Interest on a loan is charged: negative as it hits the loan.
	interest := createSchedule(t, h, fmt.Sprintf(`displayName: "Zinsen", type: INTEREST, interval: MONTHLY, dayOfMonth: 31,
		start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 62542 }]`, loan))
	if interest["currentAmount"].(map[string]any)["amount"] != -62542.0 {
		t.Errorf("loan interest = %v, want -62542", interest["currentAmount"])
	}

	// The Giro sees the Sparrate too: it is paid from there.
	data := exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { schedules { displayName } } }`, map[string]any{"id": giro})
	if got := data["cashAccount"].(map[string]any)["schedules"].([]any); len(got) != 1 || got[0].(map[string]any)["displayName"] != "Sparrate" {
		t.Errorf("Giro schedules = %v, want the Sparrate", got)
	}

	// Steps replace; a step from a later date raises the amount; end and clearEnd.
	id := sparrate["id"].(string)
	data = exec(t, schema, ctx, fmt.Sprintf(`mutation ($id: ID!) { updateSchedule(id: $id, input: {
		steps: [{ from: "2026-01-01T00:00:00Z", amount: 53582 }, { from: "2027-01-01T00:00:00Z", amount: 60000 }],
		end: "2030-12-31T00:00:00Z" }) { %s } }`, scheduleFields), map[string]any{"id": id})
	upd := data["updateSchedule"].(map[string]any)
	if len(upd["steps"].([]any)) != 2 || upd["end"] == nil || upd["currentAmount"].(map[string]any)["amount"] != 53582.0 {
		t.Errorf("after update = %v, want two steps, an end, and 535.82 still in force", upd)
	}
	data = exec(t, schema, ctx, fmt.Sprintf(`mutation ($id: ID!) { updateSchedule(id: $id, input: { clearEnd: true }) { %s } }`, scheduleFields), map[string]any{"id": id})
	if data["updateSchedule"].(map[string]any)["end"] != nil {
		t.Error("clearEnd left an end")
	}

	// Expected flows: one Sparrate per month (+ on the Bausparvertrag) and
	// the interest charge (− on the loan), by date.
	data = exec(t, schema, ctx, `{ expectedFlows(from: "2026-03-01T00:00:00Z", to: "2026-05-01T00:00:00Z") {
		date amount { amount } booked schedule { displayName } } }`, nil)
	flows := data["expectedFlows"].([]any)
	var got []string
	for _, f := range flows {
		m := f.(map[string]any)
		got = append(got, fmt.Sprintf("%s %s %v", m["date"].(string)[:10], m["schedule"].(map[string]any)["displayName"], m["amount"].(map[string]any)["amount"]))
	}
	want := []string{"2026-03-01 Sparrate 53582", "2026-03-31 Zinsen -62542", "2026-04-01 Sparrate 53582", "2026-04-30 Zinsen -62542"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("flows = %v, want %v", got, want)
	}

	// Validation.
	for name, input := range map[string]string{
		"transfer without source":  fmt.Sprintf(`displayName: "x", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 1 }]`, bauspar),
		"transfer onto itself":     fmt.Sprintf(`displayName: "x", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 1 }]`, giro, giro),
		"day 32":                   fmt.Sprintf(`displayName: "x", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 32, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 1 }]`, giro),
		"no steps":                 fmt.Sprintf(`displayName: "x", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: []`, giro),
		"zero amount":              fmt.Sprintf(`displayName: "x", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 0 }]`, giro),
		"end before start":         fmt.Sprintf(`displayName: "x", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", end: "2025-01-01T00:00:00Z", cashAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 1 }]`, giro),
		"counterpart on a deposit": fmt.Sprintf(`displayName: "x", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, steps: [{ from: "2026-01-01T00:00:00Z", amount: 1 }]`, giro, bauspar),
		"unknown account":          `displayName: "x", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 1, start: "2026-01-01T00:00:00Z", cashAccountID: "nope", steps: [{ from: "2026-01-01T00:00:00Z", amount: 1 }]`,
	} {
		t.Run(name, func(t *testing.T) {
			execExpectError(t, schema, ctx, fmt.Sprintf(`mutation { createSchedule(input: { %s }) { id } }`, input), nil)
		})
	}

	// Another person sees nothing and cannot touch it.
	other := auth.WithPerson(auth.WithUser(context.Background(), &persistence.User{ID: "other"}), &persistence.Person{ID: "other"})
	if got := len(exec(t, schema, other, `{ schedules { id } }`, nil)["schedules"].([]any)); got != 0 {
		t.Errorf("other person sees %d schedules", got)
	}
	execExpectError(t, schema, other, `mutation ($id: ID!) { deleteSchedule(id: $id) }`, map[string]any{"id": id})

	exec(t, schema, ctx, `mutation ($id: ID!) { deleteSchedule(id: $id) }`, map[string]any{"id": id})
	if data := exec(t, schema, ctx, `query ($id: ID!) { schedule(id: $id) { id } }`, map[string]any{"id": id}); data["schedule"] != nil {
		t.Error("the schedule survived its deletion")
	}
}

// TestSchedule_Books: a booking schedule writes its past occurrences as
// real transactions, once each; a statement later confirms them instead
// of duplicating them.
func TestSchedule_Books(t *testing.T) {
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
	bauspar := exec(t, schema, ctx, `mutation { createCashAccount(input: { displayName: "LBS Bausparvertrag", currency: "EUR", kind: SAVINGS, reference: "1234567890" }) { id } }`, nil)["createCashAccount"].(map[string]any)["id"].(string)
	loan := exec(t, schema, ctx, `mutation { createCashAccount(input: { displayName: "LBS Vorfinanzierung", currency: "EUR", kind: LOAN, reference: "1234567890" }) { id } }`, nil)["createCashAccount"].(map[string]any)["id"].(string)

	// Creating a booking schedule books right away, up to today: the test
	// pins "today" by starting in the past and reading what was booked
	// through the schedule itself, then books explicitly up to a fixed day.
	sparrate := createSchedule(t, h, fmt.Sprintf(`displayName: "Sparrate", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1,
		start: "2025-02-01T00:00:00Z", end: "2025-04-30T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, books: true,
		steps: [{ from: "2025-01-01T00:00:00Z", amount: 20000 }]`, bauspar, giro))
	if sparrate["booked"] != 3.0 {
		t.Fatalf("booked = %v, want 3 (February to April 2025)", sparrate["booked"])
	}
	if h.balance(bauspar) != 60000 {
		t.Errorf("Bausparvertrag = %v, want 60000", h.balance(bauspar))
	}
	data := exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { balance { amount } ledgerBalance { amount }
		transactions { type scheduleID scheduledFor counterpart { id } document { id } } } }`, map[string]any{"id": giro})
	g := data["cashAccount"].(map[string]any)
	if g["balance"].(map[string]any)["amount"] != 150000.0 || g["ledgerBalance"].(map[string]any)["amount"] != -60000.0 {
		t.Errorf("Giro = %v, want the declared 150000 and a ledger of -60000", g)
	}
	for _, raw := range g["transactions"].([]any) {
		tx := raw.(map[string]any)
		if tx["type"] != "WITHDRAW_CASH" || tx["scheduleID"] != sparrate["id"] || tx["scheduledFor"] == nil || tx["counterpart"] == nil || tx["document"] != nil {
			t.Errorf("Giro leg = %v, want a scheduled, linked, unconfirmed withdrawal", tx)
		}
	}

	// Booking again writes nothing; the loan's interest books as a charge.
	zinsen := createSchedule(t, h, fmt.Sprintf(`displayName: "Zinsen", type: INTEREST, interval: MONTHLY, dayOfMonth: 31,
		start: "2025-01-01T00:00:00Z", end: "2025-03-31T00:00:00Z", cashAccountID: %q, books: true,
		steps: [{ from: "2025-01-01T00:00:00Z", amount: 12500 }]`, loan))
	if zinsen["booked"] != 3.0 || h.balance(loan) != -37500 {
		t.Errorf("Zinsen booked %v, loan %v; want 3 and -37500", zinsen["booked"], h.balance(loan))
	}
	if n := exec(t, schema, ctx, `mutation { bookSchedules }`, nil)["bookSchedules"]; n != 0.0 {
		t.Errorf("a second booking run wrote %v, want 0", n)
	}
	if n := (&Bookkeeper{DB: db}).BookAll; n == nil {
		t.Fatal("no bookkeeper")
	} else if written, err := n(ctx); err != nil || written != 0 {
		t.Errorf("Bookkeeper.BookAll = %d, %v; want 0, nil", written, err)
	}

	// A schedule that does not book yet writes nothing until switched on.
	jahresentgelt := createSchedule(t, h, fmt.Sprintf(`displayName: "Jahresentgelt", type: ACCOUNT_FEES, interval: YEARLY, dayOfMonth: 2,
		start: "2025-01-02T00:00:00Z", end: "2025-12-31T00:00:00Z", cashAccountID: %q, steps: [{ from: "2025-01-01T00:00:00Z", amount: 3000 }]`, bauspar))
	if jahresentgelt["booked"] != 0.0 {
		t.Errorf("a non-booking schedule booked %v", jahresentgelt["booked"])
	}
	data = exec(t, schema, ctx, `mutation ($id: ID!) { updateSchedule(id: $id, input: { books: true }) { booked } }`, map[string]any{"id": jahresentgelt["id"]})
	if data["updateSchedule"].(map[string]any)["booked"] != 1.0 || h.balance(bauspar) != 57000 {
		t.Errorf("after switching booking on: %v, Bausparvertrag %v; want 1 and 57000", data["updateSchedule"], h.balance(bauspar))
	}

	// The LBS bundle lists three Sparraten of 200 in February, March and
	// April (on the 3rd, 3rd and 1st) and the loan's interest; those rows
	// confirm the scheduled bookings: no second transaction, the bank's
	// date wins, the document is attached. The fee rows and the rest
	// import as new transactions.
	text, err := os.ReadFile("../importer/testdata/lbs/jahresschluss.txt")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "LBS.txt", "text/plain", text)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}
	// Shift the fixture's 2025 rows? They are 2025 already: the Sparraten
	// match February–April, the loan's interest rows (Oct–Dec) do not.
	data = exec(t, schema, ctx, `{ documents { stagedTransactions { id accountHint type } } }`, nil)
	var overrides []any
	for _, raw := range data["documents"].([]any)[0].(map[string]any)["stagedTransactions"].([]any) {
		stx := raw.(map[string]any)
		target := bauspar
		if hint, _ := stx["accountHint"].(string); len(hint) > 0 && hint[0] == 'V' {
			target = loan
		}
		overrides = append(overrides, map[string]any{"id": stx["id"], "cashAccountID": target})
	}
	exec(t, schema, ctx, `mutation ($id: ID!, $acct: ID!, $giro: ID!, $tx: [StagedTransactionConfirmInput!]!) {
		confirmImport(documentID: $id, input: { cashAccountID: $acct, depositsFromCashAccountID: $giro, transactions: $tx }) { id }
	}`, map[string]any{"id": doc.ID, "acct": bauspar, "giro": giro, "tx": overrides})

	data = exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { balance { amount }
		transactions { type time cashDelta { amount } scheduleID document { id } counterpart { time } } } }`, map[string]any{"id": bauspar})
	b := data["cashAccount"].(map[string]any)
	// 3 × 200 Sparrate (confirmed, not doubled) − 30 Jahresentgelt (scheduled)
	// − 30 − 5 (the statement's fees: the 30 on 2025-07-01 is more than a
	// week from the scheduled 2025-01-02, so it is a new row) + 0.05.
	if got := b["balance"].(map[string]any)["amount"]; got != 53505.0 {
		t.Errorf("Bausparvertrag after the statement = %v, want 53505", got)
	}
	confirmed := 0
	for _, raw := range b["transactions"].([]any) {
		tx := raw.(map[string]any)
		if tx["type"] == "DEPOSIT_CASH" {
			if tx["scheduleID"] == nil || tx["document"] == nil {
				t.Errorf("Sparrate %v: want scheduled and confirmed by the document", tx)
			}
			confirmed++
			// The bank's date moved onto both legs.
			if cp := tx["counterpart"].(map[string]any); cp["time"] != tx["time"] {
				t.Errorf("legs disagree on the date: %v vs %v", tx["time"], cp["time"])
			}
		}
	}
	if confirmed != 3 {
		t.Errorf("%d deposits on the Bausparvertrag, want exactly the 3 scheduled ones", confirmed)
	}
	// The schedule reports nothing unconfirmed any more.
	data = exec(t, schema, ctx, `query ($id: ID!) { schedule(id: $id) { unconfirmed { id } } }`, map[string]any{"id": sparrate["id"]})
	if got := data["schedule"].(map[string]any)["unconfirmed"].([]any); len(got) != 0 {
		t.Errorf("unconfirmed = %v, want none", got)
	}

	// Deleting the schedule leaves its bookings as plain transactions; a
	// new schedule for the same Sparrate adopts them instead of booking
	// them again, counterpart legs included.
	exec(t, schema, ctx, `mutation ($id: ID!) { deleteSchedule(id: $id) }`, map[string]any{"id": sparrate["id"]})
	again := createSchedule(t, h, fmt.Sprintf(`displayName: "Sparrate", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1,
		start: "2025-02-01T00:00:00Z", end: "2025-04-30T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, books: true,
		steps: [{ from: "2025-01-01T00:00:00Z", amount: 20000 }]`, bauspar, giro))
	if again["booked"] != 3.0 || h.balance(bauspar) != 53505 {
		t.Errorf("recreated: booked %v, Bausparvertrag %v; want the 3 adopted and 53505 unchanged", again["booked"], h.balance(bauspar))
	}
	data = exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { transactions { type cashDelta { amount } scheduleID } } }`, map[string]any{"id": giro})
	for _, raw := range data["cashAccount"].(map[string]any)["transactions"].([]any) {
		// The Sparrate legs; the loan repayment's leg from the statement is not one.
		if tx := raw.(map[string]any); tx["cashDelta"].(map[string]any)["amount"] == -20000.0 && tx["scheduleID"] != again["id"] {
			t.Errorf("Giro leg %v not adopted by the new schedule", tx)
		}
	}
}

// TestSchedule_PooledDebit: the LBS collects the Sparrate and the Zinsrate
// in one direct debit, so the Giro statement shows 1,161.24 where the
// schedules booked two legs of 535.82 and 625.42. The row confirms both.
func TestSchedule_PooledDebit(t *testing.T) {
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
	bauspar := h.account("Bausparvertrag", "SAVINGS", nil)
	loan := h.account("Vorfinanzierung", "LOAN", nil)
	createSchedule(t, h, fmt.Sprintf(`displayName: "Sparrate", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1,
		start: "2026-09-01T00:00:00Z", end: "2026-09-30T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, books: true,
		steps: [{ from: "2026-09-01T00:00:00Z", amount: 53582 }]`, bauspar, giro))
	createSchedule(t, h, fmt.Sprintf(`displayName: "Zinszahlung", type: TRANSFER, interval: MONTHLY, dayOfMonth: 1,
		start: "2026-09-01T00:00:00Z", end: "2026-09-30T00:00:00Z", cashAccountID: %q, counterpartAccountID: %q, books: true,
		steps: [{ from: "2026-09-01T00:00:00Z", amount: 62542 }]`, loan, giro))
	if h.balance(giro) != -116124 {
		t.Fatalf("Giro after booking = %v, want -116124", h.balance(giro))
	}

	// A Giro statement in the ING layout, naming the account by its IBAN.
	exec(t, schema, ctx, `mutation ($id: ID!) { updateCashAccount(id: $id, input: { iban: "DE02120300000000202051" }) { id } }`, map[string]any{"id": giro})
	statement := func(rows string) string {
		return "ING-DiBa AG\nKontoauszug Girokonto\nIBAN DE02 1203 0000 0000 2020 51\n" +
			"Buchung    Buchung / Verwendungszweck                      Betrag (EUR)\n" + rows
	}
	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "giro.txt", "text/plain", []byte(statement("01.09.2026 Lastschrifteinzug LBS SUED Bausparen     -1.161,24\n")))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}
	// The review says so beforehand: the row confirms both legs and
	// doubles nothing.
	stagedOf := func(id string) map[string]any {
		data := exec(t, schema, ctx, `{ documents { id stagedTransactions { confirms { id schedule { displayName } } duplicateOf { id schedule { displayName } } } } }`, nil)
		for _, raw := range data["documents"].([]any) {
			if d := raw.(map[string]any); d["id"] == id {
				return d["stagedTransactions"].([]any)[0].(map[string]any)
			}
		}
		t.Fatalf("document %s not listed", id)
		return nil
	}
	row := stagedOf(doc.ID)
	if got := row["confirms"].([]any); len(got) != 2 || row["duplicateOf"] != nil {
		t.Errorf("review hint = %v, want two bookings confirmed and no duplicate", row)
	}

	data := exec(t, schema, ctx, `mutation ($id: ID!, $acct: ID!) {
		confirmImport(documentID: $id, input: { cashAccountID: $acct, transactions: [] }) { cashDelta { amount } schedule { displayName } document { id } }
	}`, map[string]any{"id": doc.ID, "acct": giro})
	confirmed := data["confirmImport"].([]any)
	if len(confirmed) != 2 {
		t.Fatalf("confirmed %d transactions, want the 2 pooled legs: %v", len(confirmed), confirmed)
	}
	for _, raw := range confirmed {
		tx := raw.(map[string]any)
		if tx["schedule"] == nil || tx["document"] == nil {
			t.Errorf("%v: want a scheduled leg confirmed by the statement", tx)
		}
	}
	if h.balance(giro) != -116124 {
		t.Errorf("Giro after the statement = %v, want -116124 still (nothing doubled)", h.balance(giro))
	}

	// A second statement listing one of the legs again confirms it once
	// more: the leg then carries both documents, and nothing is doubled.
	doc2, err := svc.Upload(ctx, person.ID, "giro2.txt", "text/plain", []byte(statement("02.09.2026 Lastschrifteinzug LBS SUED Sparrate     -535,82\n")))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc2.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}
	row = stagedOf(doc2.ID)
	if c := row["confirms"].([]any); len(c) != 1 || c[0].(map[string]any)["schedule"].(map[string]any)["displayName"] != "Sparrate" || row["duplicateOf"] != nil {
		t.Errorf("second statement's hint = %v, want the Sparrate leg confirmed again", row)
	}
	data = exec(t, schema, ctx, `mutation ($id: ID!, $acct: ID!) {
		confirmImport(documentID: $id, input: { cashAccountID: $acct, transactions: [] }) { schedule { displayName } documents { id } }
	}`, map[string]any{"id": doc2.ID, "acct": giro})
	if got := data["confirmImport"].([]any); len(got) != 1 || len(got[0].(map[string]any)["documents"].([]any)) != 2 {
		t.Errorf("after the second statement: %v, want the one leg with two documents", got)
	}
	if h.balance(giro) != -116124 {
		t.Errorf("Giro after the second statement = %v, want -116124 still", h.balance(giro))
	}
}

// TestSchedule_AdoptsByCounterparty: a salary month with a bonus has the
// same payer and date but a different amount. The schedule adopts it all
// the same, and an expectation booked before the statement arrived gives
// way to the real row on the next booking run.
func TestSchedule_AdoptsByCounterparty(t *testing.T) {
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

	statement := func(rows string) []byte {
		return []byte("ING-DiBa AG\nKontoauszug 2026\nIBAN DE02 1203 0000 0000 2020 51\n" +
			"Buchung      Buchung / Verwendungszweck                                     Betrag (EUR)\nValuta\n\n" + rows)
	}
	svc := importer.NewService(db)
	importStatement := func(name, rows string) {
		doc, err := svc.Upload(ctx, person.ID, name, "text/plain", statement(rows))
		if err != nil {
			t.Fatalf("Upload: %v", err)
		}
		if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
			t.Fatalf("Process: %v", err)
		}
		exec(t, schema, ctx, `mutation ($id: ID!, $acct: ID!) { confirmImport(documentID: $id, input: { cashAccountID: $acct, transactions: [] }) { id } }`, map[string]any{"id": doc.ID, "acct": giro})
	}
	// May and June at the usual amount, July with a bonus on top.
	importStatement("q2.txt",
		"28.05.2026                 Gehalt/Rente Musterfirma GmbH                                   4.850,00\n"+
			"28.05.2026                 Lohn/Gehalt 05/2026\n"+
			"26.06.2026                 Gehalt/Rente Musterfirma GmbH                                   4.850,00\n"+
			"26.06.2026                 Lohn/Gehalt 06/2026\n"+
			"28.07.2026                 Gehalt/Rente Musterfirma GmbH                                   7.850,00\n"+
			"28.07.2026                 Lohn/Gehalt 07/2026 inkl. Bonus\n")
	if h.balance(giro) != 1755000 {
		t.Fatalf("Giro after the statement = %v, want 1755000", h.balance(giro))
	}

	// Named "Gehalt", not after the employer: the schedule learns the
	// payer from the rows it adopts by amount.
	gehalt := createSchedule(t, h, fmt.Sprintf(`displayName: "Gehalt", type: DEPOSIT_CASH, interval: MONTHLY, dayOfMonth: 28,
		start: "2026-05-01T00:00:00Z", end: "2026-08-31T00:00:00Z", cashAccountID: %q, books: true,
		steps: [{ from: "2026-05-01T00:00:00Z", amount: 485000 }]`, giro))
	// May and June by amount, July by counterparty; August booked as usual.
	if gehalt["booked"] != 4.0 || h.balance(giro) != 1755000+485000 {
		t.Errorf("booked %v, Giro %v; want 4 and 2240000 (the July bonus month adopted, August booked)", gehalt["booked"], h.balance(giro))
	}

	// August's statement arrives: a bonus again, so the row cannot confirm
	// the booking by amount and imports beside it. The next booking run
	// adopts the real row and drops the expectation.
	importStatement("aug.txt",
		"28.08.2026                 Gehalt/Rente Musterfirma GmbH                                   5.350,00\n"+
			"28.08.2026                 Lohn/Gehalt 08/2026 inkl. Praemie\n")
	if h.balance(giro) != 1755000+485000+535000 {
		t.Fatalf("Giro after August's statement = %v, want the expectation still beside the real row", h.balance(giro))
	}
	exec(t, schema, ctx, `mutation { bookSchedules(until: "2026-08-31T00:00:00Z") }`, nil)
	if h.balance(giro) != 1755000+535000 {
		t.Errorf("Giro after the booking run = %v, want 2290000: the real August row only", h.balance(giro))
	}
	data := exec(t, schema, ctx, `query ($id: ID!) { schedule(id: $id) { booked unconfirmed { id } } }`, map[string]any{"id": gehalt["id"]})
	if s := data["schedule"].(map[string]any); s["booked"] != 4.0 || len(s["unconfirmed"].([]any)) != 0 {
		t.Errorf("schedule after the run = %v, want 4 occurrences, none unconfirmed", s)
	}
}
