package quotes

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// fake is a Provider returning a fixed quote, remembering what it was asked.
type fake struct {
	name string
	ins  []Instrument
	err  error
}

func (f *fake) Name() string { return f.name }

func (f *fake) LatestQuote(_ context.Context, ins Instrument) (Quote, error) {
	f.ins = append(f.ins, ins)
	if f.err != nil {
		return Quote{}, f.err
	}
	return Quote{Price: 12345, Time: time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)}, nil
}

func TestUpdater(t *testing.T) {
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()

	// One security with an ISIN and two listings: one with a provider, one
	// without (which must be left alone).
	security, err := db.CreateSecurity(ctx, persistence.CreateSecurityParams{
		ID: "sec", DisplayName: "Test Security",
	})
	if err != nil {
		t.Fatalf("CreateSecurity() error = %v", err)
	}
	err = db.CreateSecurityIdentifier(ctx, persistence.CreateSecurityIdentifierParams{
		SecurityID: security.ID, Kind: "ISIN", Value: "IE00BK5BQT80",
	})
	if err != nil {
		t.Fatalf("CreateSecurityIdentifier() error = %v", err)
	}
	withProvider, err := db.CreateListing(ctx, persistence.CreateListingParams{
		ID: "l1", SecurityID: security.ID, Ticker: "TST", Currency: "EUR",
		QuoteProvider: sql.NullString{String: "fake", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateListing() error = %v", err)
	}
	_, err = db.CreateListing(ctx, persistence.CreateListingParams{
		ID: "l2", SecurityID: security.ID, Ticker: "TST2", Currency: "EUR",
	})
	if err != nil {
		t.Fatalf("CreateListing() error = %v", err)
	}

	provider := &fake{name: "fake"}
	updater := &Updater{DB: db, Registry: NewRegistry(provider)}

	updated, err := updater.UpdateAll(ctx)
	if err != nil {
		t.Fatalf("UpdateAll() error = %v", err)
	}
	if updated != 1 {
		t.Errorf("updated = %d, want 1", updated)
	}

	// The provider got the listing attributes and the ISIN.
	if len(provider.ins) != 1 {
		t.Fatalf("provider calls = %d, want 1", len(provider.ins))
	}
	if ins := provider.ins[0]; ins.Ticker != "TST" || ins.ISIN != "IE00BK5BQT80" {
		t.Errorf("instrument = %+v", ins)
	}

	// The quote landed in the history.
	quote, err := db.GetLatestQuote(ctx, withProvider.ID)
	if err != nil {
		t.Fatalf("GetLatestQuote() error = %v", err)
	}
	if quote.Price != 12345 {
		t.Errorf("stored price = %d, want 12345", quote.Price)
	}

	// A failing provider is reported but counted as not updated.
	provider.err = errors.New("boom")
	updated, err = updater.UpdateAll(ctx)
	if updated != 0 || err == nil {
		t.Errorf("UpdateAll() with failing provider = %d, %v; want 0 and an error", updated, err)
	}

	// UpdateSecurities goes through the same path.
	provider.err = nil
	updated, err = updater.UpdateSecurities(ctx, []string{security.ID})
	if err != nil || updated != 1 {
		t.Errorf("UpdateSecurities() = %d, %v; want 1, nil", updated, err)
	}
}
