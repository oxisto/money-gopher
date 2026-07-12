package importer

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ING parses ING-DiBa Germany brokerage confirmations.
//
// Supported document types (all one transaction per file):
//   - Wertpapierabrechnung Kauf / Kauf aus Sparplan  → BUY
//   - Wertpapierabrechnung Verkauf                   → SELL
//   - Dividendengutschrift                            → DIVIDEND (stock, may be USD→EUR)
//   - Zinsgutschrift                                  → DIVIDEND (bond coupon, may be USD→EUR)
//   - Ertragsgutschrift                               → DIVIDEND (ETF distribution, EUR)
type ING struct{}

func (ING) Name() string { return "ing" }

func (ING) Matches(text string) bool {
	return strings.Contains(text, "ING-DiBa AG") &&
		(strings.Contains(text, "Wertpapierabrechnung") ||
			strings.Contains(text, "Dividendengutschrift") ||
			strings.Contains(text, "Zinsgutschrift") ||
			strings.Contains(text, "Ertragsgutschrift") ||
			strings.Contains(text, "Ertragsthesaurierung") ||
			strings.Contains(text, "Rückzahlung") ||
			strings.Contains(text, "Wertpapier Eingang") ||
			strings.Contains(text, "Vorabpauschale"))
}

// ING emits one document per transaction, so Parse always returns a single row.
func (ING) Parse(text string) ([]StagedTransaction, error) {
	// pdftotext uses \f (form feed) as a page separator. Only the first page
	// has the transaction summary; later pages are tax detail sheets.
	page1, _, _ := strings.Cut(text, "\f")

	// Storno documents reverse a prior order. We skip them — the original
	// transaction is already in the books (or should have been skipped too).
	// "Neuabrechnung nach Storno" is a replacement document and must NOT be skipped.
	if strings.Contains(page1, "Storno der Ordernummer") ||
		strings.Contains(page1, "*** Storno ***") {
		return nil, fmt.Errorf("storno: %w", ErrSkipped)
	}

	// Pure Vorabpauschale documents (no trade or income header) are a German
	// prepayment tax notice with zero cash flow — nothing to import.
	if strings.Contains(page1, "Vorabpauschale") &&
		!strings.Contains(page1, "Wertpapierabrechnung") &&
		!strings.Contains(page1, "Dividendengutschrift") &&
		!strings.Contains(page1, "Zinsgutschrift") &&
		!strings.Contains(page1, "Ertragsgutschrift") &&
		!strings.Contains(page1, "Ertragsthesaurierung") {
		return nil, fmt.Errorf("vorabpauschale: %w", ErrSkipped)
	}

	if strings.Contains(page1, "Wertpapierabrechnung") {
		tx, err := ingParseTrade(page1)
		if err != nil {
			return nil, err
		}
		return []StagedTransaction{tx}, nil
	}

	if strings.Contains(page1, "Rückzahlung") {
		tx, err := ingParseRepayment(page1)
		if err != nil {
			return nil, err
		}
		return []StagedTransaction{tx}, nil
	}

	if strings.Contains(page1, "Wertpapier Eingang") {
		tx, err := ingParseDelivery(page1)
		if err != nil {
			return nil, err
		}
		return []StagedTransaction{tx}, nil
	}

	tx, err := ingParseIncome(page1)
	if err != nil {
		return nil, err
	}
	return []StagedTransaction{tx}, nil
}

// ── compiled regexes ─────────────────────────────────────────────────────────

var (
	ingISIN     = regexp.MustCompile(`ISIN \(WKN\)\s+([A-Z]{2}[A-Z0-9]{10})`)
	ingIBAN     = regexp.MustCompile(`Abrechnungs-IBAN\s+([A-Z]{2}\d{2}[\d ]+)`)
	ingName     = regexp.MustCompile(`Wertpapierbezeichnung\s+(\S.+)`)
	ingValuta   = regexp.MustCompile(`Valuta\s+(\d{2}\.\d{2}\.\d{4})`)
	ingUnitsStk = regexp.MustCompile(`Nominale\s+Stück\s+([\d,]+)`)          // equity buy: "Nominale Stück 13,15412"
	ingUnitsFV  = regexp.MustCompile(`Nominale\s+[A-Z]{3}\s+([\d.]+,\d+)`)  // bond buy:   "Nominale USD 10.000,00"
	ingUnitsNom = regexp.MustCompile(`Nominale\s+([\d.,]+)\s+Stück`)         // income:     "361,0823 Stück"
	// Currency on the Kurs line; optional qualifier like "(Festpreisgeschäft)" is tolerated.
	ingKursCurr = regexp.MustCompile(`Kurs\s+(?:[^\n]*?\s+)?([A-Z]{3})\s`)
	// Execution date: old format omits " / -zeit".
	ingExecDate = regexp.MustCompile(`Ausführungstag(?: / -zeit)?\s+(\d{2}\.\d{2}\.\d{4})`)
	ingPayDate  = regexp.MustCompile(`Zahltag\s+(\d{2}\.\d{2}\.\d{4})`)
	ingGesamtC  = regexp.MustCompile(`Gesamtbetrag zu Ihren Gunsten\s+([A-Z]{3})\s`)
	// Matches German-format decimals on a line (used by ingFindLineAmount).
	// Allows 2-6 decimal places so fund prices like "8,2656" are captured.
	reGermanDecimal = regexp.MustCompile(`[\d]+(?:\.[\d]{3})*,\d{2,6}`)
)

