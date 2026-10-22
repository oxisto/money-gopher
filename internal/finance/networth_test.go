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

// TestNetWorthAt_EstimatedAccount: the Giro is declared at 1,500 and not
// tracked. Its deposits and withdrawals do not exist for net worth; a
// purchase paid from it is capital brought in, a dividend landing on it
// money taken out — the declared balance never moves.
func TestNetWorthAt_EstimatedAccount(t *testing.T) {
	giro := func(tx *persistence.Transaction) *persistence.Transaction {
		tx.CashAccountID = sql.NullString{String: "giro", Valid: true}
		return tx
	}
	txs := []*persistence.Transaction{
		giro(tx("DEPOSIT_CASH", t0, "", 0, 0, 200000)), // ignored: the Giro is an estimate
		giro(tx("BUY", t1, "msci", 10, 10000, -100000)),
		giro(tx("DIVIDEND", t2, "msci", 0, 0, 500)),
	}
	quote := quotes(map[string][]struct {
		at    time.Time
		price int64
	}{"msci": {{t1, 10000}}})

	got := NetWorthAt(txs, map[string]int64{"giro": 150000}, []time.Time{t0, t1, t2}, quote)
	want := []NetWorthPoint{
		{Time: t0, NetWorth: 150000, Invested: 0, Estimated: 150000},
		{Time: t1, NetWorth: 250000, Invested: 100000, Estimated: 150000},
		{Time: t2, NetWorth: 250000, Invested: 99500, Estimated: 150000},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("point %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// TestNetWorthAt_LoanIsNegative: a loan is an account like any other; its
// balance simply runs below zero and counts against net worth.
func TestNetWorthAt_LoanIsNegative(t *testing.T) {
	loan := func(tx *persistence.Transaction) *persistence.Transaction {
		tx.CashAccountID = sql.NullString{String: "loan", Valid: true}
		return tx
	}
	txs := []*persistence.Transaction{
		loan(tx("WITHDRAW_CASH", t0, "", 0, 0, -19000000)), // disbursed
		loan(tx("INTEREST", t1, "", 0, 0, -47500)),         // charged
		loan(tx("DEPOSIT_CASH", t2, "", 0, 0, 47500)),      // paid
	}
	got := NetWorthAt(txs, nil, []time.Time{t0, t1, t2}, NoQuotes)
	for i, want := range []int64{-19000000, -19047500, -19000000} {
		if got[i].NetWorth != want {
			t.Errorf("point %d net worth = %d, want %d", i, got[i].NetWorth, want)
		}
	}
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
	got := NetWorthAt(txs, nil, []time.Time{before, t0, t1, t2, t3}, quote)

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
