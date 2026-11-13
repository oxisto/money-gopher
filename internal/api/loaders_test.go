package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

// TestHandlerLoaders serves a transaction list through the HTTP handler,
// where every request carries its own loaders: the same account and the
// document links of a whole account are fetched once, not per row.
func TestHandlerLoaders(t *testing.T) {
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
	tagesgeld := h.account("Tagesgeld", "SAVINGS", nil)
	h.transfer(giro, tagesgeld, 10000, "2026-03-01")

	// One row from a document, so a link exists to load.
	svc := importer.NewService(db)
	doc, err := svc.Upload(ctx, person.ID, "giro.csv", "text/csv", []byte("Date,Type,Security,ISIN,Units,Price,Fees,Taxes,CashDelta,Currency\n2026-03-02,DEPOSIT_CASH,,,,,,,250.00,EUR\n"))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}
	exec(t, schema, ctx, `mutation ($id: ID!, $acct: ID!) { confirmImport(documentID: $id, input: { cashAccountID: $acct, transactions: [] }) { id } }`, map[string]any{"id": doc.ID, "acct": giro})

	body := `{"query":"{ cashAccounts { id displayName transactions { id type documents { id filename } document { id } cashAccount { displayName } counterpart { id } schedule { id } } } }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewHandler(schema).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Data struct {
			CashAccounts []struct {
				DisplayName  string
				Transactions []struct {
					Type        string
					Documents   []struct{ Filename string }
					Document    *struct{ ID string }
					CashAccount struct{ DisplayName string }
				}
			}
		}
		Errors []any
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v: %s", err, rec.Body.String())
	}
	if len(out.Errors) > 0 {
		t.Fatalf("errors: %v", out.Errors)
	}
	documented := 0
	for _, a := range out.Data.CashAccounts {
		for _, tx := range a.Transactions {
			if tx.CashAccount.DisplayName != a.DisplayName {
				t.Errorf("%s row resolved its account as %q", a.DisplayName, tx.CashAccount.DisplayName)
			}
			if len(tx.Documents) > 0 {
				documented++
				if tx.Document == nil || tx.Documents[0].Filename != "giro.csv" {
					t.Errorf("documented row = %+v, want giro.csv first", tx)
				}
			}
		}
	}
	if documented != 1 {
		t.Errorf("%d documented rows, want the imported deposit only", documented)
	}

	// Without loaders in the context the resolvers query directly and agree.
	data := exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { transactions { documents { filename } } } }`, map[string]any{"id": giro})
	direct := 0
	for _, raw := range data["cashAccount"].(map[string]any)["transactions"].([]any) {
		if len(raw.(map[string]any)["documents"].([]any)) > 0 {
			direct++
		}
	}
	if direct != 1 {
		t.Errorf("direct path: %d documented rows, want 1", direct)
	}
}
