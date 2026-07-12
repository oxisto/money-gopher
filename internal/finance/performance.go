package finance

import (
	"math"
	"sort"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// TimeWeightedReturn computes the time-weighted return of the portfolio's
// security positions over (from, to] as a fraction, e.g. 0.05 for +5 %.
//
// The return is chained over sub-periods delimited by external flows: buys
// and sells (valued at their cash leg, so fees and taxes count as cost),
// deliveries (valued at their price), and dividends (treated as a
// distribution out of the portfolio, so they add to the return). Cash
// accounts are outside the portfolio by design, so pure cash events do not
// appear here at all.
//
// Without quote history, positions are valued at their purchase price and
// the return over a flat period is 0 — honest, if boring, until quotes
// arrive.
func TimeWeightedReturn(txs []*persistence.Transaction, from, to time.Time, quote QuoteFunc) float64 {
	// Collect the net external flow per distinct transaction time within
	// (from, to].
	type flow struct {
		time  time.Time
		value int64
	}
	var flows []flow
	byTime := make(map[time.Time]int64)

	for _, tx := range txs {
		if !tx.Time.After(from) || tx.Time.After(to) {
			continue
		}

		var f int64
		switch tx.Type {
		case "BUY", "SELL", "DIVIDEND":
			// The cash leg carries the full external value including fees
			// and taxes; its sign is the cash account's view, so the
			// portfolio's flow is the negation.
			f = -tx.CashDelta
		case "DELIVERY_INBOUND":
			f = int64(math.Round(float64(tx.Price.Int64) * tx.Units))
		case "DELIVERY_OUTBOUND":
			f = -int64(math.Round(float64(tx.Price.Int64) * tx.Units))
		default:
			continue
		}

		if _, ok := byTime[tx.Time]; !ok {
			flows = append(flows, flow{time: tx.Time})
		}
		byTime[tx.Time] += f
	}

	sort.Slice(flows, func(i, j int) bool { return flows[i].time.Before(flows[j].time) })
	for i := range flows {
		flows[i].value = byTime[flows[i].time]
	}

	// valueBefore prices the positions held strictly before t at time t;
	// valueAt includes transactions at t itself.
	valueBefore := func(t time.Time) int64 { return marketValue(books(txs, t, true), t, quote) }
	valueAt := func(t time.Time) int64 { return marketValue(books(txs, t, false), t, quote) }

	r := 1.0
	prev := float64(valueAt(from)) // start value, after any flows at exactly `from`

	for _, f := range flows {
		v := float64(valueBefore(f.time))
		if prev > 0 {
			r *= v / prev
		}
		prev = v + float64(f.value)
	}

	if end := float64(valueAt(to)); prev > 0 {
		r *= end / prev
	}

	return r - 1
}
