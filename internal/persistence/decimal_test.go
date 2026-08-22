package persistence

import (
	"context"
	"testing"
	"time"
)

// TestDecimal_RoundTripsAtDeclaredScale writes a value with more decimal
// digits than any previously-seen document needed, through a real
// DECIMAL(19,8) column (quotes.price), and reads it back. This is the
// assertion that would have failed before persistence.Decimal's
// String()/Value() were fixed: they were hardcoded to StringPrec(4)
// independent of the column's actual declared width, so widening the
// schema to DECIMAL(19,8) alone changed nothing — the extra precision was
// still being silently truncated at serialization. A schema type is only
// true if every layer agrees with it; this test is the check that they do.
func TestDecimal_RoundTripsAtDeclaredScale(t *testing.T) {
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	security, err := db.CreateSecurity(ctx, CreateSecurityParams{ID: "sec", DisplayName: "Test"})
	if err != nil {
		t.Fatalf("CreateSecurity() error = %v", err)
	}
	listing, err := db.CreateListing(ctx, CreateListingParams{
		ID: "l1", SecurityID: security.ID, Ticker: "TST", Currency: "EUR",
	})
	if err != nil {
		t.Fatalf("CreateListing() error = %v", err)
	}

	const want = "1.05543234" // 8 decimal digits — the column's full declared width.
	price, err := NewDecimal(want)
	if err != nil {
		t.Fatalf("NewDecimal() error = %v", err)
	}

	if err := db.CreateQuote(ctx, CreateQuoteParams{
		ListingID: listing.ID,
		Time:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Price:     price,
	}); err != nil {
		t.Fatalf("CreateQuote() error = %v", err)
	}

	quote, err := db.GetLatestQuote(ctx, listing.ID)
	if err != nil {
		t.Fatalf("GetLatestQuote() error = %v", err)
	}
	if got := quote.Price.StringPrec(8); got != want {
		t.Errorf("round-tripped price = %s, want %s — a layer between the Go value and the column is truncating precision the DECIMAL(19,8) column itself allows", got, want)
	}
}
