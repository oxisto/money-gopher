package finance

import (
	"database/sql"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

func TestReturnsOver(t *testing.T) {
	at := func(day int) time.Time { return time.Date(2026, 1, day, 12, 0, 0, 0, time.UTC) }
	sec := sql.NullString{String: "sec", Valid: true}
	dec := func(s string) persistence.Decimal {
		d, err := persistence.NewDecimal(s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	txs := []*persistence.Transaction{
		// 10 shares at 100, fee 10: cash −1,010.
		{ID: "buy", Type: "BUY", Time: at(2), SecurityID: sec, Units: 10, Price: dec("100"), Fees: 1000, CashDelta: -101000, Currency: "EUR"},
		// A dividend of 20 gross, 5 withheld: cash +15.
		{ID: "div", Type: "DIVIDEND", Time: at(10), SecurityID: sec, Taxes: 500, CashDelta: 1500, Currency: "EUR"},
		// 5 shares sold at 120, fee 2: cash +598; FIFO basis 500.
		{ID: "sell", Type: "SELL", Time: at(20), SecurityID: sec, Units: 5, Price: dec("120"), Fees: 200, CashDelta: 59800, Currency: "EUR"},
	}
	quote := func(string, string, time.Time) (persistence.Decimal, bool) { return dec("110"), true }

	r := ReturnsOver(txs, at(1), at(31), quote)
	want := Returns{From: at(1), To: at(31), Currency: "EUR", Realized: 10000, Income: 2000, Costs: 1700, ValueChange: 5000, Total: 15300}
	if r != want {
		t.Errorf("ReturnsOver = %+v\n          want %+v", r, want)
	}

	// A later period holds no flows: only the open 5 shares moving from
	// 110 to 110 — nothing — and the start value equals the end value.
	later := ReturnsOver(txs, at(31), time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC), quote)
	if later.Total != 0 || later.ValueChange != 0 || later.Realized != 0 || later.Income != 0 || later.Costs != 0 {
		t.Errorf("a flat later period = %+v, want all zero", later)
	}
}
