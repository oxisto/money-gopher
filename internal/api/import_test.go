package api

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

// ingBuyText is a minimal ING-DiBa buy statement with an IBAN that we can
// match against a cash account. It mirrors the real testdata/ing/buy.txt.
const ingBuyText = `                     ING-DiBa AG · 60628 Frankfurt am Main

                     Wertpapierabrechnung                                           Kauf aus Sparplan
                     Ordernummer                                                    345361721.001
                     ISIN (WKN)                                                     US7170811035 (852009)
                     Wertpapierbezeichnung                                          Pfizer Inc.
                                                                                    Registered Shares DL -,05

                     Nominale                                                       Stück                   13,15412
                     Kurs                                                           EUR                         26,15
                     Handelsplatz                                                   Xetra
                     Ausführungstag / -zeit                                         01.07.2024 um 09:04:09 Uhr
                     Kurswert                                                       EUR                                               343,98
                     Provision                                                      EUR                                                 6,02
                     Endbetrag zu Ihren Lasten                                      EUR                                              350,00

                     Abrechnungs-IBAN                                               DE12 3456 7890 1234 5678 90
                     Valuta                                                         03.07.2024`

// TestImportPipeline exercises the full document import flow:
// upload → parse → IBAN-match → confirm import. It uses a real in-memory
// SQLite database so any schema/migration/SQL-mapping bug surfaces here.
func TestImportPipeline(t *testing.T) {
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
	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry()}
	schema := NewSchema(db, updater)

	// 1. Create a cash account whose IBAN matches the document.
	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: {
			displayName: "ING Giro"
			currency: "EUR"
			iban: "DE12345678901234567890"
		}) { id iban }
	}`, nil)
	accountRaw := data["createCashAccount"].(map[string]any)
	accountID := accountRaw["id"].(string)
	if got := accountRaw["iban"]; got != "DE12345678901234567890" {
		t.Errorf("cash account iban = %v, want DE12345678901234567890", got)
	}

	// 2. Create a portfolio (required for BUY transactions).
	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		createPortfolio(input: { displayName: "My Portfolio", cashAccountID: $account }) { id }
	}`, map[string]any{"account": accountID})
	portfolioID := data["createPortfolio"].(map[string]any)["id"].(string)

	// 2b. Create the security so the importer can match it by ISIN.
	exec(t, schema, ctx, `mutation {
		createAsset(input: {
			displayName: "Pfizer Inc."
			identifiers: [{ kind: "ISIN", value: "US7170811035" }]
			listings: [{ ticker: "PFE", exchange: "XETR", currency: "EUR" }]
		}) { id }
	}`, nil)

	// 3. Upload and process the document using the importer service.
	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "buy.txt", "text/plain", []byte(ingBuyText))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}

	// 4. Query documents and verify IBAN auto-match.
	data = exec(t, schema, ctx, `{
		documents {
			id state
			suggestedCashAccount { id displayName }
			stagedTransactions { id type units currency }
		}
	}`, nil)
	docs := data["documents"].([]any)
	if len(docs) != 1 {
		t.Fatalf("documents count = %d, want 1", len(docs))
	}
	d := docs[0].(map[string]any)
	if got := d["state"]; got != "PARSED" {
		t.Errorf("document state = %v, want PARSED", got)
	}
	suggested := d["suggestedCashAccount"]
	if suggested == nil {
		t.Error("suggestedCashAccount is nil, want the ING Giro account")
	} else if got := suggested.(map[string]any)["id"]; got != accountID {
		t.Errorf("suggestedCashAccount.id = %v, want %v", got, accountID)
	}

	txs := d["stagedTransactions"].([]any)
	if len(txs) != 1 {
		t.Fatalf("stagedTransactions count = %d, want 1", len(txs))
	}
	docID := d["id"].(string)

	// 5. Confirm the import.
	data = exec(t, schema, ctx, `mutation ($docID: ID!, $portfolio: ID!, $account: ID!) {
		confirmImport(documentID: $docID, input: {
			portfolioID: $portfolio
			cashAccountID: $account
			transactions: []
		}) { id type }
	}`, map[string]any{
		"docID":     docID,
		"portfolio": portfolioID,
		"account":   accountID,
	})
	created := data["confirmImport"].([]any)
	if len(created) != 1 {
		t.Fatalf("confirmed transactions count = %d, want 1", len(created))
	}
	if got := created[0].(map[string]any)["type"]; got != "BUY" {
		t.Errorf("confirmed transaction type = %v, want BUY", got)
	}

	// 6. Document should now be IMPORTED.
	data = exec(t, schema, ctx, `{ documents { state } }`, nil)
	docs = data["documents"].([]any)
	if got := docs[0].(map[string]any)["state"]; got != "IMPORTED" {
		t.Errorf("document state after confirm = %v, want IMPORTED", got)
	}

	// 7. The security should have been matched or staged; the cash account
	// balance should reflect the BUY cash outflow.
	data = exec(t, schema, ctx, `query ($id: ID!) {
		cashAccount(id: $id) { balance { amount currency } transactions { id type } }
	}`, map[string]any{"id": accountID})
	account := data["cashAccount"].(map[string]any)
	bal := account["balance"].(map[string]any)
	if amt := bal["amount"].(float64); amt >= 0 {
		t.Errorf("balance after BUY = %v, want negative (cash left the account)", amt)
	}
}