// ── buy / sell ───────────────────────────────────────────────────────────────

func ingParseTrade(text string) (StagedTransaction, error) {
	txType := "BUY"
	if strings.Contains(text, "Verkauf") {
		txType = "SELL"
	}

	isin, err := ingFind(ingISIN, text, "ISIN")
	if err != nil {
		return StagedTransaction{}, err
	}

	name, _ := ingFind(ingName, text, "")

	// Units: standard equity has "Nominale Stück <n>"; bonds have "Nominale <CCY> <face-value>".
	var units float64
	isBond := false
	if s, uerr := ingFind(ingUnitsStk, text, ""); uerr == nil {
		units, err = ingFloat(s)
		if err != nil {
			return StagedTransaction{}, fmt.Errorf("units: %w", err)
		}
	} else if s, berr := ingFind(ingUnitsFV, text, ""); berr == nil {
		units, err = ingFloat(s)
		if err != nil {
			return StagedTransaction{}, fmt.Errorf("units: %w", err)
		}
		isBond = true
	} else {
		return StagedTransaction{}, fmt.Errorf("could not find Nominale in document")
	}

	// Currency and price.
	// Bonds show "Kurs  94,44 %" — no currency code, price is a percentage.
	// Equity shows "Kurs [optional qualifier]  EUR  26,15".
	currency := "EUR"
	var price *int64
	if !isBond {
		c, cerr := ingFind(ingKursCurr, text, "Kurs currency")
		if cerr != nil {
			return StagedTransaction{}, cerr
		}
		currency = c
		priceStr, perr := ingFindLineAmount(text, "Kurs", currency)
		if perr != nil {
			return StagedTransaction{}, fmt.Errorf("Kurs amount: %w", perr)
		}
		p, perr := ingMinor(priceStr)
		if perr != nil {
			return StagedTransaction{}, fmt.Errorf("price: %w", perr)
		}
		price = &p
	}

	// Execution date; old ING format (pre-2017) uses "Ausführungstag" without " / -zeit".
	// The regex already makes that suffix optional; fall back to Valuta if neither is found.
	dateStr, err := ingFind(ingExecDate, text, "")
	if err != nil {
		dateStr, err = ingFind(ingValuta, text, "Ausführungstag/Valuta")
		if err != nil {
			return StagedTransaction{}, err
		}
	}
	t, err := ingDate(dateStr)
	if err != nil {
		return StagedTransaction{}, err
	}

	// Fees (optional – not all trade types carry a Provision line).
	var fees int64
	if s, ferr := ingFindLineAmount(text, "Provision", ""); ferr == nil {
		fees, err = ingMinor(s)
		if err != nil {
			return StagedTransaction{}, fmt.Errorf("fees: %w", err)
		}
	}

	// Taxes (optional – only on sell with a realised gain).
	taxes := ingTaxSum(text)

	var cashDelta int64
	switch txType {
	case "BUY":
		s, err := ingFindLineAmount(text, "Endbetrag zu Ihren Lasten", "")
		if err != nil {
			return StagedTransaction{}, fmt.Errorf("Endbetrag zu Ihren Lasten: %w", err)
		}
		total, err := ingMinor(s)
		if err != nil {
			return StagedTransaction{}, fmt.Errorf("cash out: %w", err)
		}
		cashDelta = -total

	case "SELL":
		s, err := ingFindLineAmount(text, "Endbetrag zu Ihren Gunsten", "")
		if err != nil {
			return StagedTransaction{}, fmt.Errorf("Endbetrag zu Ihren Gunsten: %w", err)
		}
		cashDelta, err = ingMinor(s)
		if err != nil {
			return StagedTransaction{}, fmt.Errorf("cash in: %w", err)
		}
	}

	iban, _ := ingFind(ingIBAN, text, "")
	iban = strings.ReplaceAll(iban, " ", "")

	return StagedTransaction{
		Type:           txType,
		Time:           t,
		Units:          units,
		Price:          price,
		Fees:           fees,
		Taxes:          taxes,
		CashDelta:      cashDelta,
		Currency:       currency,
		SecurityHint:   strings.TrimSpace(name),
		ISIN:           isin,
		SettlementIBAN: iban,
	}, nil
}

