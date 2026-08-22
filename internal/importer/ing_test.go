package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestING_Matches verifies the fingerprint against a representative sample.
func TestING_Matches(t *testing.T) {
	p := ING{}

	entries, err := os.ReadDir("testdata/ing")
	if err != nil {
		t.Skipf("testdata/ing not found: %v", err)
	}

	for _, e := range entries {
		text, err := os.ReadFile(filepath.Join("testdata/ing", e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if !p.Matches(string(text)) {
			t.Errorf("%s: Matches() = false", e.Name())
		}
	}

	if p.Matches("Some random text without an ING header") {
		t.Error("Matches() = true for a non-ING document")
	}
}

// TestING_ParseAll runs every extracted ING text through the parser and checks
// that the result is a single, internally consistent transaction.
func TestING_ParseAll(t *testing.T) {
	entries, err := os.ReadDir("testdata/ing")
	if err != nil {
		t.Skipf("testdata/ing not found: %v", err)
	}

	for _, e := range entries {
		e := e
		t.Run(strings.TrimSuffix(e.Name(), ".txt"), func(t *testing.T) {
			text, err := os.ReadFile(filepath.Join("testdata/ing", e.Name()))
			if err != nil {
				t.Fatal(err)
			}

			txs, err := ING{}.Parse(string(text))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if len(txs) != 1 {
				t.Fatalf("len = %d, want 1", len(txs))
			}
			tx := txs[0]

			// Every transaction must have a valid type.
			validTypes := map[string]bool{
				"BUY": true, "SELL": true, "DIVIDEND": true,
			}
			if !validTypes[tx.Type] {
				t.Errorf("Type = %q, want BUY/SELL/DIVIDEND", tx.Type)
			}

			// ISIN: 12-char alphanumeric.
			if len(tx.ISIN) != 12 {
				t.Errorf("ISIN = %q, want 12 chars", tx.ISIN)
			}

			if tx.SecurityHint == "" {
				t.Error("SecurityHint is empty")
			}

			if tx.Time.IsZero() {
				t.Error("Time is zero")
			}

			// BUY: cash must leave the account (negative delta). Bonds report
			// Units as face-value/100 "lots" and Price as the Kurs percentage
			// (e.g. 100.00 for a bond bought at par) — units × price then
			// reproduces Kurswert, same as it does for equities.
			if tx.Type == "BUY" {
				if tx.CashDelta >= 0 {
					t.Errorf("BUY CashDelta = %d, want negative", tx.CashDelta)
				}
				if !tx.Price.Valid || tx.Price.Sign() <= 0 {
					t.Errorf("BUY Price = %v, want positive", tx.Price)
				}
				if tx.Units <= 0 {
					t.Errorf("BUY Units = %v, want positive", tx.Units)
				}
			}

			// SELL: cash must arrive (positive delta).
			if tx.Type == "SELL" && tx.CashDelta <= 0 {
				t.Errorf("SELL CashDelta = %d, want positive", tx.CashDelta)
			}

			// DIVIDEND: cash must arrive, currency must be set.
			if tx.Type == "DIVIDEND" {
				if tx.CashDelta <= 0 {
					t.Errorf("DIVIDEND CashDelta = %d, want positive", tx.CashDelta)
				}
				if tx.Currency == "" {
					t.Error("DIVIDEND Currency is empty")
				}
			}

			t.Logf("OK  %s  %s  %s  units=%.4f  cashDelta=%d  taxes=%d  iban=%s",
				tx.Type, tx.ISIN, tx.Time.Format("2006-01-02"),
				tx.Units, tx.CashDelta, tx.Taxes, tx.SettlementIBAN)
		})
	}
}

// TestING_ParseTrade_BondAtPar verifies the bond price/units convention: a
// bond's "Kurs" is a percentage of face value (not a per-unit price), so
// Units becomes face-value/100 "lots" and Price becomes the percentage
// itself — units × price then reproduces Kurswert exactly, the same
// invariant equities already satisfy.
func TestING_ParseTrade_BondAtPar(t *testing.T) {
	text, err := os.ReadFile("testdata/ing/buy_bond_par.txt")
	if err != nil {
		t.Fatal(err)
	}

	tx, err := ingParseTrade(string(text))
	if err != nil {
		t.Fatalf("ingParseTrade() error = %v", err)
	}

	if tx.Units != 50 {
		t.Errorf("Units = %v, want 50 (5000 nominal / 100)", tx.Units)
	}
	if got := tx.Price.StringPrec(2); got != "100.00" {
		t.Errorf("Price = %s, want 100.00 (bought at par)", got)
	}
	if tx.CashDelta != -500000 {
		t.Errorf("CashDelta = %d, want -500000", tx.CashDelta)
	}

	// units × price must reproduce Kurswert (500000, in cents) exactly.
	kurswert := tx.Price.MulFloat(tx.Units).MulInt(100).RoundToInt64()
	if kurswert != 500000 {
		t.Errorf("Units*Price = %d, want 500000 (Kurswert)", kurswert)
	}
}

// TestING_ParseTrade_ForeignTransactionTax verifies that a foreign financial
// transaction tax (e.g. France's "Franz. Transaktionssteuer") lands in Taxes,
// not Fees (Provision) — and that Kurswert + Fees + Taxes reconciles exactly
// against the Endbetrag-derived CashDelta.
func TestING_ParseTrade_ForeignTransactionTax(t *testing.T) {
	text, err := os.ReadFile("testdata/ing/buy_french_tax.txt")
	if err != nil {
		t.Fatal(err)
	}

	tx, err := ingParseTrade(string(text))
	if err != nil {
		t.Fatalf("ingParseTrade() error = %v", err)
	}

	if tx.Taxes != 630 {
		t.Errorf("Taxes = %d, want 630 (the 0.30%% French FTT)", tx.Taxes)
	}
	if tx.Fees != 990 {
		t.Errorf("Fees = %d, want 990 (Provision only, not the FTT)", tx.Fees)
	}

	// Kurswert (210000) + Fees (990) + Taxes (630) must equal the Endbetrag.
	const kurswert = 210000
	if got, want := kurswert+tx.Fees+tx.Taxes, int64(211620); got != want {
		t.Errorf("Kurswert+Fees+Taxes = %d, want %d (Endbetrag)", got, want)
	}
	if tx.CashDelta != -211620 {
		t.Errorf("CashDelta = %d, want -211620", tx.CashDelta)
	}
}
