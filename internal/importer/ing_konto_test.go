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
	// Stromversorger (withdrawal), a standalone Kapitalertragsteuer
	// correction (tax refund, see below), and Zinsertrag (interest) — six
	// real rows. The three withholding-tax lines dated the same day as
	// Zinsertrag (Kapitalertragsteuer, Solidaritätszuschlag, Kirchensteuer)
	// fold into the interest row's Taxes rather than standing alone, so
	// they don't add three more. The "%"-marked breakdown table row and the
	// Abschlussbetrag restatement must not produce a row either. Nor must
	// the Wertpapierkauf (-100.00) and Zins/Dividende WP (+24.50) rows:
	// both have an ISIN-bearing continuation line, so a dedicated
	// trade/dividend confirmation already covers them elsewhere — counting
	// them here too would double the cash impact of the same real-world
	// event. Nor the second Wertpapierkauf (-5,432.10), whose ISIN
	// continuation lands on a simulated next page — past a page-footer
	// watermark, a blank gap, and a repeated page-header/column-heading
	// block — rather than the very next line. The Sparplan's
	// Lastschrift-Einzug (+350.00) names an ISIN in its continuation too,
	// but it is the money pulled in to fund the plan, not the trade, and
	// must be kept as a deposit.
	const wantRows = 7
	if len(txs) != wantRows {
		t.Fatalf("len(txs) = %d, want %d", len(txs), wantRows)
	}
	// The note is the statement's wording: the amount line's Buchungsart
	// and counterparty, then the Verwendungszweck from the Valuta line and
	// its wrapped continuation, without the SEPA Referenz line.
	if got := txs[0].Note; got != "Gutschrift Max Mustermann · Gehalt Januar" {
		t.Errorf("note = %q, want %q", got, "Gutschrift Max Mustermann · Gehalt Januar")
	}
	if got := txs[2].Note; got != "Ueberweisung Max Mustermann · Miete Januar" {
		t.Errorf("note = %q, want %q", got, "Ueberweisung Max Mustermann · Miete Januar")
	}
	for _, tx := range txs {
		if tx.CashDelta == -10000 || tx.CashDelta == 2450 || tx.CashDelta == -543210 {
			t.Errorf("%+v: a securities-related row (ISIN continuation line) leaked through", tx)
		}
	}

	// A Kapitalertragsteuer line with no same-date interest row to fold
	// into (a "Nachtr. Verlustverrechnung" — a retroactive capital-gains
	// loss offset) is a standalone tax adjustment: TAX_REFUND, negative
	// here since this one nets to an additional small charge rather than a
	// refund — the type allows either sign for exactly this reason.
	var foundTaxRefund bool
	for _, tx := range txs {
		if tx.Type == "TAX_REFUND" {
			foundTaxRefund = true
			if tx.CashDelta != -2 {
				t.Errorf("TAX_REFUND CashDelta = %d, want -2", tx.CashDelta)
			}
		}
	}
	if !foundTaxRefund {
		t.Error("no TAX_REFUND transaction found")
	}

	// Zinsertrag (10.25) minus Kapitalertragsteuer (2.56) + Solidaritätszuschlag
	// (0.14) + Kirchensteuer (0.55) = 7.00 net, with 3.25 folded into Taxes.
	var foundInterest bool
	for _, tx := range txs {
		if tx.Type != "INTEREST" {
			continue
		}
		foundInterest = true
		if tx.CashDelta != 700 {
			t.Errorf("INTEREST CashDelta = %d, want 700 (net of withheld tax)", tx.CashDelta)
		}
		if tx.Taxes != 325 {
			t.Errorf("INTEREST Taxes = %d, want 325", tx.Taxes)
		}
	}
	if !foundInterest {
		t.Error("no INTEREST transaction found")
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
		case "TAX_REFUND":
			// Either sign is valid; checked explicitly above instead.
		default:
			t.Errorf("unexpected type %q", tx.Type)
		}
	}

	var foundSparplan bool
	for _, tx := range txs {
		if tx.CashDelta == 35000 {
			foundSparplan = true
			if tx.Type != "DEPOSIT_CASH" {
				t.Errorf("Sparplan Lastschrift-Einzug: Type = %q, want DEPOSIT_CASH", tx.Type)
			}
		}
	}
	if !foundSparplan {
		t.Error("the Sparplan Lastschrift-Einzug (+350.00) was skipped as if it were the trade")
	}

	wantByType := map[string]int{
		"DEPOSIT_CASH":  2, // Gehalt, Sparplan Lastschrift-Einzug
		"WITHDRAW_CASH": 2, // Miete, Stromversorger
		"ACCOUNT_FEES":  1,
		"INTEREST":      1,
		"TAX_REFUND":    1,
	}
	for typ, want := range wantByType {
		if byType[typ] != want {
			t.Errorf("count[%s] = %d, want %d", typ, byType[typ], want)
		}
	}

	// The seven tracked rows' own net: 2500.00 (Gehalt) + 350.00 (Sparplan
	// debit) - 800.00 (Miete) - 11.90 (Kontoführungsentgelt) - 455.98
	// (Stromversorger) - 0.02 (TAX_REFUND) + 7.00 (net interest) = 1589.10.
	// This is smaller than
	// the statement's own Alter Saldo -> Neuer Saldo delta (5,000.00 ->
	// 731.50, i.e. -4,268.50): the difference is exactly the two excluded
	// Wertpapierkauf rows and the excluded dividend (-100.00 - 5,432.10 +
	// 24.50 = -5,507.60), deliberately not tracked here since a dedicated
	// document covers each of them.
	const wantNet = 158910
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
