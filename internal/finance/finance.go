// Package finance contains the snapshot and performance calculations of
// moneyd: positions held at a point in time, their purchase and market
// values, and time-weighted returns over a period.
//
// All monetary values are integer minor units (e.g. euro cents), matching
// the persistence layer, and are in the currency the transactions settle
// in. A purchase priced in another currency (a US Treasury bought in USD
// with euros) is booked at its price converted with the transaction's own
// FX rate, and quotes are requested in the settlement currency, so a
// portfolio's figures all share one currency even when its holdings don't.
package finance

import (
	"math"
	"sort"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// QuoteFunc looks up the market price (an exact per-unit Decimal) of a
// security at the given time, expressed in the given currency, typically
// from the latest quote at or before that time. It reports ok == false when
// no quote is known (or it cannot be converted), in which case calculations
// fall back to the last trade's price.
type QuoteFunc func(securityID, currency string, at time.Time) (price persistence.Decimal, ok bool)

// NoQuotes is a QuoteFunc that knows no quotes at all; every position is
// valued at its purchase price.
func NoQuotes(string, string, time.Time) (persistence.Decimal, bool) {
	return persistence.Decimal{}, false
}

// Position is the holding of one security at the snapshot time.
type Position struct {
	// SecurityID identifies the security held.
	SecurityID string
	// Currency is the ISO 4217 code of all monetary values of the position.
	Currency string
	// Units is the number of shares held.
	Units float64
	// PurchaseValue is the FIFO cost basis of the held shares, without fees
	// and taxes.
	PurchaseValue int64
	// MarketPrice is the price of one share: the latest quote if one is
	// known, otherwise the last trade's price, otherwise the average
	// purchase price. An exact Decimal, unlike
	// the rest of this struct's minor-units Money fields — see
	// persistence.Decimal for why per-unit prices need more precision.
	MarketPrice persistence.Decimal
	// MarketValue is MarketPrice times Units.
	MarketValue int64
}

// PurchasePrice is the average purchase price of one held share.
func (p *Position) PurchasePrice() persistence.Decimal {
	if p.Units == 0 {
		return persistence.Decimal{}
	}
	// PurchaseValue is minor units (cents); Decimal prices are major units.
	return persistence.NewDecimalFromInt64(p.PurchaseValue).DivFloat(p.Units).DivInt(100)
}

// ProfitOrLoss is the absolute unrealized gain (or loss, if negative).
func (p *Position) ProfitOrLoss() int64 {
	return p.MarketValue - p.PurchaseValue
}

// Gains is the relative unrealized gain, e.g. 0.05 for +5 %.
func (p *Position) Gains() float64 {
	if p.PurchaseValue == 0 {
		return 0
	}
	return float64(p.ProfitOrLoss()) / float64(p.PurchaseValue)
}

// Snapshot is the state of a portfolio at a point in time.
type Snapshot struct {
	// Time is the time the snapshot was taken for.
	Time time.Time
	// FirstTransactionTime is the time of the earliest transaction; zero if
	// the portfolio has none.
	FirstTransactionTime time.Time
	// Positions are the currently held securities, sorted by security ID.
	// Sold-out positions are not included.
	Positions []*Position
	// Currency is the ISO 4217 code of the totals, taken from the
	// transactions (single-currency portfolios are assumed).
	Currency string
	// TotalPurchaseValue is the sum of all position purchase values.
	TotalPurchaseValue int64
	// TotalMarketValue is the sum of all position market values.
	TotalMarketValue int64
}

// TotalProfitOrLoss is the absolute unrealized gain of all positions.
func (s *Snapshot) TotalProfitOrLoss() int64 {
	return s.TotalMarketValue - s.TotalPurchaseValue
}

// TotalGains is the relative unrealized gain of all positions.
func (s *Snapshot) TotalGains() float64 {
	if s.TotalPurchaseValue == 0 {
		return 0
	}
	return float64(s.TotalProfitOrLoss()) / float64(s.TotalPurchaseValue)
}

// lot is one FIFO entry: shares bought together at one price.
type lot struct {
	units float64
	ppu   persistence.Decimal // purchase price per unit, major units
}

// value is the cost basis of the remaining shares of the lot, in minor
// units (cents).
func (l *lot) value() int64 {
	return l.ppu.MulFloat(l.units).MulInt(100).RoundToInt64()
}

// book tracks the FIFO lots of one security.
type book struct {
	currency string
	lots     []*lot
	// lastPrice is the price of the most recent trade, in the settlement
	// currency: the best price known when no quote is. A bond that left
	// the quote feed is still worth what it last changed hands at.
	lastPrice persistence.Decimal
}

func (b *book) units() (u float64) {
	for _, l := range b.lots {
		u += l.units
	}
	return
}

func (b *book) purchaseValue() (v int64) {
	for _, l := range b.lots {
		v += l.value()
	}
	return
}

// fallbackPrice is the price of one unit when no quote is known: the last
// trade's, or the average purchase price before any trade carried one (a
// delivery from another broker, say).
func (b *book) fallbackPrice() persistence.Decimal {
	if b.lastPrice.Valid && b.lastPrice.Sign() > 0 {
		return b.lastPrice
	}
	units := b.units()
	if units == 0 {
		return persistence.Decimal{}
	}
	// purchaseValue() is minor units (cents); Decimal prices are major units.
	return persistence.NewDecimalFromInt64(b.purchaseValue()).DivFloat(units).DivInt(100)
}

// apply adjusts the book for one transaction.
func (b *book) apply(tx *persistence.Transaction) {
	if price := purchasePrice(tx); price.Valid && price.Sign() > 0 {
		switch tx.Type {
		case "BUY", "SELL", "DELIVERY_INBOUND", "DELIVERY_OUTBOUND":
			b.lastPrice = price
		}
	}

	switch tx.Type {
	case "BUY", "DELIVERY_INBOUND":
		b.lots = append(b.lots, &lot{units: tx.Units, ppu: purchasePrice(tx)})
	case "SELL", "DELIVERY_OUTBOUND":
		// Sold shares leave the book first-in-first-out.
		sold := tx.Units
		for _, l := range b.lots {
			if sold <= 0 {
				break
			}
			n := math.Min(sold, l.units)
			l.units -= n
			sold -= n
		}
	}
}

// purchasePrice is the price of one unit in the currency the transaction
// settles in. A foreign-currency purchase carries its price in the
// security's own currency and the rate the bank converted at: how many
// units of that currency one unit of the settlement currency bought. So
// dividing by it is the conversion, and it uses the very rate the money
// actually moved at.
func purchasePrice(tx *persistence.Transaction) persistence.Decimal {
	if tx.FxRate.Valid && tx.FxRate.Sign() > 0 {
		return tx.Price.Div(tx.FxRate)
	}
	return tx.Price
}

// SnapshotAt computes the portfolio snapshot at the given time from the
// portfolio's transactions (any order) and a quote lookup.
func SnapshotAt(txs []*persistence.Transaction, at time.Time, quote QuoteFunc) *Snapshot {
	snap := &Snapshot{Time: at}

	for _, tx := range txs {
		if !snap.FirstTransactionTime.IsZero() && !tx.Time.Before(snap.FirstTransactionTime) {
			continue
		}
		snap.FirstTransactionTime = tx.Time
	}

	for _, b := range books(txs, at, false) {
		if b.units() <= 0 {
			continue
		}

		pos := &Position{
			SecurityID:    b.securityID,
			Currency:      b.currency,
			Units:         b.units(),
			PurchaseValue: b.purchaseValue(),
		}

		if price, ok := quote(b.securityID, b.currency, at); ok {
			pos.MarketPrice = price
		} else {
			pos.MarketPrice = b.fallbackPrice()
		}
		pos.MarketValue = pos.MarketPrice.MulFloat(pos.Units).MulInt(100).RoundToInt64()

		snap.Positions = append(snap.Positions, pos)
		snap.TotalPurchaseValue += pos.PurchaseValue
		snap.TotalMarketValue += pos.MarketValue
		if snap.Currency == "" {
			snap.Currency = pos.Currency
		}
	}

	sort.Slice(snap.Positions, func(i, j int) bool {
		return snap.Positions[i].SecurityID < snap.Positions[j].SecurityID
	})

	return snap
}

// securityBook pairs a book with its security for iteration.
type securityBook struct {
	securityID string
	*book
}

// books replays all security transactions up to the cutoff and returns the
// FIFO books per security, sorted by security ID. With exclusive == true,
// transactions exactly at the cutoff are left out.
func books(txs []*persistence.Transaction, cutoff time.Time, exclusive bool) []*securityBook {
	m := make(map[string]*book)

	// Transactions must be replayed in time order for FIFO to be correct.
	sorted := make([]*persistence.Transaction, 0, len(txs))
	for _, tx := range txs {
		if tx.Time.After(cutoff) || (exclusive && tx.Time.Equal(cutoff)) {
			continue
		}
		if !tx.SecurityID.Valid {
			continue
		}
		sorted = append(sorted, tx)
	}
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	for _, tx := range sorted {
		b := m[tx.SecurityID.String]
		if b == nil {
			b = &book{currency: tx.Currency}
			m[tx.SecurityID.String] = b
		}
		b.apply(tx)
	}

	out := make([]*securityBook, 0, len(m))
	for id, b := range m {
		out = append(out, &securityBook{securityID: id, book: b})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].securityID < out[j].securityID })

	return out
}

