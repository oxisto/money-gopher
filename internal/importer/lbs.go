package importer

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// LBS parses the year-end bundle of a Landesbausparkasse (LBS Süd as seen;
// the layout is the Sparkassen-Finanzgruppe's). One PDF carries one
// Kontoauszug per contract part — the Bausparvertrag and, once a loan is
// taken against it, the Vorfinanzierungskredit or Bauspardarlehen — plus
// explanatory pages, a Wohnungsbauprämie form and a Steuerbescheinigung.
//
// Every statement is a ledger: booking date, value date, description and
// one amount that sits either in the Belastungen or in the Gutschriften
// column. The column decides the sign; the section decides the account.
// The IBAN printed is the LBS's own collection account, the same on every
// part, so it is useless for matching: each row carries an AccountHint
// ("Bausparvertrag 1234567890") instead and the review picks an account
// per hint.
type LBS struct{}

func (LBS) Name() string { return "LBS" }

var (
	// "Kontoauszug 2025   Seite Nr. 1 von 1   Bausparvertrag": the start of
	// one statement, naming the contract part in the right-hand column.
	lbsSection = regexp.MustCompile(`^\s*Kontoauszug\s+(\d{4})\s+Seite Nr\.\s*\d+\s+von\s+\d+\s+(\S.*?)\s*$`)
	// "Nr. 7 038 222 803" on the line after the section header.
	lbsContract = regexp.MustCompile(`^\s*Nr\.\s+((?:\d[\d ]*)\d)\s*$`)
	// A ledger row: booking date, value date, description, then the amount
	// in one of the two columns. Sidebar text (Tarif, IBAN, the
	// Finanzierungskostennachweis) may follow after a wider gap.
	lbsRow = regexp.MustCompile(`^\s*(\d{2}\.\d{2}\.\d{4})\s+(\d{2}\.\d{2}\.\d{4})\s+(.+?)\s{2,}([\d.]+,\d{2})(?:\s{2,}|\s*$)`)
	// The closing balance line ends a section.
	lbsClosing = regexp.MustCompile(`^\s*(Neuer Saldo am|Ihr .* am) \d{2}\.\d{2}\.\d{4}`)
)

// lbsTaxLabel: the LBS calls the withholding "Abgeltungsteuer Zinsen"; the
// ING statement's Kapitalertragsteuer/Kirchensteuer/Solidaritätszuschlag
// are kept for a Steuerbescheinigung-style breakdown.
func lbsTaxLabel(desc string) bool {
	return strings.Contains(desc, "Abgeltungsteuer") || ingKontoTaxLabel(desc)
}

func (LBS) Matches(text string) bool {
	if !strings.Contains(text, "Landesbausparkasse") && !strings.Contains(text, "LBS Süd") {
		return false
	}
	for _, line := range strings.Split(text, "\n") {
		if lbsSection.MatchString(line) {
			return true
		}
	}
	return false
}

// lbsRawRow is one ledger line before classification.
type lbsRawRow struct {
	dateStr string
	desc    string
	amount  int64 // signed: negative in the Belastungen column
	hint    string
}

func (LBS) Parse(text string) ([]StagedTransaction, error) {
	lines := strings.Split(text, "\n")

	var (
		rows      []lbsRawRow
		part      string // "Bausparvertrag", "Vorfinanzierungskredit", ...
		contract  string
		splitCol  = -1 // rune column between the two amount columns; -1 before the header
		inSection bool
	)
	for i, line := range lines {
		if m := lbsSection.FindStringSubmatch(line); m != nil {
			part, contract, splitCol, inSection = m[2], "", -1, true
			// The contract number follows on one of the next lines.
			for j := i + 1; j < len(lines) && j <= i+3; j++ {
				if c := lbsContract.FindStringSubmatch(lines[j]); c != nil {
					contract = strings.ReplaceAll(c[1], " ", "")
					break
				}
			}
			continue
		}
		if !inSection {
			continue
		}
		if idx := strings.Index(line, "Belastungen"); idx >= 0 && strings.Contains(line, "Gutschriften") {
			// Amounts are right-aligned a few columns past each heading's
			// end, so the split sits halfway between the two headings.
			debitEnd := utf8.RuneCountInString(line[:idx]) + len("Belastungen")
			creditEnd := utf8.RuneCountInString(line[:strings.Index(line, "Gutschriften")]) + len("Gutschriften")
			splitCol = (debitEnd + creditEnd) / 2
			continue
		}
		if lbsClosing.MatchString(line) {
			inSection = false
			continue
		}
		if splitCol < 0 {
			continue
		}
		loc := lbsRow.FindStringSubmatchIndex(line)
		if loc == nil {
			continue
		}
		amount, err := ingMinor(line[loc[8]:loc[9]])
		if err != nil || amount == 0 {
			continue
		}
		// The amount's right edge says which column it sits in.
		if end := utf8.RuneCountInString(line[:loc[9]]); end <= splitCol {
			amount = -amount
		}
		hint := part
		if contract != "" {
			hint = part + " " + contract
		}
		rows = append(rows, lbsRawRow{
			dateStr: line[loc[2]:loc[3]],
			desc:    strings.TrimSpace(line[loc[6]:loc[7]]),
			amount:  amount,
			hint:    hint,
		})
	}

	// Withholding tax on the Bausparvertrag's interest is its own row on the
	// same date; fold it into that interest row like the ING statement does.
	type key struct{ hint, date string }
	hasInterest := map[key]bool{}
	for _, r := range rows {
		if strings.Contains(r.desc, "Zins") && !lbsTaxLabel(r.desc) {
			hasInterest[key{r.hint, r.dateStr}] = true
		}
	}
	taxes := map[key]int64{}
	consumed := make([]bool, len(rows))
	for i, r := range rows {
		k := key{r.hint, r.dateStr}
		if lbsTaxLabel(r.desc) && hasInterest[k] {
			taxes[k] += -r.amount
			consumed[i] = true
		}
	}

	var out []StagedTransaction
	folded := map[key]bool{}
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
		case lbsTaxLabel(r.desc):
			txType = "TAX_REFUND"
		case strings.Contains(r.desc, "Zins"):
			// Charged on the loan (negative), credited on the savings.
			txType = "INTEREST"
		case strings.Contains(strings.ToLower(r.desc), "entgelt"), strings.Contains(strings.ToLower(r.desc), "gebühr"):
			// "Jahresentgelt", "Abschlussgebühr".
			txType = "ACCOUNT_FEES"
		}

		amount := r.amount
		var tax int64
		k := key{r.hint, r.dateStr}
		if txType == "INTEREST" && !folded[k] {
			tax = taxes[k]
			amount -= tax
			folded[k] = true
		}

		out = append(out, StagedTransaction{
			Type:        txType,
			Time:        t,
			CashDelta:   amount,
			Taxes:       tax,
			Currency:    "EUR",
			AccountHint: r.hint,
			Note:        r.desc,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no transactions found in LBS Kontoauszug")
	}
	return out, nil
}
