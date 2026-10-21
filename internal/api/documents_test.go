package api

import (
	"context"
	"database/sql"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

// newTestEnv is like newTestSchema but also returns the db and active person so
// tests can set up fixture data directly without going through GraphQL.
func newTestEnv(t *testing.T) (*graphql.Schema, context.Context, *persistence.DB, *persistence.Person) {
	t.Helper()

	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	user, person, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatalf("EnsureDevUser() error = %v", err)
	}

	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry()}
	schema := NewSchema(db, updater)
	ctx := auth.WithPerson(auth.WithUser(context.Background(), user), person)

	return schema, ctx, db, person
}

// TestDocumentsLifecycle exercises the full import flow through GraphQL:
// upload (via importer.Service), list, confirm, delete.
func TestDocumentsLifecycle(t *testing.T) {
	schema, ctx, db, person := newTestEnv(t)

	// Seed entities.
	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)

	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		createPortfolio(input: { displayName: "My Portfolio", cashAccountID: $account }) { id }
	}`, map[string]any{"account": accountID})
	portfolioID := data["createPortfolio"].(map[string]any)["id"].(string)

	exec(t, schema, ctx, `mutation {
		createSecurity(input: {
			displayName: "iShares Core MSCI World"
			identifiers: [{ kind: "ISIN", value: "IE00B4L5Y983" }]
		}) { id }
	}`, nil)

	// Upload and process a CSV via the importer service (simulating /upload).
	const sampleCSV = "Date,Type,Security,ISIN,Units,Price,Fees,Taxes,CashDelta,Currency\n" +
		"2026-01-03,BUY,iShares Core MSCI World,IE00B4L5Y983,10,100.50,1.50,0,,EUR\n" +
		"2026-01-04,DEPOSIT_CASH,,,,,0,0,200.00,EUR\n"

	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "test.csv", "text/csv", []byte(sampleCSV))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// List documents via GraphQL; should see the PARSED document.
	data = exec(t, schema, ctx, `query {
		documents {
			id filename state detectedBank
			stagedTransactions {
				id type isin security { id }
				cashDelta { amount }
				calculatedCashDelta { amount }
			}
		}
	}`, nil)
	docs := data["documents"].([]any)
	if len(docs) != 1 {
		t.Fatalf("documents = %d, want 1", len(docs))
	}
	docMap := docs[0].(map[string]any)
	if docMap["state"] != "PARSED" {
		t.Errorf("state = %v, want PARSED", docMap["state"])
	}
	if docMap["detectedBank"] != "csv" {
		t.Errorf("detectedBank = %v, want csv", docMap["detectedBank"])
	}
	docID := docMap["id"].(string)

	staged := docMap["stagedTransactions"].([]any)
	if len(staged) != 2 {
		t.Fatalf("stagedTransactions = %d, want 2", len(staged))
	}

	// BUY should have auto-linked security.
	buy := staged[0].(map[string]any)
	if buy["type"] != "BUY" {
		t.Errorf("staged[0].type = %v, want BUY", buy["type"])
	}
	if buy["security"] == nil {
		t.Error("staged[0].security is nil; expected auto-linked security")
	}
	buyID := buy["id"].(string)

	// The CSV left CashDelta blank for the BUY; calculatedCashDelta fills in
	// what it should be: -(10*100.50 + 1.50) = -100650 minor units.
	if got := buy["calculatedCashDelta"].(map[string]any)["amount"].(float64); got != -100650 {
		t.Errorf("BUY calculatedCashDelta = %v, want -100650", got)
	}

	// DEPOSIT_CASH has no security, and no price to calculate a cash delta
	// from — calculatedCashDelta must be nil, not a bogus 0-minus-fees value.
	depositMap := staged[1].(map[string]any)
	if depositMap["security"] != nil {
		t.Error("staged[1].security should be nil for DEPOSIT_CASH")
	}
	if depositMap["calculatedCashDelta"] != nil {
		t.Errorf("DEPOSIT_CASH calculatedCashDelta = %v, want nil", depositMap["calculatedCashDelta"])
	}

	// Filter by state.
	data = exec(t, schema, ctx, `query { documents(state: PARSED) { id } }`, nil)
	if len(data["documents"].([]any)) != 1 {
		t.Errorf("filtered documents = %d, want 1", len(data["documents"].([]any)))
	}
	data = exec(t, schema, ctx, `query { documents(state: IMPORTED) { id } }`, nil)
	if len(data["documents"].([]any)) != 0 {
		t.Errorf("filtered IMPORTED docs = %d, want 0", len(data["documents"].([]any)))
	}

	// Confirm the import — BUY override is empty (defaults apply), DEPOSIT uses defaults.
	data = exec(t, schema, ctx, `mutation ($docID: ID!, $portfolio: ID!, $account: ID!, $buyID: ID!) {
		confirmImport(documentID: $docID, input: {
			portfolioID: $portfolio
			cashAccountID: $account
			transactions: [{ id: $buyID }]
		}) { id type source cashDelta { amount } document { id filename } }
	}`, map[string]any{
		"docID":     docID,
		"portfolio": portfolioID,
		"account":   accountID,
		"buyID":     buyID,
	})
	committed := data["confirmImport"].([]any)
	if len(committed) != 2 {
		t.Fatalf("confirmImport = %d transactions, want 2", len(committed))
	}
	for _, tx := range committed {
		m := tx.(map[string]any)
		if got := m["source"]; got != "import:csv" {
			t.Errorf("source = %v, want import:csv", got)
		}
		// The committed transaction must trace back to the document it came from.
		doc, ok := m["document"].(map[string]any)
		if !ok {
			t.Fatalf("document = %v, want a linked document", m["document"])
		}
		if doc["id"] != docID {
			t.Errorf("document.id = %v, want %v", doc["id"], docID)
		}
	}
	// BUY: cashDelta should be -(10*10050 + 150) = -100650 minor units.
	buyTx := committed[0].(map[string]any)
	if got := buyTx["cashDelta"].(map[string]any)["amount"].(float64); got != -100650 {
		t.Errorf("BUY cashDelta = %v, want -100650", got)
	}

	// Document state is now IMPORTED.
	data = exec(t, schema, ctx, `query { documents { state } }`, nil)
	if got := data["documents"].([]any)[0].(map[string]any)["state"]; got != "IMPORTED" {
		t.Errorf("state after confirm = %v, want IMPORTED", got)
	}

	// Reprocessing an already-IMPORTED document must not touch its
	// already-committed transactions — only regenerate staged ones.
	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		reprocessDocument(id: $id) {
			state
			stagedTransactions { id }
			transactions { id }
		}
	}`, map[string]any{"id": docID})
	reprocessed := data["reprocessDocument"].(map[string]any)
	if got := reprocessed["state"]; got != "PARSED" {
		t.Errorf("state after reprocess = %v, want PARSED", got)
	}
	if got := len(reprocessed["stagedTransactions"].([]any)); got != 2 {
		t.Errorf("stagedTransactions after reprocess = %d, want 2 (regenerated)", got)
	}
	if got := len(reprocessed["transactions"].([]any)); got != 2 {
		t.Errorf("transactions after reprocess = %d, want 2 (untouched)", got)
	}

	// Delete the document.
	exec(t, schema, ctx, `mutation ($id: ID!) { deleteDocument(id: $id) }`,
		map[string]any{"id": docID})

	data = exec(t, schema, ctx, `query { documents { id } }`, nil)
	if len(data["documents"].([]any)) != 0 {
		t.Errorf("documents after delete = %d, want 0", len(data["documents"].([]any)))
	}
}

