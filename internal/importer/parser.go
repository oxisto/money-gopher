package importer

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// StagedTransaction is one transaction parsed from a document, before user
// review. It carries matching hints (security name, ISIN) instead of hard
// references; those are resolved during review and confirmation.
type StagedTransaction struct {
	Type  string
	Time  time.Time
	Units float64
	// Price per unit in minor units; nil when the document has none.
	Price *int64
	Fees  int64
	Taxes int64
	// CashDelta in minor units; 0 lets the server derive it for trades.
	CashDelta int64
	Currency  string
	// SecurityHint is the security name as printed on the document.
	SecurityHint string
	// ISIN as printed on the document.
	ISIN string
}

// Parser understands one bank's (or format's) documents.
type Parser interface {
	// Name identifies the parser, stored as the document's detected bank.
	Name() string
	// Matches is the fingerprint: does the extracted text look like a
	// document of this format?
	Matches(text string) bool
	// Parse returns the staged transactions found in the text.
	Parse(text string) ([]StagedTransaction, error)
}

// DefaultParsers returns all built-in parsers. Adding a bank means adding a
// package and listing its parser here.
func DefaultParsers() []Parser {
	return []Parser{ING{}, CSV{}}
}

// detect returns the first parser whose fingerprint matches.
func detect(parsers []Parser, text string) (Parser, bool) {
	for _, p := range parsers {
		if p.Matches(text) {
			return p, true
		}
	}
	return nil, false
}

// parseDecimal parses a decimal amount in major units (e.g. "1234.56") into
// minor units. The empty string is zero.
func parseDecimal(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", s)
	}

	return int64(math.Round(value * 100)), nil
}
