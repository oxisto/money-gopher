package finance

import (
	"database/sql"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// onAccount settles tx against the cash account "giro".
func onAccount(tx *persistence.Transaction) *persistence.Transaction {
	tx.CashAccountID = sql.NullString{String: "giro", Valid: true}
	return tx
}

func TestNetWorthAt(t *testing.T) {
	txs := []*persistence.Transaction{
		onAccount(tx("DEPOSIT_CASH", t0, "", 0, 0, 200000)),
		// €1,000 + €10 fees out of the account into 10 shares.
		onAccount(tx("BUY", t1, "msci", 10, 10000, -101000)),
		onAccount(tx("INTEREST", t2, "", 0, 0, 500)),
		// Brought in from another broker, no cash.
		tx("DELIVERY_INBOUND", t2, "sap", 2, 15000, 0),
		onAccount(tx("WITHDRAW_CASH", t3, "", 0, 0, -50000)),
	}
	quote := quotes(map[string][]struct {
		at    time.Time
		price int64
	}{
		"msci": {{t1, 10000}, {t2, 12000}},
		"sap":  {{t2, 15000}},
	})

	before := t0.Add(-time.Hour)
	got := NetWorthAt(txs, []time.Time{before, t0, t1, t2, t3}, quote)

	want := []struct{ netWorth, invested int64 }{
		{0, 0},
		{200000, 200000},
		// 99,000 cash + 10 × 100.00
		{199000, 200000},
		// 99,500 cash + 10 × 120.00 + 2 × 150.00
		{99500 + 120000 + 30000, 200000 + 30000},
		{49500 + 120000 + 30000, 150000 + 30000},
	}
	for i, w := range want {
		if got[i].NetWorth != w.netWorth || got[i].Invested != w.invested {
			t.Errorf("point %d = (netWorth %d, invested %d), want (%d, %d)",
				i, got[i].NetWorth, got[i].Invested, w.netWorth, w.invested)
		}
	}
}
