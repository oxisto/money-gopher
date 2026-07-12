package persistence

import (
	"context"
	"testing"
	"time"
)

func TestOpenDB(t *testing.T) {
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	user, err := db.CreateUser(context.Background(), CreateUserParams{
		ID:          "user-1",
		Issuer:      "https://issuer.example.com",
		Subject:     "subject-1",
		DisplayName: "Money Gopher",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	got, err := db.GetUserByIdentity(context.Background(), GetUserByIdentityParams{
		Issuer:  "https://issuer.example.com",
		Subject: "subject-1",
	})
	if err != nil {
		t.Fatalf("GetUserByIdentity() error = %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("GetUserByIdentity() ID = %v, want %v", got.ID, user.ID)
	}
}

// TestQuoteTimeOrdering guards the time storage format: SQLite orders
// DATETIME columns as strings, so timestamps must be written in a fixed,
// lexicographically sortable format (UTC via _time_format=sqlite). A quote
// written with a non-UTC offset must not shadow a newer UTC quote.
func TestQuoteTimeOrdering(t *testing.T) {
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()

	_, err = db.CreateSecurity(ctx, CreateSecurityParams{ID: "sec", DisplayName: "Test"})
	if err != nil {
		t.Fatalf("CreateSecurity() error = %v", err)
	}
	listing, err := db.CreateListing(ctx, CreateListingParams{
		ID: "l1", SecurityID: "sec", Ticker: "TST", Currency: "EUR",
	})
	if err != nil {
		t.Fatalf("CreateListing() error = %v", err)
	}

	// Older quote carrying a +02:00 offset, newer quote in UTC.
	older := time.Date(2026, 7, 1, 12, 0, 0, 0, time.FixedZone("CEST", 2*60*60))
	newer := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	for i, q := range []struct {
		at    time.Time
		price int64
	}{{older, 100}, {newer, 200}} {
		err = db.CreateQuote(ctx, CreateQuoteParams{ListingID: listing.ID, Time: q.at.UTC(), Price: q.price})
		if err != nil {
			t.Fatalf("CreateQuote(%d) error = %v", i, err)
		}
	}

	latest, err := db.GetLatestQuote(ctx, listing.ID)
	if err != nil {
		t.Fatalf("GetLatestQuote() error = %v", err)
	}
	if latest.Price != 200 {
		t.Errorf("latest price = %d, want 200 (time ordering is broken)", latest.Price)
	}

	before, err := db.GetLatestQuoteBefore(ctx, GetLatestQuoteBeforeParams{
		ListingID: listing.ID,
		Time:      time.Date(2026, 7, 1, 23, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("GetLatestQuoteBefore() error = %v", err)
	}
	if before.Price != 100 {
		t.Errorf("quote before = %d, want 100", before.Price)
	}
}
