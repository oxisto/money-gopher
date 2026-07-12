package importer

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// csvHeader is the fingerprint and column order of the generic CSV format.
const csvHeader = "Date,Type,Security,ISIN,Units,Price,Fees,Taxes,CashDelta,Currency"

// CSV parses money-gopher's own generic CSV format, the reference
// implementation of the staging pipeline (bank PDFs ride the same path).
//
// The format is one header line followed by one transaction per line:
//
//	Date,Type,Security,ISIN,Units,Price,Fees,Taxes,CashDelta,Currency
//	2026-01-03,BUY,iShares Core MSCI World,IE00B4L5Y983,10,100.50,1.50,0,,EUR
//
// Dates are YYYY-MM-DD (or RFC 3339); amounts are decimals in major units
// with a dot separator and no thousands separators. CashDelta may be empty
// for trades and dividends — it is derived at confirmation.
type CSV struct{}

func (CSV) Name() string { return "csv" }

func (CSV) Matches(text string) bool {
	first, _, _ := strings.Cut(strings.TrimLeft(text, "\uFEFF\n\r "), "\n")
	return strings.EqualFold(strings.TrimSpace(first), csvHeader)
}

func (CSV) Parse(text string) ([]StagedTransaction, error) {
	r := csv.NewReader(strings.NewReader(strings.TrimLeft(text, "\uFEFF")))
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("no transactions in CSV")
	}

	staged := make([]StagedTransaction, 0, len(records)-1)
	for i, rec := range records[1:] {
		line := i + 2

		if len(rec) != 10 {
			return nil, fmt.Errorf("line %d: expected 10 columns, got %d", line, len(rec))
		}

		tx := StagedTransaction{
			Type:         strings.ToUpper(strings.TrimSpace(rec[1])),
			SecurityHint: strings.TrimSpace(rec[2]),
			ISIN:         strings.ToUpper(strings.TrimSpace(rec[3])),
			Currency:     strings.ToUpper(strings.TrimSpace(rec[9])),
		}

		tx.Time, err = parseDate(rec[0])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}

		if units := strings.TrimSpace(rec[4]); units != "" {
			tx.Units, err = strconv.ParseFloat(units, 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid units %q", line, units)
			}
		}

		if price := strings.TrimSpace(rec[5]); price != "" {
			p, err := parseDecimal(price)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
			tx.Price = &p
		}

		for _, f := range []struct {
			value string
			dst   *int64
		}{{rec[6], &tx.Fees}, {rec[7], &tx.Taxes}, {rec[8], &tx.CashDelta}} {
			*f.dst, err = parseDecimal(f.value)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
		}

		staged = append(staged, tx)
	}

	return staged, nil
}

// parseDate accepts YYYY-MM-DD or RFC 3339, always in UTC.
func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}

	return time.Time{}, fmt.Errorf("invalid date %q", s)
}