// TestConfirmImport_ReconcilesReprocessedDocument verifies that
// re-confirming a reprocessed document (which already has committed
// transactions from an earlier confirm — see the reprocess step at the end
// of TestDocumentsLifecycle) reconciles each staged row against its
// already-committed counterpart in place, instead of creating a duplicate.
func TestConfirmImport_ReconcilesReprocessedDocument(t *testing.T) {
	schema, ctx, db, person := newTestEnv(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)

	const sampleCSV = "Date,Type,Security,ISIN,Units,Price,Fees,Taxes,CashDelta,Currency\n" +
		"2026-01-04,DEPOSIT_CASH,,,,,0,0,200.00,EUR\n"

	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "test.csv", "text/csv", []byte(sampleCSV))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	data = exec(t, schema, ctx, `mutation ($docID: ID!, $account: ID!) {
		confirmImport(documentID: $docID, input: { cashAccountID: $account, transactions: [] })
		{ id }
	}`, map[string]any{"docID": doc.ID, "account": accountID})
	firstTxID := data["confirmImport"].([]any)[0].(map[string]any)["id"].(string)

	// Re-process: regenerates staged transactions, leaves the committed one alone.
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process() (reprocess) error = %v", err)
	}

	// Re-confirm the reprocessed document.
	data = exec(t, schema, ctx, `mutation ($docID: ID!, $account: ID!) {
		confirmImport(documentID: $docID, input: { cashAccountID: $account, transactions: [] })
		{ id cashDelta { amount } }
	}`, map[string]any{"docID": doc.ID, "account": accountID})
	second := data["confirmImport"].([]any)
	if len(second) != 1 {
		t.Fatalf("confirmImport (reconcile) = %d transactions, want 1", len(second))
	}
	secondTxID := second[0].(map[string]any)["id"].(string)
	if secondTxID != firstTxID {
		t.Errorf("reconciled transaction ID = %q, want %q (same row, not a duplicate)", secondTxID, firstTxID)
	}

	// The document must still have exactly one linked transaction — no
	// duplicate was created alongside the reconciled one.
	txs, err := db.ListTransactionsByDocument(ctx, sql.NullString{String: doc.ID, Valid: true})
	if err != nil {
		t.Fatalf("ListTransactionsByDocument() error = %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("transactions linked to document = %d, want 1", len(txs))
	}
}

