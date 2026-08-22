package importer

import (
	"fmt"
	"regexp"
	"strings"
)

// Kontoauszug: ING-DiBa cash/savings account statements (Girokonto or
// Extra-Konto). Unlike the brokerage confirmations above — one document, one
// transaction — a Kontoauszug covers a whole statement period and lists many
// cash movements, so ingParseKontoauszug returns one StagedTransaction per
// dated, amount-bearing row instead of exactly one.
var (
	ingKontoIBAN = regexp.MustCompile(`(?m)^\s*IBAN\s+([A-Z]{2}\d{2}[\d ]+)\s*$`)

	// A transaction row: a date, a description, and a trailing German-format
	// amount (optionally negative), separated by 2+ spaces (the column gap
	// pdftotext -layout preserves). Every entry also has a second "Valuta"
	// line repeating the date with no trailing amount, plus optional wrapped
	// purpose-text lines with no date at all — neither matches this pattern,
	// so they're skipped without special-casing.
	ingKontoRow = regexp.MustCompile(`^\s*(\d{2}\.\d{2}\.\d{4})\s+(.+?)\s{2,}(-?[\d.]+,\d{2})\s*$`)
)

// ingIsKontoauszug reports whether text looks like an ING account statement,
// as opposed to a brokerage confirmation.
func ingIsKontoauszug(text string) bool {
	return strings.Contains(text, "Kontoauszug")
}

// ingParseKontoauszug parses every dated cash movement in an ING account
// statement. Amounts are classified by sign (DEPOSIT_CASH / WITHDRAW_CASH),
// refined to INTEREST or ACCOUNT_FEES when the description says so.
func ingParseKontoauszug(text string) ([]StagedTransaction, error) {
	iban := ""
	if m := ingKontoIBAN.FindStringSubmatch(text); m != nil {
		iban = strings.ReplaceAll(strings.TrimSpace(m[1]), " ", "")
	}

	var out []StagedTransaction
	for _, line := range strings.Split(text, "\n") {
		m := ingKontoRow.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		desc := m[2]
		// The year-end interest breakdown ("Zeitraum / Zins p.a. / Ertrag")
		// restates the same total across rate tiers and date ranges — not
		// real ledger postings, and it happens to fit the row pattern above
		// (a leading date followed by a trailing amount).
		if strings.Contains(desc, "%") {
			continue
		}

		t, err := ingDate(m[1])
		if err != nil {
			continue // header/footer noise that happens to look date-shaped
		}
		amount, err := ingMinor(m[3])
		if err != nil {
			continue
		}
		if amount == 0 {
			continue
		}

		txType := "DEPOSIT_CASH"
		if amount < 0 {
			txType = "WITHDRAW_CASH"
		}
		switch {
		case strings.Contains(desc, "Zins"):
			txType = "INTEREST"
		case strings.Contains(desc, "Entgelt"), strings.Contains(desc, "Gebühr"):
			txType = "ACCOUNT_FEES"
		}

		out = append(out, StagedTransaction{
			Type:           txType,
			Time:           t,
			CashDelta:      amount,
			Currency:       "EUR",
			SettlementIBAN: iban,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no transactions found in Kontoauszug")
	}
	return out, nil
}
