package finance

import (
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// decimalCents converts minor units (e.g. 10000 = "100.00") to a Decimal, so
// test tables can keep using plain integers while Transaction.Price and
// QuoteFunc are exact major-unit Decimals. Exact integer division — no
// float64 involved.
func decimalCents(minorUnits int64) persistence.Decimal {
	return persistence.NewDecimalFromInt64(minorUnits).DivInt(100)
}

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
		Price:      decimalCents(price),
		CashDelta:  cashDelta,
		Currency:   "EUR",
	}
}

// quotes returns a QuoteFunc serving fixed prices per security, each valid
// from its given time on. price is minor units, like tx's.
func quotes(m map[string][]struct {
	at    time.Time
	price int64
}) QuoteFunc {
	return func(securityID, _ string, at time.Time) (persistence.Decimal, bool) {
		var (
			price persistence.Decimal
			ok    bool
		)
		for _, q := range m[securityID] {
			if q.at.After(at) {
				break
			}
			price, ok = decimalCents(q.price), true
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
	// 170000 / 15 / 100 = 113.333... repeating.
	if want := "113.33"; pos.PurchasePrice().StringPrec(2) != want {
		t.Errorf("purchase price = %s, want %s", pos.PurchasePrice().StringPrec(2), want)
	}
	// Without quotes, the market value equals the purchase value exactly:
	// MarketPrice is PurchasePrice, and multiplying it back by units*100
	// exactly undoes the division that computed it.
	if pos.MarketValue != pos.PurchaseValue {
		t.Errorf("market value = %d, want purchase value %d", pos.MarketValue, pos.PurchaseValue)
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
	if pos.MarketPrice.StringPrec(2) != "130.00" {
		t.Errorf("market price = %s, want 130.00", pos.MarketPrice.StringPrec(2))
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
	if snap := SnapshotAt(txs, t1, q); snap.Positions[0].MarketPrice.StringPrec(2) != "100.00" {
		t.Errorf("market price before quote = %s, want 100.00", snap.Positions[0].MarketPrice.StringPrec(2))
	}
}

// TestSnapshotAt_ForeignCurrencyPurchase books a US Treasury bought with
// euros: the price is in USD and the bank's rate says what the euros
// bought, so the cost basis is in euros — and so is the quote the
// snapshot asks for.
func TestSnapshotAt_ForeignCurrencyPurchase(t *testing.T) {
	buy := tx("BUY", t1, "tsy", 300, 9811, -2806961) // 300 × 98.11 USD at 1.0518
	buy.PriceCurrency = sql.NullString{String: "USD", Valid: true}
	buy.FxRate, _ = persistence.NewDecimal("1.0518")
	txs := []*persistence.Transaction{buy}

	var askedFor string
	quote := func(_, currency string, _ time.Time) (persistence.Decimal, bool) {
		askedFor = currency
		return persistence.Decimal{}, false
	}
	snap := SnapshotAt(txs, t2, quote)

	pos := snap.Positions[0]
	if pos.PurchaseValue != 2798346 {
		t.Errorf("PurchaseValue = %d, want 2798346 (29433 USD at 1.0518)", pos.PurchaseValue)
	}
	if got := pos.PurchasePrice().StringPrec(2); got != "93.28" {
		t.Errorf("PurchasePrice = %s EUR, want 93.28", got)
	}
	if askedFor != "EUR" {
		t.Errorf("quote asked for %q, want the settlement currency EUR", askedFor)
	}

	// Selling 20 at par realizes the gain against the euro cost basis.
	sell := tx("SELL", t3, "tsy", 20, 10000, 188350)
	sell.ID = "redemption"
	sell.PriceCurrency = sql.NullString{String: "USD", Valid: true}
	sell.FxRate, _ = persistence.NewDecimal("1.0554")
	gain, ok := RealizedGainOf(append(txs, sell), sell.ID)
	if !ok {
		t.Fatal("RealizedGainOf() not ok")
	}
	// 1883.50 EUR proceeds − 20 × 93.28 EUR basis.
	if gain != 188350-186556 {
		t.Errorf("realized gain = %d, want %d", gain, 188350-186556)
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

func TestRealizedGains(t *testing.T) {
	txs := []*persistence.Transaction{
		tx("BUY", t0, "msci", 10, 10000, -100000), // 10 @ 100.00
		tx("SELL", t1, "msci", 4, 12000, 47600),   // sell 4 @ 120.00, net 476.00
		tx("DIVIDEND", t1, "msci", 0, 0, 5000),
		{Type: "INTEREST", Time: t2, CashDelta: 700, Currency: "EUR"},
		tx("SELL", t3, "msci", 2, 11000, 21800), // outside the [t0,t3) window below
	}

	gains := RealizedGains(txs, t0, t3)

	// Cost basis of the 4 sold shares (FIFO out of the 10 bought at 100.00
	// each) is 400.00; net proceeds were 476.00, so the realized gain is
	// 76.00 — the t3 sale must not appear at all, since [t0,t3) excludes it.
	var dividends, interest, sell int64
	for _, g := range gains {
		switch g.Type {
		case "DIVIDEND":
			dividends += g.Amount
		case "INTEREST":
			interest += g.Amount
		case "SELL":
			sell += g.Amount
			if g.Time != t1 {
				t.Errorf("unexpected SELL gain at %v (window is [t0,t3))", g.Time)
			}
		}
	}
	if dividends != 5000 {
		t.Errorf("dividends = %d, want 5000", dividends)
	}
	if interest != 700 {
		t.Errorf("interest = %d, want 700", interest)
	}
	if sell != 7600 {
		t.Errorf("sell gain = %d, want 7600", sell)
	}
}

func TestRealizedGainOf(t *testing.T) {
	firstSale := &persistence.Transaction{
		ID: "sale1", Type: "SELL", Time: t1,
		SecurityID: sql.NullString{String: "msci", Valid: true},
		Units:      4, CashDelta: 47600, Currency: "EUR",
	}
	secondSale := &persistence.Transaction{
		ID: "sale2", Type: "SELL", Time: t1, // same day, must not be confused with sale1
		SecurityID: sql.NullString{String: "msci", Valid: true},
		Units:      2, CashDelta: 21800, Currency: "EUR",
	}
	txs := []*persistence.Transaction{
		tx("BUY", t0, "msci", 10, 10000, -100000), // 10 @ 100.00
		firstSale,
		secondSale,
	}

	if amount, ok := RealizedGainOf(txs, "sale1"); !ok || amount != 7600 {
		t.Errorf("sale1 gain = %d, %v; want 7600, true", amount, ok)
	}
	// Cost basis continues FIFO from where sale1 left off: 2 more shares
	// @ 100.00 = 20000; proceeds 21800 -> gain 1800.
	if amount, ok := RealizedGainOf(txs, "sale2"); !ok || amount != 1800 {
		t.Errorf("sale2 gain = %d, %v; want 1800, true", amount, ok)
	}

	if _, ok := RealizedGainOf(txs, "nonexistent"); ok {
		t.Error("RealizedGainOf found a transaction that isn't in txs")
	}

	buy := tx("BUY", t0, "msci", 10, 10000, -100000)
	buy.ID = "buy1"
	if _, ok := RealizedGainOf([]*persistence.Transaction{buy}, "buy1"); ok {
		t.Error("RealizedGainOf reported ok for a BUY, not a SELL")
	}
}