// TestImportFXBond exercises a foreign-currency bond end to end through
// GraphQL: Price resolves in the bond's own currency (not settlement),
// fxRate is exposed, and calculatedCashDelta matches the parsed cashDelta
// exactly despite the FX conversion.
func TestImportFXBond(t *testing.T) {
	text, err := os.ReadFile("../importer/testdata/ing/buy_bond_fx.txt")
	if err != nil {
		t.Fatal(err)
	}

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
	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry()}
	schema := NewSchema(db, updater)

	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "bond_fx.txt", "text/plain", text)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}

	data := exec(t, schema, ctx, `{
		documents {
			id
			stagedTransactions {
				price { amount currency }
				fxRate
				cashDelta { amount }
				calculatedCashDelta { amount }
			}
		}
	}`, nil)
	docs := data["documents"].([]any)
	if len(docs) != 1 {
		t.Fatalf("documents count = %d, want 1", len(docs))
	}
	txs := docs[0].(map[string]any)["stagedTransactions"].([]any)
	if len(txs) != 1 {
		t.Fatalf("stagedTransactions count = %d, want 1", len(txs))
	}
	tx := txs[0].(map[string]any)

	priceObj := tx["price"].(map[string]any)
	if got := priceObj["currency"]; got != "GBP" {
		t.Errorf("price.currency = %v, want GBP (the bond's own currency, not EUR settlement)", got)
	}
	if got := tx["fxRate"]; got != "0.80000000" {
		t.Errorf("fxRate = %v, want \"0.80000000\"", got)
	}

	cashDelta := tx["cashDelta"].(map[string]any)["amount"].(float64)
	calculated := tx["calculatedCashDelta"].(map[string]any)["amount"].(float64)
	if cashDelta != calculated {
		t.Errorf("cashDelta = %v, calculatedCashDelta = %v; want them to match exactly", cashDelta, calculated)
	}
}

// TestImportNoIBANMatch verifies that suggestedCashAccount is nil when
// the document IBAN has no corresponding cash account.
func TestImportNoIBANMatch(t *testing.T) {
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
	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry()}
	schema := NewSchema(db, updater)

	// Cash account with a different IBAN than the document.
	exec(t, schema, ctx, `mutation {
		createCashAccount(input: {
			displayName: "Other bank"
			currency: "EUR"
			iban: "DE00000000000000000000"
		}) { id }
	}`, nil)

	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "buy.txt", "text/plain", []byte(ingBuyText))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}

	data := exec(t, schema, ctx, `{ documents { suggestedCashAccount { id } } }`, nil)
	docs := data["documents"].([]any)
	if docs[0].(map[string]any)["suggestedCashAccount"] != nil {
		t.Error("suggestedCashAccount should be nil when IBAN does not match any account")
	}
}

