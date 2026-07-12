package api

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

func TestSnapshot(t *testing.T) {
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
	schema, ctx := NewSchema(db, updater), auth.WithUser(context.Background(), user)

	// A portfolio with one security: 10 shares bought at 100.00 EUR.
	// The cash account must be created first so its ID can be passed to createPortfolio.
	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
		createSecurity(input: {
			displayName: "Vanguard FTSE All-World"
			listings: [{ ticker: "VWCE", currency: "EUR" }]
		}) { id listings { id } }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)
	security := data["createSecurity"].(map[string]any)
	securityID := security["id"].(string)
	listingID := security["listings"].([]any)[0].(map[string]any)["id"].(string)

	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		createPortfolio(input: { displayName: "Snapshot Portfolio", cashAccountID: $account }) { id }
	}`, map[string]any{"account": accountID})
	portfolioID := data["createPortfolio"].(map[string]any)["id"].(string)

	exec(t, schema, ctx, `mutation ($portfolio: ID!, $security: ID!, $account: ID!) {
		createTransaction(input: {
			type: BUY, time: "2026-01-03T10:00:00Z", currency: "EUR"
			portfolioID: $portfolio, securityID: $security, cashAccountID: $account
			units: 10, price: 10000
		}) { id }
	}`, map[string]any{"portfolio": portfolioID, "security": securityID, "account": accountID})

	// No quotes yet: the market value falls back to the purchase value.
	q := `query ($id: ID!, $time: Time) {
		portfolio(id: $id) {
			snapshot(time: $time) {
				firstTransactionTime
				positions {
					security { displayName }
					units
					purchaseValue { amount currency }
					marketValue { amount }
					gains
				}
				totalMarketValue { amount }
				totalProfitOrLoss { amount }
				totalGains
				performance(period: ALL_TIME) { timeWeightedReturn from to }
			}
		}
	}`

	snap := func(at string) map[string]any {
		vars := map[string]any{"id": portfolioID, "time": nil}
		if at != "" {
			vars["time"] = at
		}
		data := exec(t, schema, ctx, q, vars)
		return data["portfolio"].(map[string]any)["snapshot"].(map[string]any)
	}

	s := snap("")
	positions := s["positions"].([]any)
	if len(positions) != 1 {
		t.Fatalf("positions = %d, want 1", len(positions))
	}
	pos := positions[0].(map[string]any)
	if got := pos["units"].(float64); got != 10 {
		t.Errorf("units = %v, want 10", got)
	}
	if got := pos["purchaseValue"].(map[string]any)["amount"].(float64); got != 100000 {
		t.Errorf("purchase value = %v, want 100000", got)
	}
	if got := pos["marketValue"].(map[string]any)["amount"].(float64); got != 100000 {
		t.Errorf("market value without quotes = %v, want 100000", got)
	}
	if got := s["performance"].(map[string]any)["timeWeightedReturn"].(float64); got != 0 {
		t.Errorf("return without quotes = %v, want 0", got)
	}

	// Insert a quote: 110.00 EUR as of Feb 1.
	err = db.CreateQuote(ctx, persistence.CreateQuoteParams{
		ListingID: listingID,
		Time:      time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		Price:     11000,
	})
	if err != nil {
		t.Fatalf("CreateQuote() error = %v", err)
	}

	s = snap("")
	pos = s["positions"].([]any)[0].(map[string]any)
	if got := pos["marketValue"].(map[string]any)["amount"].(float64); got != 110000 {
		t.Errorf("market value with quote = %v, want 110000", got)
	}
	if got := pos["gains"].(float64); math.Abs(got-0.1) > 1e-9 {
		t.Errorf("gains = %v, want 0.1", got)
	}
	if got := s["totalProfitOrLoss"].(map[string]any)["amount"].(float64); got != 10000 {
		t.Errorf("total profit = %v, want 10000", got)
	}
	if got := s["performance"].(map[string]any)["timeWeightedReturn"].(float64); math.Abs(got-0.1) > 1e-9 {
		t.Errorf("all-time return = %v, want 0.1", got)
	}

	// A snapshot before the buy is empty, and one between buy and quote
	// still values at the purchase price.
	if s := snap("2026-01-01T00:00:00Z"); len(s["positions"].([]any)) != 0 {
		t.Errorf("positions before first buy = %v, want none", s["positions"])
	}
	s = snap("2026-01-15T00:00:00Z")
	pos = s["positions"].([]any)[0].(map[string]any)
	if got := pos["marketValue"].(map[string]any)["amount"].(float64); got != 100000 {
		t.Errorf("market value before quote = %v, want 100000", got)
	}
}
