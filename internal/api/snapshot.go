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
	return func(securityID string, at time.Time) (persistence.Decimal, bool) {
		listings, err := db.ListListings(ctx, securityID)
		if err != nil || len(listings) == 0 {
			return persistence.Decimal{}, false
		}

		quote, err := db.GetLatestQuoteBefore(ctx, persistence.GetLatestQuoteBeforeParams{
			ListingID: listings[0].ID,
			Time:      at,
		})
		if err != nil {
			return persistence.Decimal{}, false
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

	from, err := periodStart(args.Period, to, r.snap.FirstTransactionTime)
	if err != nil {
		return nil, err
	}

	return &PerformanceResolver{
		period: args.Period,
		from:   from,
		to:     to,
		twr:    finance.TimeWeightedReturn(r.txs, from, to, r.quote),
	}, nil
}

// periodStart returns the start of a Period ending at to. ALL_TIME starts at
// first, the earliest transaction, or at to if there is none.
func periodStart(period string, to, first time.Time) (time.Time, error) {
	switch period {
	case "ONE_WEEK":
		return to.AddDate(0, 0, -7), nil
	case "ONE_MONTH":
		return to.AddDate(0, -1, 0), nil
	case "SIX_MONTHS":
		return to.AddDate(0, -6, 0), nil
	case "YEAR_TO_DATE":
		return time.Date(to.Year(), time.January, 1, 0, 0, 0, 0, to.Location()), nil
	case "ONE_YEAR":
		return to.AddDate(-1, 0, 0), nil
	case "ALL_TIME":
		if first.IsZero() {
			return to, nil
		}
		return first, nil
	default:
		return time.Time{}, fmt.Errorf("unknown period %q", period)
	}
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

func (r *PositionResolver) PurchasePrice() Price {
	return price(r.position.PurchasePrice(), r.position.Currency)
}

func (r *PositionResolver) MarketPrice() Price {
	return price(r.position.MarketPrice, r.position.Currency)
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