// costBasis returns the FIFO cost basis of the next `units` shares in the
// book (in minor units), without draining them — the caller applies the
// actual sale separately.
func (b *book) costBasis(units float64) (v int64) {
	remaining := units
	for _, l := range b.lots {
		if remaining <= 0 {
			break
		}
		n := math.Min(remaining, l.units)
		v += l.ppu.MulFloat(n).MulInt(100).RoundToInt64()
		remaining -= n
	}
	return v
}

// RealizedGain is one cash-realized event within a period: a dividend or
// interest payment (already net of withheld tax), or the realized profit or
// loss of a completed sale (net proceeds minus the FIFO cost basis of the
// shares sold). Unlike Position.ProfitOrLoss, this is money that has
// actually moved, not a mark-to-market valuation.
type RealizedGain struct {
	TransactionID string
	Time          time.Time
	Type          string // "DIVIDEND", "INTEREST", or "SELL".
	SecurityID    string // Set for DIVIDEND and SELL; empty for INTEREST.
	// PortfolioID/CashAccountID identify where the underlying transaction can
	// be found (and edited) in the UI — a trade/dividend has a portfolio, a
	// cash account statement's interest doesn't.
	PortfolioID   string
	CashAccountID string
	Currency      string
	Amount        int64
}

// RealizedGains replays txs (any order, from any number of portfolios/cash
// accounts) to find every dividend, interest payment, and completed sale
// whose time falls within [from, to). Sales are costed FIFO against the full
// transaction history, not just the ones inside the window, so a sale of
// shares bought before `from` still gets the correct cost basis.
func RealizedGains(txs []*persistence.Transaction, from, to time.Time) []RealizedGain {
	sorted := make([]*persistence.Transaction, len(txs))
	copy(sorted, txs)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	books := map[string]*book{}
	inWindow := func(t time.Time) bool { return !t.Before(from) && t.Before(to) }

	var out []RealizedGain
	for _, tx := range sorted {
		switch tx.Type {
		case "BUY", "DELIVERY_INBOUND":
			if !tx.SecurityID.Valid {
				continue
			}
			b := books[tx.SecurityID.String]
			if b == nil {
				b = &book{currency: tx.Currency}
				books[tx.SecurityID.String] = b
			}
			b.apply(tx)

		case "SELL", "DELIVERY_OUTBOUND":
			if !tx.SecurityID.Valid {
				continue
			}
			b := books[tx.SecurityID.String]
			var basis int64
			if b != nil {
				basis = b.costBasis(tx.Units)
				b.apply(tx)
			}
			if tx.Type == "SELL" && inWindow(tx.Time) {
				out = append(out, RealizedGain{
					TransactionID: tx.ID, Time: tx.Time, Type: "SELL", SecurityID: tx.SecurityID.String,
					PortfolioID: tx.PortfolioID.String, CashAccountID: tx.CashAccountID.String,
					Currency: tx.Currency, Amount: tx.CashDelta - basis,
				})
			}

		case "DIVIDEND":
			if inWindow(tx.Time) {
				out = append(out, RealizedGain{
					TransactionID: tx.ID, Time: tx.Time, Type: "DIVIDEND", SecurityID: tx.SecurityID.String,
					PortfolioID: tx.PortfolioID.String, CashAccountID: tx.CashAccountID.String,
					Currency: tx.Currency, Amount: tx.CashDelta,
				})
			}

		case "INTEREST":
			if inWindow(tx.Time) {
				out = append(out, RealizedGain{
					TransactionID: tx.ID, Time: tx.Time, Type: "INTEREST",
					PortfolioID: tx.PortfolioID.String, CashAccountID: tx.CashAccountID.String,
					Currency: tx.Currency, Amount: tx.CashDelta,
				})
			}
		}
	}

	return out
}

