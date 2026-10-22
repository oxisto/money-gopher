package finance

import (
	"sort"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// Returns is what a portfolio earned over a period, in money. Four parts
// add up to the total: what completed sales realised, what the holdings
// paid out, what the trades and payouts cost, and how the open positions
// moved in value. Realized and Income are gross — the costs stand on
// their own — so nothing is counted twice.
type Returns struct {
	From, To time.Time
	Currency string
	// Realized is the FIFO profit or loss of sales completed in the
	// period, before the sales' own fees and taxes.
	Realized int64
	// Income is dividends, coupons and rent received in the period, before
	// withholding.
	Income int64
	// Costs is every fee and tax paid in the period: on purchases, sales
	// and payouts.
	Costs int64
	// ValueChange is how the open positions moved in value over the
	// period, net of what was bought and sold: the unrealised part.
	ValueChange int64
	// Total is Realized + Income - Costs + ValueChange: the period's
	// absolute return, i.e. the value at the end minus the value at the
	// start minus the money put in plus the money taken out.
	Total int64
}

// ReturnsOver computes the Returns of the portfolio transactions txs over
// (from, to], valued with quote. A transaction exactly at from belongs to
// the period before, as TimeWeightedReturn has it.
func ReturnsOver(txs []*persistence.Transaction, from, to time.Time, quote QuoteFunc) Returns {
	sorted := make([]*persistence.Transaction, len(txs))
	copy(sorted, txs)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	r := Returns{From: from, To: to}
	inWindow := func(t time.Time) bool { return t.After(from) && !t.After(to) }
	byID := map[string]*book{}
	var flows int64 // money that left the portfolio for cash, net; purchases negative
	for _, tx := range sorted {
		if r.Currency == "" && tx.Currency != "" {
			r.Currency = tx.Currency
		}
		switch tx.Type {
		case "BUY", "DELIVERY_INBOUND":
			if !tx.SecurityID.Valid {
				continue
			}
			b := byID[tx.SecurityID.String]
			if b == nil {
				b = &book{currency: tx.Currency}
				byID[tx.SecurityID.String] = b
			}
			b.apply(tx)
			if inWindow(tx.Time) {
				if tx.Type == "BUY" {
					r.Costs += tx.Fees + tx.Taxes
					flows += tx.CashDelta
				} else {
					flows -= tx.Price.MulFloat(tx.Units).MulInt(100).RoundToInt64()
				}
			}
		case "SELL", "DELIVERY_OUTBOUND":
			if !tx.SecurityID.Valid {
				continue
			}
			var basis int64
			if b := byID[tx.SecurityID.String]; b != nil {
				basis = b.costBasis(tx.Units)
				b.apply(tx)
			}
			if inWindow(tx.Time) {
				if tx.Type == "SELL" {
					r.Realized += tx.CashDelta + tx.Fees + tx.Taxes - basis
					r.Costs += tx.Fees + tx.Taxes
					flows += tx.CashDelta
				} else {
					flows += tx.Price.MulFloat(tx.Units).MulInt(100).RoundToInt64()
				}
			}
		case "DIVIDEND":
			if inWindow(tx.Time) {
				r.Income += tx.CashDelta + tx.Taxes
				r.Costs += tx.Fees + tx.Taxes
				flows += tx.CashDelta
			}
		}
	}

	start := marketValue(books(txs, from, false), from, quote)
	end := marketValue(books(txs, to, false), to, quote)
	r.Total = end - start + flows
	r.ValueChange = r.Total - (r.Realized + r.Income - r.Costs)
	return r
}
