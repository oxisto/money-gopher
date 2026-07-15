package api

import (
	"context"
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
		createSecurity(input: {
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
