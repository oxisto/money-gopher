package finance

import (
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

var (
	t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	t2 = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	t3 = time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
)

// tx builds a security transaction. price and cashDelta are minor units.
func tx(typ string, at time.Time, security string, units float64, price int64, cashDelta int64) *persistence.Transaction {
	return &persistence.Transaction{
		Type:       typ,
		Time:       at,
		SecurityID: sql.NullString{String: security, Valid: security != ""},
		Units:      units,
		Price:      sql.NullInt64{Int64: price, Valid: price != 0},
		CashDelta:  cashDelta,
		Currency:   "EUR",
	}
}

// quotes returns a QuoteFunc serving fixed prices per security, each valid
// from its given time on.
func quotes(m map[string][]struct {
	at    time.Time
	price int64
}) QuoteFunc {
	return func(securityID string, at time.Time) (int64, bool) {
		var (
			price int64
			ok    bool
		)
		for _, q := range m[securityID] {
			if q.at.After(at) {
				break
			}
			price, ok = q.price, true
		}
		return price, ok
	}
}

func TestSnapshotAt_FIFO(t *testing.T) {
	txs := []*persistence.Transaction{
		tx("BUY", t1, "msci", 10, 10000, -100000),
		tx("BUY", t2, "msci", 10, 12000, -120000),
		tx("SELL", t3, "msci", 5, 13000, 65000),
	}

	snap := SnapshotAt(txs, t3, NoQuotes)

	if len(snap.Positions) != 1 {
		t.Fatalf("positions = %d, want 1", len(snap.Positions))
	}

	pos := snap.Positions[0]
	if pos.Units != 15 {
		t.Errorf("units = %v, want 15", pos.Units)
	}
	// FIFO: the 5 sold shares come out of the first lot (bought at 100.00),
	// leaving 5×100.00 + 10×120.00 = 1700.00.
	if pos.PurchaseValue != 170000 {
		t.Errorf("purchase value = %d, want 170000", pos.PurchaseValue)
	}
	if want := int64(math.Round(170000.0 / 15)); pos.PurchasePrice() != want {
		t.Errorf("purchase price = %d, want %d", pos.PurchasePrice(), want)
	}
	// Without quotes, the market value equals the purchase value (modulo
	// per-unit rounding).
	if pos.MarketValue != int64(math.Round(float64(pos.PurchasePrice())*15)) {
		t.Errorf("market value = %d, want purchase-price based", pos.MarketValue)
	}
	if snap.Currency != "EUR" {
		t.Errorf("currency = %q, want EUR", snap.Currency)
	}
	if !snap.FirstTransactionTime.Equal(t1) {
		t.Errorf("first transaction time = %v, want %v", snap.FirstTransactionTime, t1)
	}
}

func TestSnapshotAt_WithQuotes(t *testing.T) {
	txs := []*persistence.Transaction{
		tx("BUY", t1, "msci", 10, 10000, -100000),
	}
	q := quotes(map[string][]struct {
		at    time.Time
		price int64
	}{
		"msci": {{t2, 13000}},
	})

	snap := SnapshotAt(txs, t3, q)

	pos := snap.Positions[0]
	if pos.MarketPrice != 13000 {
		t.Errorf("market price = %d, want 13000", pos.MarketPrice)
	}
	if pos.MarketValue != 130000 {
		t.Errorf("market value = %d, want 130000", pos.MarketValue)
	}
	if pos.ProfitOrLoss() != 30000 {
		t.Errorf("profit = %d, want 30000", pos.ProfitOrLoss())
	}
	if g := pos.Gains(); math.Abs(g-0.3) > 1e-9 {
		t.Errorf("gains = %v, want 0.3", g)
	}
	if snap.TotalMarketValue != 130000 || snap.TotalPurchaseValue != 100000 {
		t.Errorf("totals = %d/%d, want 130000/100000",
			snap.TotalMarketValue, snap.TotalPurchaseValue)
	}
	if g := snap.TotalGains(); math.Abs(g-0.3) > 1e-9 {
		t.Errorf("total gains = %v, want 0.3", g)
	}

	// A snapshot before the quote falls back to the purchase price.
	if snap := SnapshotAt(txs, t1, q); snap.Positions[0].MarketPrice != 10000 {
		t.Errorf("market price before quote = %d, want 10000", snap.Positions[0].MarketPrice)
	}
}

func TestSnapshotAt_TimeTravel(t *testing.T) {
	txs := []*persistence.Transaction{
		tx("BUY", t1, "msci", 10, 10000, -100000),
		tx("SELL", t3, "msci", 10, 11000, 110000),
	}

	// At t2 the shares are still held ...
	if snap := SnapshotAt(txs, t2, NoQuotes); len(snap.Positions) != 1 {
		t.Errorf("positions at t2 = %d, want 1", len(snap.Positions))
	}
	// ... at t3 the position is sold out and disappears.
	if snap := SnapshotAt(txs, t3, NoQuotes); len(snap.Positions) != 0 {
		t.Errorf("positions at t3 = %d, want 0", len(snap.Positions))
	}
	// ... before the first transaction there is nothing.
	if snap := SnapshotAt(txs, t0, NoQuotes); len(snap.Positions) != 0 {
		t.Errorf("positions at t0 = %d, want 0", len(snap.Positions))
	}
}

func TestSnapshotAt_IgnoresCashEvents(t *testing.T) {
	txs := []*persistence.Transaction{
		tx("BUY", t1, "msci", 10, 10000, -100000),
		tx("DIVIDEND", t2, "msci", 0, 0, 5000),
		tx("DEPOSIT_CASH", t2, "", 0, 0, 100000),
	}

	snap := SnapshotAt(txs, t3, NoQuotes)
	if len(snap.Positions) != 1 || snap.Positions[0].Units != 10 {
		t.Fatalf("dividend/cash events must not change positions: %+v", snap.Positions)
	}
}

func TestSnapshotAt_Empty(t *testing.T) {
	snap := SnapshotAt(nil, t1, NoQuotes)
	if len(snap.Positions) != 0 || snap.TotalMarketValue != 0 || snap.TotalGains() != 0 {
		t.Errorf("empty snapshot not empty: %+v", snap)
	}
	if !snap.FirstTransactionTime.IsZero() {
		t.Errorf("first transaction time = %v, want zero", snap.FirstTransactionTime)
	}
}

func TestTimeWeightedReturn(t *testing.T) {
	q := quotes(map[string][]struct {
		at    time.Time
		price int64
	}{
		"msci": {{t1, 10000}, {t2, 12000}, {t3, 12600}},
	})

	t.Run("single buy", func(t *testing.T) {
		txs := []*persistence.Transaction{
			tx("BUY", t1, "msci", 10, 10000, -100000),
		}
		// 100.00 → 126.00 over the full period: +26 %.
		if r := TimeWeightedReturn(txs, t0, t3, q); math.Abs(r-0.26) > 1e-9 {
			t.Errorf("return = %v, want 0.26", r)
		}
	})

	t.Run("flow between periods", func(t *testing.T) {
		txs := []*persistence.Transaction{
			tx("BUY", t1, "msci", 10, 10000, -100000),
			tx("BUY", t2, "msci", 10, 12000, -120000),
		}
		// +20 % until the second buy, +5 % after: 1.2 × 1.05 − 1 = 0.26.
		// The second buy itself must not count as gain.
		if r := TimeWeightedReturn(txs, t0, t3, q); math.Abs(r-0.26) > 1e-9 {
			t.Errorf("return = %v, want 0.26", r)
		}
	})

	t.Run("dividend adds to return", func(t *testing.T) {
		txs := []*persistence.Transaction{
			tx("BUY", t1, "msci", 10, 10000, -100000),
			tx("DIVIDEND", t2, "msci", 0, 0, 12000),
		}
		// Price return is +26 %; the 120.00 dividend at t2 is a distribution
		// out of a 1200.00 portfolio, adding 10 % for the last sub-period:
		// 1.2 × (1260/1080) − 1 = 0.4.
		if r := TimeWeightedReturn(txs, t0, t3, q); math.Abs(r-0.4) > 1e-9 {
			t.Errorf("return = %v, want 0.4", r)
		}
	})

	t.Run("no transactions", func(t *testing.T) {
		if r := TimeWeightedReturn(nil, t0, t3, q); r != 0 {
			t.Errorf("return = %v, want 0", r)
		}
	})

	t.Run("fees reduce return", func(t *testing.T) {
		withFees := []*persistence.Transaction{
			// Same trade, but the cash leg includes 10.00 fees.
			tx("BUY", t1, "msci", 10, 10000, -101000),
		}
		r := TimeWeightedReturn(withFees, t0, t3, q)
		if r >= 0.26 {
			t.Errorf("return with fees = %v, want < 0.26", r)
		}
	})
}