// TestConfirmImport_RollsBackOnFailure verifies that a confirm failing partway
// through commits nothing: the rows before the failing one must not be left
// behind, or confirming again after fixing the input would duplicate them.
func TestConfirmImport_RollsBackOnFailure(t *testing.T) {
	schema, ctx, db, person := newTestEnv(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)

	// The DEPOSIT_CASH is valid; the BUY is not, because no portfolio is given.
	const sampleCSV = "Date,Type,Security,ISIN,Units,Price,Fees,Taxes,CashDelta,Currency\n" +
		"2026-01-03,DEPOSIT_CASH,,,,,0,0,200.00,EUR\n" +
		"2026-01-04,BUY,iShares Core MSCI World,IE00B4L5Y983,10,100.50,1.50,0,,EUR\n"

	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "test.csv", "text/csv", []byte(sampleCSV))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	execExpectError(t, schema, ctx, `mutation ($docID: ID!, $account: ID!) {
		confirmImport(documentID: $docID, input: { cashAccountID: $account, transactions: [] })
		{ id }
	}`, map[string]any{"docID": doc.ID, "account": accountID})

	txs, err := db.ListTransactionsByDocument(ctx, sql.NullString{String: doc.ID, Valid: true})
	if err != nil {
		t.Fatalf("ListTransactionsByDocument() error = %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("transactions linked to document = %d, want 0 (rolled back)", len(txs))
	}

	got, err := db.GetDocument(ctx, persistence.GetDocumentParams{ID: doc.ID, PersonID: person.ID})
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if got.State != "PARSED" {
		t.Errorf("state = %v, want PARSED", got.State)
	}
}

// TestConfirmImportNotFound verifies that confirmImport returns an error for
// unknown or foreign documents.
func TestConfirmImportNotFound(t *testing.T) {
	schema, ctx := newTestSchema(t)

	execExpectError(t, schema, ctx, `mutation {
		confirmImport(documentID: "no-such-doc", input: { transactions: [] }) { id }
	}`, nil)
}
