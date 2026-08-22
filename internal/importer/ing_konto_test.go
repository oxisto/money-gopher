package importer

import (
	"os"
	"testing"
)

// TestING_ParseKontoauszug parses a sanitized ING account statement fixture
// (invented data, same structure as a real one) and checks that every row is
// found, classified, and signed correctly, and that nothing from the
// year-end interest breakdown table leaks through as a spurious transaction.
func TestING_ParseKontoauszug(t *testing.T) {
	text, err := os.ReadFile("testdata/ing_konto/kontoauszug.txt")
	if err != nil {
		t.Skipf("testdata/ing_konto not found: %v", err)
	}

	p := ING{}
	if !p.Matches(string(text)) {
		t.Fatal("Matches() = false")
	}

	txs, err := p.Parse(string(text))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Gehalt (deposit), Miete (withdrawal), Kontoführungsentgelt (fee),
	// Stromversorger (withdrawal), Zinsertrag (interest), and three tax
	// withholding lines (withdrawals) — eight real rows. The "%"-marked
	// breakdown table row and the Abschlussbetrag restatement must not
	// produce a ninth.
	const wantRows = 8
	if len(txs) != wantRows {
		t.Fatalf("len(txs) = %d, want %d", len(txs), wantRows)
	}

	for _, tx := range txs {
		if tx.Time.IsZero() {
			t.Errorf("%+v: Time is zero", tx)
		}
		if tx.Currency != "EUR" {
			t.Errorf("%+v: Currency = %q, want EUR", tx, tx.Currency)
		}
		if tx.SettlementIBAN != "DE02123456780000000042" {
			t.Errorf("%+v: SettlementIBAN = %q", tx, tx.SettlementIBAN)
		}
		if tx.CashDelta == 0 {
			t.Errorf("%+v: CashDelta is zero", tx)
		}
	}

	byType := map[string]int{}
	var net int64
	for _, tx := range txs {
		byType[tx.Type]++
		net += tx.CashDelta

		switch tx.Type {
		case "DEPOSIT_CASH", "INTEREST":
			if tx.CashDelta <= 0 {
				t.Errorf("%s: CashDelta = %d, want positive", tx.Type, tx.CashDelta)
			}
		case "WITHDRAW_CASH", "ACCOUNT_FEES":
			if tx.CashDelta >= 0 {
				t.Errorf("%s: CashDelta = %d, want negative", tx.Type, tx.CashDelta)
			}
		default:
			t.Errorf("unexpected type %q", tx.Type)
		}
	}

	wantByType := map[string]int{
		"DEPOSIT_CASH":  1,
		"WITHDRAW_CASH": 5, // Miete, Stromversorger, and the 3 tax lines
		"ACCOUNT_FEES":  1,
		"INTEREST":      1,
	}
	for typ, want := range wantByType {
		if byType[typ] != want {
			t.Errorf("count[%s] = %d, want %d", typ, byType[typ], want)
		}
	}

	// The statement's own opening/closing balance ties out the net exactly:
	// 5,000.00 -> 6,239.12, i.e. +1,239.12.
	const wantNet = 123912
	if net != wantNet {
		t.Errorf("net = %d, want %d", net, wantNet)
	}
}

// TestING_KontoauszugDoesNotMatchTrade guards against the two fingerprints
// overlapping: a brokerage confirmation must never be routed to the
// multi-row Kontoauszug parser, and vice versa.
func TestING_KontoauszugDoesNotMatchTrade(t *testing.T) {
	text, err := os.ReadFile("testdata/ing/buy.txt")
	if err != nil {
		t.Skipf("testdata/ing not found: %v", err)
	}
	if ingIsKontoauszug(string(text)) {
		t.Error("ingIsKontoauszug() = true for a brokerage confirmation")
	}
}