// ── income (dividend / interest / ETF distribution) ──────────────────────────

func ingParseIncome(text string) (StagedTransaction, error) {
	isin, err := ingFind(ingISIN, text, "ISIN")
	if err != nil {
		return StagedTransaction{}, err
	}

	name, _ := ingFind(ingName, text, "")

	// Units: shares for stocks/ETFs; bonds report a face value in a currency,
	// so the Stück pattern won't match and units stay at 0.
	var units float64
	if s, uerr := ingFind(ingUnitsNom, text, ""); uerr == nil {
		units, err = ingFloat(s)
		if err != nil {
			return StagedTransaction{}, fmt.Errorf("units: %w", err)
		}
	}

	// Payment date; fall back to Valuta if not found.
	dateStr, err := ingFind(ingPayDate, text, "")
	if err != nil {
		dateStr, err = ingFind(ingValuta, text, "Zahltag/Valuta")
		if err != nil {
			return StagedTransaction{}, err
		}
	}
	t, err := ingDate(dateStr)
	if err != nil {
		return StagedTransaction{}, err
	}

	// Settlement is always in EUR on ING (foreign dividends are converted).
	currency := "EUR"
	if c, cerr := ingFind(ingGesamtC, text, ""); cerr == nil {
		currency = c
	}

	cashStr, findErr := ingFindLineAmount(text, "Gesamtbetrag zu Ihren Gunsten", "")
	if findErr != nil {
		// Old Vorabpauschale documents (German prepayment tax on fund gains) use
		// "Gesamtbetrag   0,00" without the "zu Ihren Gunsten" suffix when the
		// amount is zero. Skip them — no cash changed hands.
		if strings.Contains(text, "Vorabpauschale") {
			return StagedTransaction{}, fmt.Errorf("vorabpauschale: %w", ErrSkipped)
		}
		return StagedTransaction{}, fmt.Errorf("Gesamtbetrag zu Ihren Gunsten: %w", findErr)
	}
	cashDelta, err := ingMinor(cashStr)
	if err != nil {
		return StagedTransaction{}, fmt.Errorf("cash delta: %w", err)
	}

	taxes := ingTaxSum(text)

	iban, _ := ingFind(ingIBAN, text, "")
	iban = strings.ReplaceAll(iban, " ", "")

	return StagedTransaction{
		Type:           "DIVIDEND",
		Time:           t,
		Units:          units,
		CashDelta:      cashDelta,
		Taxes:          taxes,
		Currency:       currency,
		SecurityHint:   strings.TrimSpace(name),
		ISIN:           isin,
		SettlementIBAN: iban,
	}, nil
}

// ── repayment (Rückzahlung) ───────────────────────────────────────────────────

// ingParseRepayment handles "Rückzahlung" documents — a bond or warrant that
// was redeemed at maturity. It maps to a SELL (the issuer buys it back).
// The Endbetrag line gives the total cash received; Nominale is in Stück.
var ingEndbetrag = regexp.MustCompile(`Endbetrag\s+([A-Z]{3})\s`)

func ingParseRepayment(text string) (StagedTransaction, error) {
	isin, err := ingFind(ingISIN, text, "ISIN")
	if err != nil {
		return StagedTransaction{}, err
	}
	name, _ := ingFind(ingName, text, "")

	unitsStr, err := ingFind(ingUnitsStk, text, "")
	if err != nil {
		// Some Rückzahlungen use "Nominale 150,00 Stück" order.
		unitsStr, err = ingFind(ingUnitsNom, text, "Nominale")
		if err != nil {
			return StagedTransaction{}, err
		}
	}
	units, err := ingFloat(unitsStr)
	if err != nil {
		return StagedTransaction{}, fmt.Errorf("units: %w", err)
	}

	dateStr, err := ingFind(ingValuta, text, "Valuta")
	if err != nil {
		return StagedTransaction{}, err
	}
	t, err := ingDate(dateStr)
	if err != nil {
		return StagedTransaction{}, err
	}

	currency := "EUR"
	if c, cerr := ingFind(ingEndbetrag, text, ""); cerr == nil {
		currency = c
	}

	cashStr, err := ingFindLineAmount(text, "Endbetrag", "")
	if err != nil {
		return StagedTransaction{}, err
	}
	cashDelta, err := ingMinor(cashStr)
	if err != nil {
		return StagedTransaction{}, fmt.Errorf("cash delta: %w", err)
	}

	iban, _ := ingFind(ingIBAN, text, "")
	iban = strings.ReplaceAll(iban, " ", "")

	return StagedTransaction{
		Type:           "SELL",
		Time:           t,
		Units:          units,
		CashDelta:      cashDelta,
		Currency:       currency,
		SecurityHint:   strings.TrimSpace(name),
		ISIN:           isin,
		SettlementIBAN: iban,
	}, nil
}

