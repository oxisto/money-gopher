package api

import (
	"context"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/finance"
)

// NetWorthHistory resolves Query.netWorthHistory.
func (r *RootResolver) NetWorthHistory(ctx context.Context, args struct {
	Period string
	Points int32
}) ([]*NetWorthPointResolver, error) {
	if args.Points < 2 || args.Points > 365 {
		return nil, fmt.Errorf("points must be between 2 and 365, got %d", args.Points)
	}

	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	txs, err := allPersonTransactions(ctx, r.db, person.ID)
	if err != nil {
		return nil, err
	}

	var first time.Time
	for _, tx := range txs {
		if first.IsZero() || tx.Time.Before(first) {
			first = tx.Time
		}
	}

	to := time.Now()
	from, err := periodStart(args.Period, to, first)
	if err != nil {
		return nil, err
	}

	step := to.Sub(from) / time.Duration(args.Points-1)
	times := make([]time.Time, args.Points)
	for i := range times {
		times[i] = from.Add(step * time.Duration(i))
	}
	// Land exactly on now, so the last point matches the current balances.
	times[len(times)-1] = to

	// All of a person's accounts share one currency for now; take it from
	// the transactions, like the snapshot totals do.
	currency := defaultCurrency
	if len(txs) > 0 {
		currency = txs[0].Currency
	}

	points := finance.NetWorthAt(txs, times, latestQuote(ctx, r.db))
	resolvers := make([]*NetWorthPointResolver, len(points))
	for i := range points {
		resolvers[i] = &NetWorthPointResolver{point: points[i], currency: currency}
	}
	return resolvers, nil
}

// NetWorthPointResolver resolves the NetWorthPoint GraphQL type.
type NetWorthPointResolver struct {
	point    finance.NetWorthPoint
	currency string
}

func (r *NetWorthPointResolver) Time() graphql.Time { return graphql.Time{Time: r.point.Time} }
func (r *NetWorthPointResolver) NetWorth() Money    { return money(r.point.NetWorth, r.currency) }
func (r *NetWorthPointResolver) Invested() Money    { return money(r.point.Invested, r.currency) }
