// Package finance contains the snapshot and performance calculations of
// moneyd: positions held at a point in time, their purchase and market
// values, and time-weighted returns over a period.
//
// All monetary values are integer minor units (e.g. euro cents), matching
// the persistence layer. The calculations assume a single currency per
// portfolio; currency conversion is a future problem.
package finance

import (
	"math"
	"sort"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// QuoteFunc looks up the market price (in minor units) of a security at the
// given time, typically from the latest quote at or before that time. It
// reports ok == false when no quote is known, in which case calculations
// fall back to the purchase price.
type QuoteFunc func(securityID string, at time.Time) (price int64, ok bool)

// NoQuotes is a QuoteFunc that knows no quotes at all; every position is
// valued at its purchase price.
func NoQuotes(string, time.Time) (int64, bool) { return 0, false }

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
	// known, the average purchase price otherwise.
	MarketPrice int64
	// MarketValue is MarketPrice times Units.
	MarketValue int64
}

// PurchasePrice is the average purchase price of one held share.
func (p *Position) PurchasePrice() int64 {
	if p.Units == 0 {
		return 0
	}
	return int64(math.Round(float64(p.PurchaseValue) / p.Units))
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
	ppu   int64 // purchase price per unit in minor units
}

// value is the cost basis of the remaining shares of the lot.
func (l *lot) value() int64 {
	return int64(math.Round(float64(l.ppu) * l.units))
}

// book tracks the FIFO lots of one security.
type book struct {
	currency string
	lots     []*lot
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

// apply adjusts the book for one transaction.
func (b *book) apply(tx *persistence.Transaction) {
	switch tx.Type {
	case "BUY", "DELIVERY_INBOUND":
		b.lots = append(b.lots, &lot{units: tx.Units, ppu: tx.Price.Int64})
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

		if price, ok := quote(b.securityID, at); ok {
			pos.MarketPrice = price
		} else {
			pos.MarketPrice = pos.PurchasePrice()
		}
		pos.MarketValue = int64(math.Round(float64(pos.MarketPrice) * pos.Units))

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

// marketValue prices the books at the given time.
func marketValue(bs []*securityBook, at time.Time, quote QuoteFunc) (v int64) {
	for _, b := range bs {
		units := b.units()
		if units <= 0 {
			continue
		}

		price, ok := quote(b.securityID, at)
		if !ok {
			// Fall back to the average purchase price, like SnapshotAt.
			price = int64(math.Round(float64(b.purchaseValue()) / units))
		}
		v += int64(math.Round(float64(price) * units))
	}
	return
}
