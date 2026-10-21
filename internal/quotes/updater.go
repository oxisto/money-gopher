package quotes

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// Updater fetches the latest quotes for listings from their configured
// providers and appends them to the quote history.
type Updater struct {
	DB       *persistence.DB
	Registry *Registry
}

// UpdateAll updates every listing that has a quote provider. It returns the
// number of listings that got a new quote and the collected errors of the
// ones that failed; one broken provider does not stop the others.
func (u *Updater) UpdateAll(ctx context.Context) (int, error) {
	listings, err := u.DB.ListListingsWithProvider(ctx)
	if err != nil {
		return 0, err
	}

	return u.update(ctx, listings)
}

// UpdateSecurities updates all provider-backed listings of the given
// securities.
func (u *Updater) UpdateSecurities(ctx context.Context, securityIDs []string) (int, error) {
	var listings []*persistence.Listing
	for _, id := range securityIDs {
		ls, err := u.DB.ListListings(ctx, id)
		if err != nil {
			return 0, err
		}
		for _, l := range ls {
			if l.QuoteProvider.Valid {
				listings = append(listings, l)
			}
		}
	}

	return u.update(ctx, listings)
}

func (u *Updater) update(ctx context.Context, listings []*persistence.Listing) (updated int, errs error) {
	for _, listing := range listings {
		if err := u.updateListing(ctx, listing); err != nil {
			errs = errors.Join(errs, fmt.Errorf("listing %s (%s): %w", listing.ID, listing.Ticker, err))
			continue
		}
		updated++
	}

	return updated, errs
}

// updateListing fetches and stores one quote.
func (u *Updater) updateListing(ctx context.Context, listing *persistence.Listing) error {
	provider, ok := u.Registry.Get(listing.QuoteProvider.String)
	if !ok {
		return fmt.Errorf("unknown quote provider %q", listing.QuoteProvider.String)
	}

	ins := Instrument{
		Ticker:   listing.Ticker,
		Exchange: listing.Exchange.String,
		Currency: listing.Currency,
	}

	// Providers like ING identify instruments by external identifiers.
	identifiers, err := u.DB.ListSecurityIdentifiers(ctx, listing.SecurityID)
	if err != nil {
		return err
	}
	for _, id := range identifiers {
		switch id.Kind {
		case "ISIN":
			ins.ISIN = id.Value
		case "WKN":
			ins.WKN = id.Value
		}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	quote, err := provider.LatestQuote(ctx, ins)
	if err != nil {
		return err
	}

	return u.DB.CreateQuote(ctx, persistence.CreateQuoteParams{
		ListingID: listing.ID,
		// Stored in UTC so the string representation sorts chronologically.
		Time:  quote.Time.UTC(),
		Price: quote.Price,
	})
}

// Run updates all quotes now and then every interval, until the context is
// cancelled. Errors are logged, not fatal.
func (u *Updater) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		updated, err := u.UpdateAll(ctx)
		if err != nil {
			slog.Error("quote update failed", "err", err, "updated", updated)
		} else if updated > 0 {
			slog.Info("quotes updated", "listings", updated)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
