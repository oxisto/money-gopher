package fx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// euro is the currency the ECB publishes against; it needs no rate.
const euro = "EUR"

// ErrNoRate is returned when no rate is stored for a currency at or before
// the requested time.
var ErrNoRate = errors.New("no exchange rate known")

// RateStore is the part of the persistence layer Convert needs. Both
// *persistence.DB and *persistence.Queries satisfy it.
type RateStore interface {
	GetFxRateBefore(ctx context.Context, arg persistence.GetFxRateBeforeParams) (*persistence.FxRate, error)
}

// Convert expresses amount, denominated in from, in to at the reference
// rates in force at the given time: the newest stored at or before it. The
// result is exact; callers round when they turn it into money.
func Convert(ctx context.Context, store RateStore, amount persistence.Decimal, from, to string, at time.Time) (persistence.Decimal, error) {
	if from == to {
		return amount, nil
	}

	euros := amount
	if from != euro {
		rate, err := rateAt(ctx, store, from, at)
		if err != nil {
			return persistence.Decimal{}, err
		}
		euros = amount.Div(rate)
	}
	if to == euro {
		return euros, nil
	}

	rate, err := rateAt(ctx, store, to, at)
	if err != nil {
		return persistence.Decimal{}, err
	}
	return euros.Mul(rate), nil
}

func rateAt(ctx context.Context, store RateStore, currency string, at time.Time) (persistence.Decimal, error) {
	r, err := store.GetFxRateBefore(ctx, persistence.GetFxRateBeforeParams{Currency: currency, Time: at})
	if errors.Is(err, sql.ErrNoRows) {
		return persistence.Decimal{}, fmt.Errorf("%w for %s at %s", ErrNoRate, currency, at.Format("2006-01-02"))
	} else if err != nil {
		return persistence.Decimal{}, err
	}
	if r.Rate.Sign() <= 0 {
		return persistence.Decimal{}, fmt.Errorf("stored rate for %s at %s is %s", currency, at.Format("2006-01-02"), r.Rate)
	}
	return r.Rate, nil
}
