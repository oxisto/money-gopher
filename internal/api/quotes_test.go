package api

import (
	"context"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

// fakeProvider serves a fixed quote under the name "fake".
type fakeProvider struct{}

func (fakeProvider) Name() string { return "fake" }

func (fakeProvider) LatestQuote(context.Context, quotes.Instrument) (quotes.Quote, error) {
	return quotes.Quote{
		Price: 11111,
		Time:  time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC),
	}, nil
}

func TestTriggerQuoteUpdate(t *testing.T) {
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	user, person, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatalf("EnsureDevUser() error = %v", err)
	}

	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry(fakeProvider{})}
	schema, ctx := NewSchema(db, updater), auth.WithPerson(auth.WithUser(context.Background(), user), person)

	exec(t, schema, ctx, `mutation {
		createSecurity(input: {
			displayName: "Vanguard FTSE All-World"
			listings: [{ ticker: "VWCE", currency: "EUR", quoteProvider: "fake" }]
		}) { id }
	}`, nil)

	data := exec(t, schema, ctx, `mutation {
		triggerQuoteUpdate { updatedListings errors }
	}`, nil)
	result := data["triggerQuoteUpdate"].(map[string]any)
	if got := result["updatedListings"].(float64); got != 1 {
		t.Errorf("updatedListings = %v, want 1", got)
	}
	if got := result["errors"].([]any); len(got) != 0 {
		t.Errorf("errors = %v, want none", got)
	}

	// The quote is now visible as latestQuote and in the history.
	data = exec(t, schema, ctx, `{
		securities {
			listings {
				latestQuote { price { amount currency } time }
				quotes { price { amount } }
			}
		}
	}`, nil)
	listing := data["securities"].([]any)[0].(map[string]any)["listings"].([]any)[0].(map[string]any)
	latest := listing["latestQuote"].(map[string]any)
	if got := latest["price"].(map[string]any)["amount"].(float64); got != 11111 {
		t.Errorf("latest quote = %v, want 11111", got)
	}
	if got := len(listing["quotes"].([]any)); got != 1 {
		t.Errorf("quote history length = %v, want 1", got)
	}
}
