package api

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

// newTestEnv is like newTestSchema but also returns the db and user so tests
// can set up fixture data directly without going through GraphQL.
func newTestEnv(t *testing.T) (*graphql.Schema, context.Context, *persistence.DB, *persistence.User) {
	t.Helper()

	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	user, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatalf("EnsureDevUser() error = %v", err)
	}

	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry()}
	schema := NewSchema(db, updater)
	ctx := auth.WithUser(context.Background(), user)

	return schema, ctx, db, user
}

// TestDocumentsLifecycle exercises the full import flow through GraphQL:
// upload (via importer.Service), list, confirm, delete.
func TestDocumentsLifecycle(t *testing.T) {
	schema, ctx, db, user := newTestEnv(t)

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
	doc, err := svc.Upload(ctx, user.ID, "test.csv", "text/csv", []byte(sampleCSV))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := svc.Process(ctx, doc.ID, user.ID); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// List documents via GraphQL; should see the PARSED document.
	data = exec(t, schema, ctx, `query {
		documents {
			id filename state detectedBank
			stagedTransactions { id type isin security { id } }
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

	// DEPOSIT_CASH has no security.
	if staged[1].(map[string]any)["security"] != nil {
		t.Error("staged[1].security should be nil for DEPOSIT_CASH")
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
		}) { id type source cashDelta { amount } }
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

	// Delete the document.
	exec(t, schema, ctx, `mutation ($id: ID!) { deleteDocument(id: $id) }`,
		map[string]any{"id": docID})

	data = exec(t, schema, ctx, `query { documents { id } }`, nil)
	if len(data["documents"].([]any)) != 0 {
		t.Errorf("documents after delete = %d, want 0", len(data["documents"].([]any)))
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
