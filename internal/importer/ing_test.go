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

			// BUY: cash must leave the account (negative delta).
			// Price may be nil for bonds (Kurs is a percentage, not a per-unit price).
			if tx.Type == "BUY" {
				if tx.CashDelta >= 0 {
					t.Errorf("BUY CashDelta = %d, want negative", tx.CashDelta)
				}
				if tx.Price != nil && *tx.Price <= 0 {
					t.Errorf("BUY Price = %v, want positive or nil", tx.Price)
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
