package persistence

import (
	"context"
	"database/sql"
	"errors"
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
		price string
	}{{older, "100"}, {newer, "200"}} {
		price, err := NewDecimal(q.price)
		if err != nil {
			t.Fatalf("NewDecimal(%d) error = %v", i, err)
		}
		err = db.CreateQuote(ctx, CreateQuoteParams{ListingID: listing.ID, Time: q.at.UTC(), Price: price})
		if err != nil {
			t.Fatalf("CreateQuote(%d) error = %v", i, err)
		}
	}

	latest, err := db.GetLatestQuote(ctx, listing.ID)
	if err != nil {
		t.Fatalf("GetLatestQuote() error = %v", err)
	}
	if latest.Price.StringPrec(0) != "200" {
		t.Errorf("latest price = %s, want 200 (time ordering is broken)", latest.Price.StringPrec(0))
	}

	before, err := db.GetLatestQuoteBefore(ctx, GetLatestQuoteBeforeParams{
		ListingID: listing.ID,
		Time:      time.Date(2026, 7, 1, 23, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("GetLatestQuoteBefore() error = %v", err)
	}
	if before.Price.StringPrec(0) != "100" {
		t.Errorf("quote before = %s, want 100", before.Price.StringPrec(0))
	}
}

// TestTx_RetriesWriteConflict forces a write conflict on the first attempt —
// another transaction holds an uncommitted write to the same row — and checks
// Tx re-runs fn instead of returning the serialization failure.
func TestTx_RetriesWriteConflict(t *testing.T) {
	ctx := context.Background()

	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	if _, err := db.CreatePerson(ctx, CreatePersonParams{ID: "person-1", DisplayName: "Before"}); err != nil {
		t.Fatalf("CreatePerson() error = %v", err)
	}

	other, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}
	defer other.Rollback()
	if _, err := db.WithTx(other).UpdatePersonDisplayName(ctx, UpdatePersonDisplayNameParams{ID: "person-1", DisplayName: "Concurrent"}); err != nil {
		t.Fatalf("UpdatePersonDisplayName() (other) error = %v", err)
	}

	attempts := 0
	err = db.Tx(ctx, func(q *Queries) error {
		attempts++

		_, err := q.UpdatePersonDisplayName(ctx, UpdatePersonDisplayNameParams{ID: "person-1", DisplayName: "After"})
		if attempts == 1 {
			// The other transaction finishes only after it made this attempt
			// fail, so the retry finds the row free.
			if cerr := other.Commit(); cerr != nil {
				t.Fatalf("Commit() (other) error = %v", cerr)
			}
		}
		return err
	})
	if err != nil {
		t.Fatalf("Tx() error = %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}

	got, err := db.GetPerson(ctx, "person-1")
	if err != nil {
		t.Fatalf("GetPerson() error = %v", err)
	}
	if got.DisplayName != "After" {
		t.Errorf("DisplayName = %q, want %q", got.DisplayName, "After")
	}
}

// TestCreatePersonWithAccess_RollsBackOnFailedGrant makes the grant fail —
// the user doesn't exist, so the foreign key rejects it — and checks the
// person was not left behind without access.
func TestCreatePersonWithAccess_RollsBackOnFailedGrant(t *testing.T) {
	ctx := context.Background()

	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	_, err = db.CreatePersonWithAccess(ctx, CreatePersonParams{ID: "person-1", DisplayName: "Orphan"}, "no-such-user")
	if err == nil {
		t.Fatal("CreatePersonWithAccess() error = nil, want a foreign key violation")
	}

	if _, err := db.GetPerson(ctx, "person-1"); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("GetPerson() error = %v, want sql.ErrNoRows (rolled back)", err)
	}
}

// TestTxWith_RepeatableReadRetriesStaleRead commits a change between fn's read
// and its write. At REPEATABLE READ that is a write conflict, so Tx re-runs fn,
// which then reads the change instead of overwriting it.
func TestTxWith_RepeatableReadRetriesStaleRead(t *testing.T) {
	ctx := context.Background()

	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	if _, err := db.CreatePerson(ctx, CreatePersonParams{ID: "person-1", DisplayName: "Before"}); err != nil {
		t.Fatalf("CreatePerson() error = %v", err)
	}

	attempts := 0
	var read string
	err = db.TxWith(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead}, func(q *Queries) error {
		attempts++

		p, err := q.GetPerson(ctx, "person-1")
		if err != nil {
			return err
		}
		read = p.DisplayName

		if attempts == 1 {
			if _, err := db.UpdatePersonDisplayName(ctx, UpdatePersonDisplayNameParams{ID: "person-1", DisplayName: "Concurrent"}); err != nil {
				return err
			}
		}

		_, err = q.UpdatePersonDisplayName(ctx, UpdatePersonDisplayNameParams{ID: "person-1", DisplayName: read + "!"})
		return err
	})
	if err != nil {
		t.Fatalf("TxWith() error = %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}

	got, err := db.GetPerson(ctx, "person-1")
	if err != nil {
		t.Fatalf("GetPerson() error = %v", err)
	}
	if got.DisplayName != "Concurrent!" {
		t.Errorf("DisplayName = %q, want %q (built on the concurrent change)", got.DisplayName, "Concurrent!")
	}
}
