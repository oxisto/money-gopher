package fx

import (
	"context"
	"log/slog"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// Updater keeps the fx_rates table current from a Source.
type Updater struct {
	DB     *persistence.DB
	Source Source
}

// UpdateAll stores the rates of every currency the data needs: those of
// the listings and of foreign-currency purchases, the euro aside. A
// currency seen for the first time gets the full history, so that past
// snapshots can be valued too; afterwards the recent file keeps the table
// current. It returns the number of rates stored.
func (u *Updater) UpdateAll(ctx context.Context) (int, error) {
	needed, err := u.neededCurrencies(ctx)
	if err != nil || len(needed) == 0 {
		return 0, err
	}

	history := false
	for currency := range needed {
		n, err := u.DB.CountFxRates(ctx, currency)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			history = true
		}
	}

	rates, err := u.Source.Rates(ctx, history)
	if err != nil {
		return 0, err
	}

	var stored int
	err = u.DB.Tx(ctx, func(q *persistence.Queries) error {
		stored = 0 // the transaction may be retried
		for _, r := range rates {
			if !needed[r.Currency] {
				continue
			}
			if err := q.CreateFxRate(ctx, persistence.CreateFxRateParams{
				Currency: r.Currency, Time: r.Time, Rate: r.Rate,
			}); err != nil {
				return err
			}
			stored++
		}
		return nil
	})
	return stored, err
}

// neededCurrencies collects the currencies a conversion can ask for.
func (u *Updater) neededCurrencies(ctx context.Context) (map[string]bool, error) {
	needed := map[string]bool{}

	listed, err := u.DB.ListListingCurrencies(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range listed {
		needed[c] = true
	}

	priced, err := u.DB.ListTransactionPriceCurrencies(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range priced {
		if c.Valid {
			needed[c.String] = true
		}
	}

	delete(needed, euro)
	return needed, nil
}

// Run updates the rates now and then every interval, until the context is
// cancelled. Errors are logged, not fatal.
func (u *Updater) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		stored, err := u.UpdateAll(ctx)
		if err != nil {
			slog.Error("fx rate update failed", "err", err)
		} else if stored > 0 {
			slog.Info("fx rates updated", "rates", stored)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
