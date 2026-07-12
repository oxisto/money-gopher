package api

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/finance"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// defaultCurrency is used for the totals of an empty snapshot, which have no
// transactions to take a currency from.
const defaultCurrency = "EUR"

// SnapshotResolver resolves the PortfolioSnapshot GraphQL type. It keeps the
// transactions and the quote lookup around so that performance() can replay
// them over arbitrary periods.
type SnapshotResolver struct {
	db    *persistence.DB
	txs   []*persistence.Transaction
	snap  *finance.Snapshot
	quote finance.QuoteFunc
}

// Snapshot resolves Portfolio.snapshot.
func (r *PortfolioResolver) Snapshot(ctx context.Context, args struct{ Time *graphql.Time }) (*SnapshotResolver, error) {
	at := time.Now()
	if args.Time != nil {
		at = args.Time.Time
	}

	txs, err := r.db.ListTransactionsByPortfolio(ctx, sql.NullString{String: r.portfolio.ID, Valid: true})
	if err != nil {
		return nil, err
	}

	quote := latestQuote(ctx, r.db)

	return &SnapshotResolver{
		db:    r.db,
		txs:   txs,
		snap:  finance.SnapshotAt(txs, at, quote),
		quote: quote,
	}, nil
}

// latestQuote returns a finance.QuoteFunc backed by the quotes table: the
// newest quote of the security's first listing at or before the requested
// time. Securities without listings or quotes report ok == false, making
// calculations fall back to the purchase price.
func latestQuote(ctx context.Context, db *persistence.DB) finance.QuoteFunc {
	return func(securityID string, at time.Time) (int64, bool) {
		listings, err := db.ListListings(ctx, securityID)
		if err != nil || len(listings) == 0 {
			return 0, false
		}

		quote, err := db.GetLatestQuoteBefore(ctx, persistence.GetLatestQuoteBeforeParams{
			ListingID: listings[0].ID,
			Time:      at,
		})
		if err != nil {
			return 0, false
		}

		return quote.Price, true
	}
}

func (r *SnapshotResolver) Time() graphql.Time {
	return graphql.Time{Time: r.snap.Time}
}

func (r *SnapshotResolver) FirstTransactionTime() *graphql.Time {
	if r.snap.FirstTransactionTime.IsZero() {
		return nil
	}
	return &graphql.Time{Time: r.snap.FirstTransactionTime}
}

// Positions resolves PortfolioSnapshot.positions, sorted by security name.
func (r *SnapshotResolver) Positions(ctx context.Context) ([]*PositionResolver, error) {
	resolvers := make([]*PositionResolver, 0, len(r.snap.Positions))
	for _, position := range r.snap.Positions {
		security, err := r.db.GetSecurity(ctx, position.SecurityID)
		if err != nil {
			return nil, fmt.Errorf("could not load security %q: %w", position.SecurityID, err)
		}

		resolvers = append(resolvers, &PositionResolver{db: r.db, position: position, security: security})
	}

	sort.Slice(resolvers, func(i, j int) bool {
		return resolvers[i].security.DisplayName < resolvers[j].security.DisplayName
	})

	return resolvers, nil
}

// currency is the currency of the snapshot totals.
func (r *SnapshotResolver) currency() string {
	if r.snap.Currency == "" {
		return defaultCurrency
	}
	return r.snap.Currency
}

func (r *SnapshotResolver) TotalPurchaseValue() Money {
	return money(r.snap.TotalPurchaseValue, r.currency())
}

func (r *SnapshotResolver) TotalMarketValue() Money {
	return money(r.snap.TotalMarketValue, r.currency())
}

func (r *SnapshotResolver) TotalProfitOrLoss() Money {
	return money(r.snap.TotalProfitOrLoss(), r.currency())
}

func (r *SnapshotResolver) TotalGains() float64 {
	return r.snap.TotalGains()
}

// Performance resolves PortfolioSnapshot.performance.
func (r *SnapshotResolver) Performance(args struct{ Period string }) (*PerformanceResolver, error) {
	to := r.snap.Time

	var from time.Time
	switch args.Period {
	case "ONE_WEEK":
		from = to.AddDate(0, 0, -7)
	case "ONE_MONTH":
		from = to.AddDate(0, -1, 0)
	case "SIX_MONTHS":
		from = to.AddDate(0, -6, 0)
	case "YEAR_TO_DATE":
		from = time.Date(to.Year(), time.January, 1, 0, 0, 0, 0, to.Location())
	case "ONE_YEAR":
		from = to.AddDate(-1, 0, 0)
	case "ALL_TIME":
		from = r.snap.FirstTransactionTime
		if from.IsZero() {
			from = to
		}
	default:
		return nil, fmt.Errorf("unknown period %q", args.Period)
	}

	return &PerformanceResolver{
		period: args.Period,
		from:   from,
		to:     to,
		twr:    finance.TimeWeightedReturn(r.txs, from, to, r.quote),
	}, nil
}

// PositionResolver resolves the Position GraphQL type.
type PositionResolver struct {
	db       *persistence.DB
	position *finance.Position
	security *persistence.Security
}

func (r *PositionResolver) Security() *SecurityResolver {
	return &SecurityResolver{db: r.db, security: r.security}
}

func (r *PositionResolver) Units() float64 {
	return r.position.Units
}

func (r *PositionResolver) PurchaseValue() Money {
	return money(r.position.PurchaseValue, r.position.Currency)
}

func (r *PositionResolver) PurchasePrice() Money {
	return money(r.position.PurchasePrice(), r.position.Currency)
}

func (r *PositionResolver) MarketPrice() Money {
	return money(r.position.MarketPrice, r.position.Currency)
}

func (r *PositionResolver) MarketValue() Money {
	return money(r.position.MarketValue, r.position.Currency)
}

func (r *PositionResolver) ProfitOrLoss() Money {
	return money(r.position.ProfitOrLoss(), r.position.Currency)
}

func (r *PositionResolver) Gains() float64 {
	return r.position.Gains()
}

// PerformanceResolver resolves the Performance GraphQL type.
type PerformanceResolver struct {
	period string
	from   time.Time
	to     time.Time
	twr    float64
}

func (r *PerformanceResolver) Period() string {
	return r.period
}

func (r *PerformanceResolver) From() graphql.Time {
	return graphql.Time{Time: r.from}
}

func (r *PerformanceResolver) To() graphql.Time {
	return graphql.Time{Time: r.to}
}

func (r *PerformanceResolver) TimeWeightedReturn() float64 {
	return r.twr
}