// TestImportLBSBundle: one LBS year-end PDF carries the Bausparvertrag and
// the Vorfinanzierungskredit statements. Rows arrive tagged with their
// section, the confirm sends them to two accounts, and the loan's interest
// lands as a charge on a LOAN account.
func TestImportLBSBundle(t *testing.T) {
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

	accountOf := func(name, kind string) string {
		data := exec(t, schema, ctx, fmt.Sprintf(`mutation { createCashAccount(input: { displayName: %q, currency: "EUR", kind: %s, reference: "1 234 567 890" }) { id reference } }`, name, kind), nil)
		acct := data["createCashAccount"].(map[string]any)
		if acct["reference"] != "1234567890" {
			t.Errorf("reference stored as %v, want 1234567890 (spaces dropped)", acct["reference"])
		}
		return acct["id"].(string)
	}
	loan := accountOf("LBS Vorfinanzierung", "LOAN")
	savings := accountOf("LBS Bausparvertrag", "SAVINGS")
	giro := exec(t, schema, ctx, `mutation { createCashAccount(input: { displayName: "Giro", currency: "EUR", declaredBalance: 150000 }) { id } }`, nil)["createCashAccount"].(map[string]any)["id"].(string)

	text, err := os.ReadFile("../importer/testdata/lbs/jahresschluss.txt")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "LBS_Jahresschlussdokumente.txt", "text/plain", text)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}

	data := exec(t, schema, ctx, `{ documents { detectedBank suggestedCashAccount { id } stagedTransactions { id accountHint type cashDelta { amount } } } }`, nil)
	d := data["documents"].([]any)[0].(map[string]any)
	if d["detectedBank"] != "LBS" || d["suggestedCashAccount"] != nil {
		t.Errorf("document = %v, want bank LBS and no suggested account (the printed IBAN is the LBS's own)", d)
	}
	var overrides []any
	for _, raw := range d["stagedTransactions"].([]any) {
		stx := raw.(map[string]any)
		hint, _ := stx["accountHint"].(string)
		target := savings
		if strings.HasPrefix(hint, "Vorfinanzierungskredit") {
			target = loan
		}
		overrides = append(overrides, map[string]any{"id": stx["id"], "cashAccountID": target})
	}
	if len(overrides) != 11 {
		t.Fatalf("staged %d rows, want 11", len(overrides))
	}
	// One row is left out: the Bausparvertrag's 5.00 fee.
	for _, raw := range d["stagedTransactions"].([]any) {
		stx := raw.(map[string]any)
		if stx["accountHint"] == "Bausparvertrag 1234567890" && stx["type"] == "ACCOUNT_FEES" && stx["cashDelta"].(map[string]any)["amount"] == -500.0 {
			for _, o := range overrides {
				if o.(map[string]any)["id"] == stx["id"] {
					o.(map[string]any)["skip"] = true
				}
			}
		}
	}

	// A deposit cannot come from the account it lands on.
	execExpectError(t, schema, ctx, `mutation ($id: ID!, $acct: ID!, $tx: [StagedTransactionConfirmInput!]!) {
		confirmImport(documentID: $id, input: { cashAccountID: $acct, depositsFromCashAccountID: $acct, transactions: $tx }) { id }
	}`, map[string]any{"id": doc.ID, "acct": savings, "tx": overrides})

	data = exec(t, schema, ctx, `mutation ($id: ID!, $acct: ID!, $giro: ID!, $tx: [StagedTransactionConfirmInput!]!) {
		confirmImport(documentID: $id, input: { cashAccountID: $acct, depositsFromCashAccountID: $giro, transactions: $tx }) {
			type transferID counterpart { type cashDelta { amount } cashAccount { id } }
		}
	}`, map[string]any{"id": doc.ID, "acct": savings, "giro": giro, "tx": overrides})
	deposits := 0
	for _, raw := range data["confirmImport"].([]any) {
		tx := raw.(map[string]any)
		if tx["type"] == "DEPOSIT_CASH" {
			deposits++
			cp, _ := tx["counterpart"].(map[string]any)
			if tx["transferID"] == nil || cp == nil || cp["type"] != "WITHDRAW_CASH" || cp["cashAccount"].(map[string]any)["id"] != giro {
				t.Errorf("deposit %v: want a linked withdrawal on the Giro", tx)
			}
		} else if tx["transferID"] != nil {
			t.Errorf("%v: only deposits become transfers", tx)
		}
	}
	if deposits != 4 {
		t.Errorf("confirmed %d deposits, want 4", deposits)
	}

	balance := func(id string) float64 {
		data := exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { balance { amount } } }`, map[string]any{"id": id})
		return data["cashAccount"].(map[string]any)["balance"].(map[string]any)["amount"].(float64)
	}
	// 3 × 200 in, the 30 fee (the 5.00 one was skipped), 0.07 interest less 0.02 tax.
	if got := balance(savings); got != 57005 {
		t.Errorf("Bausparvertrag balance = %v, want 57005", got)
	}
	if got := len(exec(t, schema, ctx, `{ documents { stagedTransactions { id } } }`, nil)["documents"].([]any)[0].(map[string]any)["stagedTransactions"].([]any)); got != 10 {
		t.Errorf("%d staged rows remain, want 10: the skipped one is dropped", got)
	}
	// 125 + 112.50 + 32.96 interest charged, 125 paid, 50,000 disbursed.
	if got := balance(loan); got != -5014546 {
		t.Errorf("loan balance = %v, want -5014546", got)
	}
	// The Giro is an estimate: it keeps its declared balance while its
	// ledger shows the 3 × 200 + 125 that left it.
	data = exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { balance { amount } ledgerBalance { amount } } }`, map[string]any{"id": giro})
	g := data["cashAccount"].(map[string]any)
	if g["balance"].(map[string]any)["amount"] != 150000.0 || g["ledgerBalance"].(map[string]any)["amount"] != -72500.0 {
		t.Errorf("Giro = %v, want balance 150000 and ledger -72500", g)
	}
}