// ── delivery (Wertpapier Eingang / Bestandsveränderung) ───────────────────────

// ingParseDelivery handles "Wertpapier Eingang" documents — spin-off or
// corporate-action deliveries where units arrive with no cash involved.
// The layout is a multi-column table; ISIN appears as "ISIN (WKN): <value>".
var (
	ingDeliveryISIN  = regexp.MustCompile(`ISIN \(WKN\):\s+([A-Z]{2}[A-Z0-9]{10})`)
	ingDeliveryUnits = regexp.MustCompile(`([\d.,]+)\s+Stück`)
	ingDeliveryName  = regexp.MustCompile(`Stück\s+(\S[^\n]+?)\s{2,}`)
	ingDeliveryDate  = regexp.MustCompile(`(\d{2}\.\d{2}\.\d{4})\s+\d{10}`)
)

func ingParseDelivery(text string) (StagedTransaction, error) {
	isin, err := ingFind(ingDeliveryISIN, text, "ISIN")
	if err != nil {
		return StagedTransaction{}, err
	}

	unitsStr, err := ingFind(ingDeliveryUnits, text, "Stück")
	if err != nil {
		return StagedTransaction{}, err
	}
	units, err := ingFloat(unitsStr)
	if err != nil {
		return StagedTransaction{}, fmt.Errorf("units: %w", err)
	}

	// Security name sits on the same line as the units, after the units value.
	name, _ := ingFind(ingDeliveryName, text, "")

	// Date appears in the table as "DD.MM.YYYY   <order-number>".
	dateStr, err := ingFind(ingDeliveryDate, text, "date")
	if err != nil {
		return StagedTransaction{}, err
	}
	t, err := ingDate(dateStr)
	if err != nil {
		return StagedTransaction{}, err
	}

	return StagedTransaction{
		Type:         "DELIVERY_INBOUND",
		Time:         t,
		Units:        units,
		CashDelta:    0,
		Currency:     "EUR",
		SecurityHint: strings.TrimSpace(name),
		ISIN:         isin,
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

// ingTaxSum sums all German withholding-tax line items.
// Income documents use "Kapitalertragsteuer"; sell documents use
// "KapSt anteilig" (split across two lines at 50% each) — both are handled.
// Kirchensteuer and Solidaritätszuschlag appear on both document types and
// may appear more than once; every occurrence is summed.
func ingTaxSum(text string) int64 {
	var total int64
	for _, line := range strings.Split(text, "\n") {
		isKapSt := strings.Contains(line, "Kapitalertragsteuer") ||
			strings.Contains(line, "KapSt anteilig")
		isKirche := strings.Contains(line, "Kirchensteuer")
		isSoli := strings.Contains(line, "Solidaritätszuschlag")
		if !isKapSt && !isKirche && !isSoli {
			continue
		}
		matches := reGermanDecimal.FindAllString(line, -1)
		if len(matches) == 0 {
			continue
		}
		if v, err := ingMinor(matches[len(matches)-1]); err == nil {
			total += v
		}
	}
	return total
}

// ingFindLineAmount finds the first line containing label (and optionally also
// containing mustContain) and returns the LAST German-format decimal on it.
// Pass mustContain="" to skip the second filter.
func ingFindLineAmount(text, label, mustContain string) (string, error) {
	for _, line := range strings.Split(text, "\n") {
		if !strings.Contains(line, label) {
			continue
		}
		if mustContain != "" && !strings.Contains(line, mustContain) {
			continue
		}
		matches := reGermanDecimal.FindAllString(line, -1)
		if len(matches) > 0 {
			return matches[len(matches)-1], nil
		}
	}
	return "", fmt.Errorf("could not find %q in document", label)
}

// ingFind returns the first capture group of re in text.
// Pass field="" to suppress the error for optional fields.
func ingFind(re *regexp.Regexp, text, field string) (string, error) {
	m := re.FindStringSubmatch(text)
	if m == nil {
		if field == "" {
			return "", fmt.Errorf("not found")
		}
		return "", fmt.Errorf("could not find %q in document", field)
	}
	return strings.TrimSpace(m[1]), nil
}

// ingDate parses a German date (DD.MM.YYYY) to UTC midnight.
func ingDate(s string) (time.Time, error) {
	t, err := time.Parse("02.01.2006", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", s, err)
	}
	return t.UTC(), nil
}

// ingFloat parses a German locale decimal string (dot=thousands, comma=decimal) to float64.
func ingFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

// ingMinor converts a German decimal string to minor monetary units (cents).
func ingMinor(s string) (int64, error) {
	f, err := ingFloat(s)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", s, err)
	}
	return int64(math.Round(f * 100)), nil
}
