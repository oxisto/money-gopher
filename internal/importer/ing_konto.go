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
// dated, amount-bearing row instead of exactly one — except rows for
// securities trades and dividends/coupons, which are deliberately skipped;
// see the ISIN check below.
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

// ingKontoTaxLabel reports whether desc is a withholding-tax line
// (Kapitalertragsteuer, Kirchensteuer, Solidaritätszuschlag) that belongs
// folded into a same-date interest row's Taxes, rather than standing alone.
func ingKontoTaxLabel(desc string) bool {
	return strings.Contains(desc, "Kapitalertragsteuer") ||
		strings.Contains(desc, "Kirchensteuer") ||
		strings.Contains(desc, "Solidaritätszuschlag")
}

// ingKontoRawRow is one matched, not-yet-filtered ledger line.
type ingKontoRawRow struct {
	dateStr string
	desc    string
	amount  int64
}

// ingParseKontoauszug parses every dated cash movement in an ING account
// statement. Amounts are classified by sign (DEPOSIT_CASH / WITHDRAW_CASH),
// refined to INTEREST or ACCOUNT_FEES when the description says so.
// Kapitalertragsteuer/Kirchensteuer/Solidaritätszuschlag lines dated the
// same day as an interest row are withholding on that interest, not a
// separate cash event — they're folded into the interest row's Taxes (with
// CashDelta reduced to the net amount actually received) instead of being
// emitted as their own misleading WITHDRAW_CASH rows. Statements have been
// seen with the tax lines both before and after the interest row, so this
// groups by date rather than assuming an order.
func ingParseKontoauszug(text string) ([]StagedTransaction, error) {
	iban := ""
	if m := ingKontoIBAN.FindStringSubmatch(text); m != nil {
		iban = strings.ReplaceAll(strings.TrimSpace(m[1]), " ", "")
	}

	lines := strings.Split(text, "\n")

	// First pass: collect every real ledger row (date, description, amount),
	// filtering out non-postings but not yet classifying or grouping them.
	var rows []ingKontoRawRow
	for i, line := range lines {
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

		// A securities-related cash movement (a trade's settlement, or a
		// dividend/coupon payout) always has a continuation line naming an
		// ISIN — "WP-ABRECHNUNG ... ISIN ..." for a trade, "Zins/Dividende
		// ISIN ..." for income. These carry no unit/price/fee breakdown
		// here (a Kontoauszug only shows the net cash effect), and the
		// user already imports a dedicated, complete confirmation PDF for
		// every one of them — so importing this row too would double-count
		// the same cash movement rather than add information.
		if i+1 < len(lines) && strings.Contains(lines[i+1], "ISIN") {
			continue
		}

		if _, err := ingDate(m[1]); err != nil {
			continue // header/footer noise that happens to look date-shaped
		}
		amount, err := ingMinor(m[3])
		if err != nil || amount == 0 {
			continue
		}

		rows = append(rows, ingKontoRawRow{dateStr: m[1], desc: desc, amount: amount})
	}

	// Second pass: for each date, sum any withholding-tax rows and fold them
	// into the first interest row on that date, wherever in the row order
	// each happens to fall. A tax-labeled row with no interest row on the
	// same date (never seen in practice, but not impossible) has nothing to
	// fold into — leave it to be classified normally rather than silently
	// dropping real money.
	hasInterest := map[string]bool{}
	for _, r := range rows {
		if strings.Contains(r.desc, "Zins") {
			hasInterest[r.dateStr] = true
		}
	}

	taxByDate := map[string]int64{}
	consumed := make([]bool, len(rows))
	for i, r := range rows {
		if ingKontoTaxLabel(r.desc) && hasInterest[r.dateStr] {
			taxByDate[r.dateStr] += -r.amount // stored negative; Taxes is positive.
			consumed[i] = true
		}
	}

	var out []StagedTransaction
	foldedDate := map[string]bool{} // only the first interest row per date gets the fold.
	for i, r := range rows {
		if consumed[i] {
			continue
		}

		t, err := ingDate(r.dateStr)
		if err != nil {
			continue
		}

		txType := "DEPOSIT_CASH"
		if r.amount < 0 {
			txType = "WITHDRAW_CASH"
		}
		switch {
		case strings.Contains(r.desc, "Zins"):
			txType = "INTEREST"
		case strings.Contains(r.desc, "Entgelt"), strings.Contains(r.desc, "Gebühr"):
			txType = "ACCOUNT_FEES"
		}

		amount := r.amount
		var taxes int64
		if txType == "INTEREST" && !foldedDate[r.dateStr] {
			taxes = taxByDate[r.dateStr]
			amount -= taxes
			foldedDate[r.dateStr] = true
		}

		out = append(out, StagedTransaction{
			Type:           txType,
			Time:           t,
			CashDelta:      amount,
			Taxes:          taxes,
			Currency:       "EUR",
			SettlementIBAN: iban,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no transactions found in Kontoauszug")
	}
	return out, nil
}
