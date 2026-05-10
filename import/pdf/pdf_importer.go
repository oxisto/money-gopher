// package pdf contains a PDF importer scaffold for portfolio events.
//
// This is a test-first scaffold: tests and expected fixtures are provided
// under `internal/testdata`. The real parser implementation is TODO and
// should be implemented to extract transactions and securities from
// broker/custodian statement PDFs.
package pdf

import (
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ErrNotImplemented is returned while the PDF importer is a scaffold.
var ErrNotImplemented = errors.New("pdf importer not implemented")

// Import reads portfolio events and securities from the provided PDF reader.
// - r: PDF bytes reader
// - pname: portfolio name (used to populate PortfolioId on events)
// Returns a slice of PortfolioEvent and Security messages or an error.
//
// This implementation is intentionally small and heuristic-driven: it
// extracts plain text from the PDF and looks for a few well-known
// keywords/regexes ("verkauf" / "dividende") to return minimal
// `PortfolioEvent` objects for the test fixtures. A full production
// parser should replace the heuristics below.
func Import(r io.Reader, pname string) (txs []*portfoliov1.PortfolioEvent, secs []*portfoliov1.Security, err error) {
	// Read all bytes (we need a ReaderAt for the PDF library)
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	pages, err := ExtractTextPagesFromPDFBytes(b)
	if err != nil {
		return nil, nil, err
	}

	// Parse each page independently first; many statements contain
	// one trade per page which simplifies extraction and avoids
	// cross-page regex matches.
	for _, pg := range pages {
		ptxs, _, _ := parseText(pg, pname)
		if len(ptxs) > 0 {
			txs = append(txs, ptxs...)
		}
	}

	// Keep a full-document version for fallback and dividend detection.
	// deduplicate any page-local parsing results collected so far (avoid
	// the same trade being appended multiple times when text repeats on
	// different pages)
	{
		seen := map[string]struct{}{}
		uniqPage := make([]*portfoliov1.PortfolioEvent, 0, len(txs))
		for _, tx := range txs {
			var priceCents int32
			if tx.Price != nil {
				priceCents = tx.Price.GetValue()
			}
			key := fmt.Sprintf("%d|%s|%v|%d", tx.Type, tx.SecurityId, tx.Amount, priceCents)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			uniqPage = append(uniqPage, tx)
		}
		txs = uniqPage
	}

	// if we found any transactions on a per-page basis, return them now
	// — skipping full-document parsing avoids duplicates caused by
	// overlapping page text (many PDFs repeat blocks across page extracts).
	if len(txs) > 0 {
		return txs, nil, nil
	}

	origText := strings.Join(pages, "\n")
	fullText := strings.ToLower(origText)

	// --- SELL / BUY blocks: find all 'Nr.' header indices and extract blocks ---
	reHeader := regexp.MustCompile(`(?m)^Nr\.\s*\d+\/\d+`)
	locs := reHeader.FindAllStringIndex(origText, -1)
	blocks := make([]string, 0, len(locs))
	for i, loc := range locs {
		start := loc[0]
		end := len(origText)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		blocks = append(blocks, origText[start:end])
	}

	for _, blk := range blocks {
		l := strings.ToLower(blk)
		var typ portfoliov1.PortfolioEventType
		if strings.Contains(l, "verkauf") || strings.Contains(l, "sell") {
			typ = portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL
		} else if strings.Contains(l, "kauf") || strings.Contains(l, "buy") {
			typ = portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY
		} else {
			continue
		}

		// extract fields
		isin := extractISIN(blk)
		amount := findFirstNumber(blk, `Ordervolumen\s*:\s*([0-9.,]+)\s*St`)
		if amount == 0 {
			// fallback: look for "von ausgef." or similar share counts
			amount = findFirstNumber(blk, `von ausgef\.\s*:\s*([0-9.,]+)\s*St`)
		}
		price := findFirstNumber(blk, `Kurs\s*:\s*([0-9.,]+)\s*EUR`)
		kurswert := findFirstNumber(blk, `Kurswert\s*:\s*([0-9.,]+)\s*EUR`)
		valutaStr := findFirstString(blk, `Valuta\s*:\s*([0-9]{2}\.[0-9]{2}\.[0-9]{4})`)
		var t time.Time
		if valutaStr != "" {
			if tt, err := time.ParseInLocation("02.01.2006", valutaStr, time.Local); err == nil {
				t = tt
			}
		}

		tx := &portfoliov1.PortfolioEvent{
			Type:        typ,
			PortfolioId: pname,
			SecurityId:  isin,
			Amount:      amount,
		}
		if !t.IsZero() {
			tx.Time = timestamppb.New(t)
		} else {
			tx.Time = timestamppb.New(time.Now())
		}

		if price > 0 {
			tx.Price = portfoliov1.Value(int32(math.Round(price * 100)))
		} else if kurswert > 0 && tx.Amount != 0 {
			// derive unit price from kurswert
			unit := kurswert / tx.Amount
			tx.Price = portfoliov1.Value(int32(math.Round(unit * 100)))
		}

		if kurswert > 0 {
			// set Fees/Taxes/other fields later if desired — not implemented
		}

		tx.MakeUniqueID()
		txs = append(txs, tx)
	}

	// --- DIVIDEND / ERTRAGSABLAGEN ---
	if strings.Contains(fullText, "dividende") || strings.Contains(fullText, "dividend") || strings.Contains(fullText, "ertragsgutschrift") || strings.Contains(fullText, "ausschütt") {
		// extract ISIN and common dividend fields
		isin := extractISIN(origText)
		shares := findFirstNumber(origText, `Nominale\s*([0-9.,]+)\s*(St|Stück)`) // number of shares
		// Prefer the "Gesamtbetrag zu Ihren Gunsten" (net EUR amount) if present
		netEUR := findFirstNumber(origText, `Gesamtbetrag\s+zu\s+Ihren\s+Gunsten\s+EUR\s*([0-9.,]+)`)
		if netEUR == 0 {
			// fallback to converted EUR gross
			netEUR = findFirstNumber(origText, `EUR\s*([0-9.,]+)\s*(?:\n|$)`) // crude fallback
		}
		valutaStr := findFirstString(origText, `Valuta\s*([0-9]{2}\.[0-9]{2}\.[0-9]{4})`)
		var t time.Time
		if valutaStr != "" {
			if tt, err := time.ParseInLocation("02.01.2006", valutaStr, time.Local); err == nil {
				t = tt
			}
		}

		tx := &portfoliov1.PortfolioEvent{
			Type:        portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DIVIDEND,
			PortfolioId: pname,
			SecurityId:  isin,
			Amount:      shares,
		}
		if !t.IsZero() {
			tx.Time = timestamppb.New(t)
		} else {
			tx.Time = timestamppb.New(time.Now())
		}

		if netEUR > 0 {
			tx.Price = portfoliov1.Value(int32(math.Round(netEUR * 100)))
		}

		tx.MakeUniqueID()
		txs = append(txs, tx)
	}

	return txs, nil, nil
}

// parseText parses plain/extracted PDF text into PortfolioEvent objects.
// This lets unit tests use sanitized plaintext fixtures instead of raw PDFs.
func parseText(origText string, pname string) (txs []*portfoliov1.PortfolioEvent, secs []*portfoliov1.Security, err error) {
	fullText := strings.ToLower(origText)

	// --- SELL / BUY blocks: find all 'Nr.' header indices and extract blocks ---
	reHeader := regexp.MustCompile(`(?m)^Nr\.\s*\d+\/\d+`)
	locs := reHeader.FindAllStringIndex(origText, -1)
	blocks := make([]string, 0, len(locs))
	for i, loc := range locs {
		start := loc[0]
		end := len(origText)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		blocks = append(blocks, origText[start:end])
	}

	for _, blk := range blocks {
		l := strings.ToLower(blk)
		var typ portfoliov1.PortfolioEventType
		if strings.Contains(l, "verkauf") || strings.Contains(l, "sell") {
			typ = portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL
		} else if strings.Contains(l, "kauf") || strings.Contains(l, "buy") {
			typ = portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY
		} else {
			continue
		}

		// extract fields
		isin := extractISIN(blk)
		amount := findFirstNumber(blk, `Ordervolumen\s*:\s*([0-9.,]+)\s*St`)
		if amount == 0 {
			// fallback: look for "von ausgef." or similar share counts
			amount = findFirstNumber(blk, `von ausgef\.\s*:\s*([0-9.,]+)\s*St`)
		}
		price := findFirstNumber(blk, `Kurs\s*:\s*([0-9.,]+)\s*EUR`)
		kurswert := findFirstNumber(blk, `Kurswert\s*:\s*([0-9.,]+)\s*EUR`)
		valutaStr := findFirstString(blk, `Valuta\s*:\s*([0-9]{2}\.[0-9]{2}\.[0-9]{4})`)
		var t time.Time
		if valutaStr != "" {
			if tt, err := time.ParseInLocation("02.01.2006", valutaStr, time.Local); err == nil {
				t = tt
			}
		}

		tx := &portfoliov1.PortfolioEvent{
			Type:        typ,
			PortfolioId: pname,
			SecurityId:  isin,
			Amount:      amount,
		}
		if !t.IsZero() {
			tx.Time = timestamppb.New(t)
		} else {
			tx.Time = timestamppb.New(time.Now())
		}

		if price > 0 {
			tx.Price = portfoliov1.Value(int32(math.Round(price * 100)))
		} else if kurswert > 0 && tx.Amount != 0 {
			// derive unit price from kurswert
			unit := kurswert / tx.Amount
			tx.Price = portfoliov1.Value(int32(math.Round(unit * 100)))
		}

		tx.MakeUniqueID()
		txs = append(txs, tx)
	}

	// --- DIVIDEND / ERTRAGSABLAGEN ---
	if strings.Contains(fullText, "dividende") || strings.Contains(fullText, "dividend") || strings.Contains(fullText, "ertragsgutschrift") || strings.Contains(fullText, "ausschütt") {
		// extract ISIN and common dividend fields
		isin := extractISIN(origText)
		shares := findFirstNumber(origText, `Nominale\s*([0-9.,]+)\s*(St|Stück)`) // number of shares
		// Prefer the "Gesamtbetrag zu Ihren Gunsten" (net EUR amount) if present
		netEUR := findFirstNumber(origText, `Gesamtbetrag\s+zu\s+Ihren\s+Gunsten\s+EUR\s*([0-9.,]+)`)
		if netEUR == 0 {
			// fallback to converted EUR gross
			netEUR = findFirstNumber(origText, `EUR\s*([0-9.,]+)\s*(?:\n|$)`) // crude fallback
		}
		valutaStr := findFirstString(origText, `Valuta\s*([0-9]{2}\.[0-9]{2}\.[0-9]{4})`)
		var t time.Time
		if valutaStr != "" {
			if tt, err := time.ParseInLocation("02.01.2006", valutaStr, time.Local); err == nil {
				t = tt
			}
		}

		tx := &portfoliov1.PortfolioEvent{
			Type:        portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DIVIDEND,
			PortfolioId: pname,
			SecurityId:  isin,
			Amount:      shares,
		}
		if !t.IsZero() {
			tx.Time = timestamppb.New(t)
		} else {
			tx.Time = timestamppb.New(time.Now())
		}

		if netEUR > 0 {
			tx.Price = portfoliov1.Value(int32(math.Round(netEUR * 100)))
		}

		tx.MakeUniqueID()
		txs = append(txs, tx)
	}

	return txs, nil, nil
}

// extractISIN looks for an ISIN-like string (2 letters + 10 alnum) and
// returns the first match or an empty string.
func extractISIN(s string) string {
	re := regexp.MustCompile(`[A-Z]{2}[A-Z0-9]{10}`)
	m := re.FindString(s)
	return strings.ToUpper(m)
}

func findFirstString(s string, pattern string) string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(s)
	if len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func findFirstNumber(s string, pattern string) float64 {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(s)
	if len(m) >= 2 {
		return parseGermanFloat(m[1])
	}
	return 0
}

func parseGermanFloat(raw string) float64 {
	rs := strings.TrimSpace(raw)
	// remove thousands separator and convert comma to dot
	rs = strings.ReplaceAll(rs, ".", "")
	rs = strings.ReplaceAll(rs, ",", ".")
	v, err := strconv.ParseFloat(rs, 64)
	if err != nil {
		return 0
	}
	// keep a reasonable precision
	return math.Round(v*1000000) / 1000000
}
