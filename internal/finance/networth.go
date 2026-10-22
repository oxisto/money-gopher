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
	// Estimated is the part of NetWorth that is declared rather than
	// tracked: the declared balances of estimated accounts.
	Estimated int64
}

// NetWorthAt evaluates the net worth at each of the given times from all of
// a person's transactions (portfolio and cash account ones, deduplicated,
// any order), the declared balances of the estimated accounts by account
// ID, and a quote lookup.
//
// An estimated account is outside the tracked world: its declared balance
// counts as a constant, its transactions never move it. Pure cash events
// on it (deposits, withdrawals, interest, fees, tax) are therefore ignored
// altogether, while a trade or dividend settling on it is money from or
// to outside — a purchase paid from the estimated Giro is capital brought
// in, a dividend landing there is earned and taken out — so it counts in
// Invested with the opposite sign instead of as cash.
func NetWorthAt(txs []*persistence.Transaction, estimated map[string]int64, times []time.Time, quote QuoteFunc) []NetWorthPoint {
	var declared int64
	for _, balance := range estimated {
		declared += balance
	}

	points := make([]NetWorthPoint, 0, len(times))
	for _, at := range times {
		p := NetWorthPoint{Time: at, Estimated: declared}
		p.NetWorth = marketValue(books(txs, at, false), at, quote) + declared

		for _, tx := range txs {
			if tx.Time.After(at) {
				continue
			}
			if tx.CashAccountID.Valid {
				if _, isEstimated := estimated[tx.CashAccountID.String]; isEstimated {
					if isCashOnly(tx.Type) {
						continue
					}
					p.Invested -= tx.CashDelta
					continue
				}
				// Every cash effect, trades included, lands on a cash
				// account; a transaction without one moves no cash.
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

// isCashOnly reports whether a transaction type moves cash and nothing
// else: no security changes hands.
func isCashOnly(typ string) bool {
	switch typ {
	case "INTEREST", "DEPOSIT_CASH", "WITHDRAW_CASH", "ACCOUNT_FEES", "TAX_REFUND":
		return true
	}
	return false
}
