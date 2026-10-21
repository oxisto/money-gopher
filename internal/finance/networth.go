package finance

import (
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// NetWorthPoint is everything a person holds at one point in time.
type NetWorthPoint struct {
	Time time.Time
	// NetWorth is the market value of all held securities plus all cash
	// balances, in minor units.
	NetWorth int64
	// Invested is the net value brought in from outside up to Time:
	// deposits minus withdrawals, plus securities delivered in at their
	// price minus those delivered out. NetWorth − Invested is what the
	// money has earned.
	Invested int64
}

// NetWorthAt evaluates the net worth at each of the given times from all of
// a person's transactions (portfolio and cash account ones, deduplicated,
// any order) and a quote lookup.
func NetWorthAt(txs []*persistence.Transaction, times []time.Time, quote QuoteFunc) []NetWorthPoint {
	points := make([]NetWorthPoint, 0, len(times))
	for _, at := range times {
		p := NetWorthPoint{Time: at}
		p.NetWorth = marketValue(books(txs, at, false), at, quote)

		for _, tx := range txs {
			if tx.Time.After(at) {
				continue
			}
			// Every cash effect, trades included, lands on a cash account;
			// a transaction without one moves no cash.
			if tx.CashAccountID.Valid {
				p.NetWorth += tx.CashDelta
			}

			switch tx.Type {
			case "DEPOSIT_CASH", "WITHDRAW_CASH":
				// CashDelta is already negative for withdrawals.
				p.Invested += tx.CashDelta
			case "DELIVERY_INBOUND":
				p.Invested += tx.Price.MulFloat(tx.Units).MulInt(100).RoundToInt64()
			case "DELIVERY_OUTBOUND":
				p.Invested -= tx.Price.MulFloat(tx.Units).MulInt(100).RoundToInt64()
			}
		}

		points = append(points, p)
	}
	return points
}
