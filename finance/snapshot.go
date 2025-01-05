package finance

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	moneygopher "github.com/oxisto/money-gopher"
	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/models"
	"github.com/oxisto/money-gopher/persistence"
)

// SnapshotDataProvider is an interface that provides the necessary data for
// building a snapshot. It includes methods for retrieving portfolio events and
// securities by their IDs.
type SnapshotDataProvider interface {
	ListListedSecuritiesBySecurityID(ctx context.Context, securityID string) ([]*persistence.ListedSecurity, error)
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
	timestamp time.Time,
	portfolioID string,
	provider SnapshotDataProvider,
) (snap *models.PortfolioSnapshot, err error) {
	var (
		events []*persistence.PortfolioEvent
		m      map[string][]*persistence.PortfolioEvent
		ids    []string
		secs   []*persistence.Security
		secmap map[string]*persistence.Security
	)

	// Retrieve events
	events, err = provider.ListPortfolioEventsByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	// Set up the snapshot
	snap = &models.PortfolioSnapshot{
		Time:               timestamp.Format(time.RFC3339),
		Positions:          make([]*models.PortfolioPosition, 0),
		TotalPurchaseValue: currency.Zero(),
		TotalMarketValue:   currency.Zero(),
		TotalProfitOrLoss:  currency.Zero(),
		Cash:               currency.Zero(),
	}

	// Record the first transaction time
	if len(events) > 0 {
		snap.FirstTransactionTime = events[0].Time.Format(time.RFC3339)
	}

	// Retrieve the event map; a map of events indexed by their security ID
	m = groupByPortfolio(events)
	ids = slices.Collect(maps.Keys(m))

	// Retrieve market value of filtered securities
	secs, err = provider.ListSecuritiesByIDs(context.Background(), ids)

	if err != nil {
		return nil, fmt.Errorf("internal error while calling ListSecurities on securities service: %w", err)
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

		pos := &models.PortfolioPosition{
			Security:      secmap[name],
			Amount:        c.Amount,
			PurchaseValue: c.NetValue(),
			PurchasePrice: c.NetPrice(),
			MarketValue:   currency.Times(marketPrice(name, c.NetPrice(), provider), c.Amount),
			MarketPrice:   marketPrice(name, c.NetPrice(), provider),
		}

		// Calculate loss and gains
		pos.ProfitOrLoss = currency.Minus(pos.MarketValue, pos.PurchaseValue)
		pos.Gains = float64(currency.Minus(pos.MarketValue, pos.PurchaseValue).Amount) / float64(pos.PurchaseValue.Amount)

		// Add to total value(s)
		snap.TotalPurchaseValue.PlusAssign(pos.PurchaseValue)
		snap.TotalMarketValue.PlusAssign(pos.MarketValue)
		snap.TotalProfitOrLoss.PlusAssign(pos.ProfitOrLoss)

		// Store position in map
		snap.Positions = append(snap.Positions, pos)
	}

	// Calculate total gains
	snap.TotalGains = float64(currency.Minus(snap.TotalMarketValue, snap.TotalPurchaseValue).Amount) / float64(snap.TotalPurchaseValue.Amount)

	// Calculate total portfolio value
	snap.TotalPortfolioValue = snap.TotalMarketValue.Plus(snap.Cash)

	return snap, nil
}

// eventsBefore returns all events that occurred before a given time.
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

// groupByPortfolio groups the events by their security ID.
func groupByPortfolio(events []*persistence.PortfolioEvent) (m map[string][]*persistence.PortfolioEvent) {
	m = make(map[string][]*persistence.PortfolioEvent)

	for _, event := range events {
		name := event.SecurityID
		if name != "" {
			m[name] = append(m[name], event)
		} else {
			// a little bit of a hack
			m["cash"] = append(m["cash"], event)
		}
	}

	return
}

func marketPrice(
	name string,
	netPrice *currency.Currency,
	provider SnapshotDataProvider,
) *currency.Currency {
	ls, _ := provider.ListListedSecuritiesBySecurityID(context.Background(), name)

	if ls == nil || ls[0].LatestQuote == nil {
		return netPrice
	} else {
		return ls[0].LatestQuote
	}
}
