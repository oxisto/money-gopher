package api

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

// TestSnapshot_ForeignListing values a US Treasury, quoted in USD, in the
// euros its portfolio settles in: the cost basis at the bank's rate on the
// trade, the market value at the ECB rate of the snapshot day.
func TestSnapshot_ForeignListing(t *testing.T) {
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
	schema, ctx := NewSchema(db, updater), auth.WithPerson(auth.WithUser(context.Background(), user), person)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
		createAsset(input: {
			displayName: "4,125 % United States of America"
			identifiers: [{ kind: "ISIN", value: "US91282CFV81" }]
			listings: [{ ticker: "US91282CFV81", currency: "USD" }]
		}) { id listings { id } }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)
	security := data["createAsset"].(map[string]any)
	assetID := security["id"].(string)
	listingID := security["listings"].([]any)[0].(map[string]any)["id"].(string)

	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		createPortfolio(input: { displayName: "Bonds", cashAccountID: $account }) { id }
	}`, map[string]any{"account": accountID})
	portfolioID := data["createPortfolio"].(map[string]any)["id"].(string)

	// 300 × 98.11 USD at 1.0518, as the ING confirmation states it; the
	// API has no input for the FX fields, the importer writes them.
	bought := time.Date(2024, 11, 18, 10, 0, 0, 0, time.UTC)
	price, _ := persistence.NewDecimal("98.11")
	rate, _ := persistence.NewDecimal("1.0518")
	if _, err := db.CreateTransaction(ctx, persistence.CreateTransactionParams{
		ID: "buy", Type: "BUY", Time: bought,
		PortfolioID:   sql.NullString{String: portfolioID, Valid: true},
		SecurityID:    sql.NullString{String: assetID, Valid: true},
		CashAccountID: sql.NullString{String: accountID, Valid: true},
		Units:         300,
		Price:         price,
		CashDelta:     -2806961,
		Currency:      "EUR",
		PriceCurrency: sql.NullString{String: "USD", Valid: true},
		FxRate:        rate,
	}); err != nil {
		t.Fatal(err)
	}

	q := `query ($id: ID!) {
		portfolio(id: $id) {
			snapshot {
				positions {
					purchasePrice { amount currency }
					purchaseValue { amount currency }
					marketPrice { amount currency }
					marketValue { amount currency }
				}
			}
		}
	}`
	position := func() map[string]any {
		data := exec(t, schema, ctx, q, map[string]any{"id": portfolioID})
		return data["portfolio"].(map[string]any)["snapshot"].(map[string]any)["positions"].([]any)[0].(map[string]any)
	}
	// Money amounts are minor units; Price amounts are decimal strings.
	moneyOf := func(m map[string]any, field string) (float64, string) {
		v := m[field].(map[string]any)
		return v["amount"].(float64), v["currency"].(string)
	}
	priceOf := func(m map[string]any, field string) (string, string) {
		v := m[field].(map[string]any)
		return v["amount"].(string), v["currency"].(string)
	}

	// The cost basis is in euros, and without a usable quote the market
	// value is the cost basis — not the USD figure.
	pos := position()
	if a, c := moneyOf(pos, "purchaseValue"); a != 2798346 || c != "EUR" {
		t.Errorf("purchaseValue = %v %s, want 2798346 EUR", a, c)
	}
	if a, c := priceOf(pos, "purchasePrice"); a[:5] != "93.27" || c != "EUR" {
		t.Errorf("purchasePrice = %s %s, want 93.278… EUR", a, c)
	}
	if a, _ := moneyOf(pos, "marketValue"); a != 2798346 {
		t.Errorf("marketValue without quotes = %v, want the cost basis 2798346", a)
	}

	// A USD quote alone changes nothing: with no rate it cannot be valued
	// in euros, so the fallback stays rather than mixing currencies.
	quoted := time.Date(2026, 10, 21, 16, 0, 0, 0, time.UTC)
	usd, _ := persistence.NewDecimal("95.552")
	if err := db.CreateQuote(ctx, persistence.CreateQuoteParams{ListingID: listingID, Time: quoted, Price: usd}); err != nil {
		t.Fatal(err)
	}
	if a, _ := moneyOf(position(), "marketValue"); a != 2798346 {
		t.Errorf("marketValue with a quote but no rate = %v, want the cost basis", a)
	}

	// With the day's rate the quote is converted: 300 × 95.552 / 1.1578.
	ecb, _ := persistence.NewDecimal("1.1578")
	if err := db.CreateFxRate(ctx, persistence.CreateFxRateParams{Currency: "USD", Time: quoted.Truncate(24 * time.Hour), Rate: ecb}); err != nil {
		t.Fatal(err)
	}
	pos = position()
	if a, c := moneyOf(pos, "marketValue"); a != 2475868 || c != "EUR" {
		t.Errorf("marketValue = %v %s, want 2475868 EUR", a, c)
	}
	if a, c := priceOf(pos, "marketPrice"); a[:5] != "82.52" || c != "EUR" {
		t.Errorf("marketPrice = %s %s, want 82.528… EUR", a, c)
	}
}
