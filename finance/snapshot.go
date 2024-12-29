package finance

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"connectrpc.com/connect"
	moneygopher "github.com/oxisto/money-gopher"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/models"
	"github.com/oxisto/money-gopher/persistence"
)

// SnapshotDataProvider is an interface that provides the necessary data for
// building a snapshot. It includes methods for retrieving portfolio events and
// securities by their IDs.
type SnapshotDataProvider interface {
	ListPortfolioEventsByPortfolioID(ctx context.Context, portfolioID string) ([]*persistence.PortfolioEvent, error)
	ListSecuritiesByIDs(ctx context.Context, ids []string) ([]*persistence.Security, error)
}

// BuildSnapshot creates a snapshot of the portfolio at a given time. It
// calculates the performance and market value of the current positions and the
// total value of the portfolio.
//
// The snapshot is built by retrieving all events and security information from
// a [SnapshotDataProvider]. The snapshot is built by iterating over the events
// and calculating the positions at the specified timestamp.
func BuildSnapshot(
	ctx context.Context,
	timestamp *time.Time,
	portfolioID string,
	provider SnapshotDataProvider,
) (snap *models.PortfolioSnapshot, err error) {
	var (
		events []*persistence.PortfolioEvent
		p      portfoliov1.Portfolio
		m      map[string][]*portfoliov1.PortfolioEvent
		ids    []string
		secs   []*persistence.Security
		secmap map[string]*persistence.Security
	)

	// Retrieve events
	events, err = provider.ListPortfolioEventsByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Set up the snapshot
	snap = &models.PortfolioSnapshot{
		Time:      timestamp.Format(time.RFC3339),
		Positions: make([]*models.PortfolioPosition, 0),
		/*TotalPurchaseValue: portfoliov1.Zero(),
		TotalMarketValue:   portfoliov1.Zero(),
		TotalProfitOrLoss:  portfoliov1.Zero(),
		Cash:               portfoliov1.Zero(),*/
	}

	// Record the first transaction time
	if len(p.Events) > 0 {
		snap.FirstTransactionTime = events[0].Time.Format(time.RFC3339)
	}

	// Retrieve the event map; a map of events indexed by their security ID
	m = p.EventMap()
	ids = slices.Collect(maps.Keys(m))

	// Retrieve market value of filtered securities
	secs, err = provider.ListSecuritiesByIDs(context.Background(), ids)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("internal error while calling ListSecurities on securities service: %w", err),
		)
	}

	// Make a map out of the securities list so we can access it easier
	secmap = moneygopher.Map(secs, func(s *persistence.Security) string {
		return s.ID
	})

	// We need to look at the portfolio events up to the time of the snapshot
	// and calculate the current positions.
	for name, txs := range m {
		txs = eventsBefore(txs, timestamp)

		c := NewCalculation(txs)

		if name == "cash" {
			// Add deposited/withdrawn cash directly
			snap.Cash.PlusAssign(c.Cash)
			continue
		}

		if c.Amount == 0 {
			continue
		}

		// Also add cash that is part of a securities' transaction (e.g., sell/buy)
		snap.Cash.PlusAssign(c.Cash)

		pos := &portfoliov1.PortfolioPosition{
			Security:      secmap[name],
			Amount:        c.Amount,
			PurchaseValue: c.NetValue(),
			PurchasePrice: c.NetPrice(),
			MarketValue:   portfoliov1.Times(marketPrice(secmap, name, c.NetPrice()), c.Amount),
			MarketPrice:   marketPrice(secmap, name, c.NetPrice()),
		}

		// Calculate loss and gains
		pos.ProfitOrLoss = portfoliov1.Minus(pos.MarketValue, pos.PurchaseValue)
		pos.Gains = float64(portfoliov1.Minus(pos.MarketValue, pos.PurchaseValue).Value) / float64(pos.PurchaseValue.Value)

		// Add to total value(s)
		snap.TotalPurchaseValue.PlusAssign(pos.PurchaseValue)
		snap.TotalMarketValue.PlusAssign(pos.MarketValue)
		snap.TotalProfitOrLoss.PlusAssign(pos.ProfitOrLoss)

		// Store position in map
		snap.Positions[name] = pos
	}

	// Calculate total gains
	snap.TotalGains = float64(portfoliov1.Minus(snap.TotalMarketValue, snap.TotalPurchaseValue).Value) / float64(snap.TotalPurchaseValue.Value)

	// Calculate total portfolio value
	snap.TotalPortfolioValue = snap.TotalMarketValue.Plus(snap.Cash)

	return connect.NewResponse(snap), nil
}

// TODO: move to SQL query
func eventsBefore(events []*persistence.PortfolioEvent, t time.Time) (out []*persistence.PortfolioEvent) {
	out = make([]*persistence.PortfolioEvent, 0, len(events))

	for _, event := range events {
		if event.Time.After(t) {
			continue
		}

		out = append(out, event)
	}

	return
}
