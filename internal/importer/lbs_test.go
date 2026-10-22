package importer

import (
	"os"
	"testing"
)

// TestLBS_ParseJahresschluss parses an invented LBS year-end bundle in the
// real layout: a Bausparvertrag statement and a Vorfinanzierungskredit
// statement under one contract number, plus the explanatory pages, the
// Wohnungsbauprämie form and the Steuerbescheinigung that carry no ledger
// rows. The column an amount sits in gives its sign, the section gives
// the account hint.
func TestLBS_ParseJahresschluss(t *testing.T) {
	text, err := os.ReadFile("testdata/lbs/jahresschluss.txt")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}

	p := LBS{}
	if !p.Matches(string(text)) {
		t.Fatal("Matches() = false")
	}
	if (ING{}).Matches(string(text)) {
		t.Fatal("the ING parser claims an LBS document")
	}

	txs, err := p.Parse(string(text))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	type want struct {
		typ   string
		date  string
		delta int64
		taxes int64
		hint  string
	}
	wants := []want{
		// Bausparvertrag: three Sparraten in, two fees, interest net of the
		// Abgeltungsteuer folded into it. The Saldovortrag and closing
		// balance carry no date and are not rows.
		{"DEPOSIT_CASH", "2025-02-03", 20000, 0, "Bausparvertrag 1234567890"},
		{"DEPOSIT_CASH", "2025-03-03", 20000, 0, "Bausparvertrag 1234567890"},
		{"DEPOSIT_CASH", "2025-04-01", 20000, 0, "Bausparvertrag 1234567890"},
		{"ACCOUNT_FEES", "2025-07-01", -3000, 0, "Bausparvertrag 1234567890"},
		{"ACCOUNT_FEES", "2025-07-01", -500, 0, "Bausparvertrag 1234567890"},
		{"INTEREST", "2025-12-31", 5, 2, "Bausparvertrag 1234567890"},
		// Vorfinanzierungskredit: commitment interest charged, one payment
		// in, the disbursement out, more interest charged.
		{"INTEREST", "2025-10-31", -12500, 0, "Vorfinanzierungskredit 1234567890"},
		{"DEPOSIT_CASH", "2025-12-02", 12500, 0, "Vorfinanzierungskredit 1234567890"},
		{"WITHDRAW_CASH", "2025-12-23", -5000000, 0, "Vorfinanzierungskredit 1234567890"},
		{"INTEREST", "2025-12-23", -11250, 0, "Vorfinanzierungskredit 1234567890"},
		{"INTEREST", "2025-12-31", -3296, 0, "Vorfinanzierungskredit 1234567890"},
	}
	if len(txs) != len(wants) {
		for _, tx := range txs {
			t.Logf("%s %s %d tax=%d %s", tx.Time.Format("2006-01-02"), tx.Type, tx.CashDelta, tx.Taxes, tx.AccountHint)
		}
		t.Fatalf("len(txs) = %d, want %d", len(txs), len(wants))
	}
	for i, w := range wants {
		tx := txs[i]
		if tx.Type != w.typ || tx.Time.Format("2006-01-02") != w.date || tx.CashDelta != w.delta || tx.Taxes != w.taxes || tx.AccountHint != w.hint {
			t.Errorf("row %d = %s %s %d tax=%d %q, want %s %s %d tax=%d %q", i,
				tx.Type, tx.Time.Format("2006-01-02"), tx.CashDelta, tx.Taxes, tx.AccountHint,
				w.typ, w.date, w.delta, w.taxes, w.hint)
		}
		if tx.Currency != "EUR" || tx.SettlementIBAN != "" {
			t.Errorf("row %d: currency %q, IBAN %q; want EUR and no IBAN (the one printed is the LBS's own)", i, tx.Currency, tx.SettlementIBAN)
		}
	}
}