// RealizedGainOf returns the realized profit or loss of one specific SELL
// transaction: its net proceeds minus the FIFO cost basis of the shares it
// sold, found by replaying the full history of that security (in any order,
// across every portfolio) up to and including it. Unlike RealizedGains'
// time-window matching, this pins to the transaction's own ID, so it stays
// unambiguous even when several sales share the same day. ok is false for
// anything that isn't a SELL, or isn't found in txs at all.
func RealizedGainOf(txs []*persistence.Transaction, transactionID string) (amount int64, ok bool) {
	var target *persistence.Transaction
	for _, tx := range txs {
		if tx.ID == transactionID {
			target = tx
			break
		}
	}
	if target == nil || target.Type != "SELL" || !target.SecurityID.Valid {
		return 0, false
	}

	sorted := make([]*persistence.Transaction, 0, len(txs))
	for _, tx := range txs {
		if tx.SecurityID.Valid && tx.SecurityID.String == target.SecurityID.String {
			sorted = append(sorted, tx)
		}
	}
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	b := &book{currency: target.Currency}
	for _, tx := range sorted {
		switch tx.Type {
		case "BUY", "DELIVERY_INBOUND":
			b.apply(tx)
		case "SELL", "DELIVERY_OUTBOUND":
			basis := b.costBasis(tx.Units)
			b.apply(tx)
			if tx.ID == transactionID {
				return tx.CashDelta - basis, true
			}
		}
	}
	return 0, false
}

// marketValue prices the books at the given time.
func marketValue(bs []*securityBook, at time.Time, quote QuoteFunc) (v int64) {
	for _, b := range bs {
		units := b.units()
		if units <= 0 {
			continue
		}

		price, ok := quote(b.securityID, b.currency, at)
		if !ok {
			price = b.fallbackPrice()
		}
		v += price.MulFloat(units).MulInt(100).RoundToInt64()
	}
	return
}
