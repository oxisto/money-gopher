package pdf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	pdfreader "github.com/ledongthuc/pdf"
)

var (
	localSubs     map[string]string
	localSubsOnce sync.Once
)

// ExtractTextPagesFromPDFBytes extracts plain text for each PDF page and
// returns a slice where element 0 corresponds to page 1. This is helpful for
// page-local parsing (many broker statements put single trades on separate
// pages). Use this when you need page-level isolation.
func ExtractTextPagesFromPDFBytes(b []byte) ([]string, error) {
	br := bytes.NewReader(b)
	rd, err := pdfreader.NewReader(br, int64(len(b)))
	if err != nil {
		return nil, err
	}

	// Attempt a cheap whole-document plain-text first — some PDFs provide a
	// single reader that yields the whole text. If that's available we still
	// split it by page-delimiter fallback below.
	if rdr, err := rd.GetPlainText(); err == nil {
		bb, _ := io.ReadAll(rdr)
		all := string(bb)

		// If the reader already contains explicit page separators, honor them.
		// support both form-feed (legacy) and the repository-standard `---`.
		if strings.Contains(all, "\f") {
			parts := strings.Split(all, "\f")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.Trim(p, "\n\r")
				if p == "" {
					continue
				}
				out = append(out, p)
			}
			if len(out) > 0 {
				return out, nil
			}
		}

		if strings.Contains(all, "\n---\n") {
			parts := strings.Split(all, "\n---\n")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.Trim(p, "\n\r")
				if p == "" {
					continue
				}
				out = append(out, p)
			}
			if len(out) > 0 {
				return out, nil
			}
		}
		// fallback to per-page iteration below if no explicit separators found
	}

	pages := make([]string, 0, rd.NumPage())
	for i := 1; i <= rd.NumPage(); i++ {
		p := rd.Page(i)
		if p.V.IsNull() {
			pages = append(pages, "")
			continue
		}
		if txt, err := p.GetPlainText(nil); err == nil {
			pages = append(pages, txt)
		} else {
			pages = append(pages, "")
		}
	}

	return pages, nil
}

// ExtractTextFromPDFBytes is kept for backward compatibility and simply joins
// the per-page texts into a single string using `---` as the page separator.
func ExtractTextFromPDFBytes(b []byte) (string, error) {
	pages, err := ExtractTextPagesFromPDFBytes(b)
	if err != nil {
		return "", err
	}
	return strings.Join(pages, "\n---\n"), nil
}

// SanitizeText applies project-standard sanitization rules and also
// applies any user-provided local substitutions (from .pdfdump.local(.json)).
// Local substitutions are intentionally loaded from working dir then $HOME
// and are not committed to git (see .gitignore).
func SanitizeText(s string) string {
	// apply user local substitutions first (if present)
	if subs := getLocalSubs(); len(subs) > 0 {
		for from, to := range subs {
			if from == "" {
				continue
			}
			s = strings.ReplaceAll(s, from, to)
		}
	}

	// Preserve ISINs (not PII): temporarily replace with placeholders so
	// subsequent redactions don't remove them.
	reISIN := regexp.MustCompile(`\b[A-Z]{2}[A-Z0-9]{10}\b`)
	savedISIN := map[string]string{}
	isn := 0
	s = reISIN.ReplaceAllStringFunc(s, func(m string) string {
		key := "__ISIN_" + fmt.Sprint(isn) + "__"
		savedISIN[key] = m
		isn++
		return key
	})

	// emails
	reEmail := regexp.MustCompile(`(?i)\b[\w._%+\-]+@[\w.\-]+\.[A-Za-z]{2,}\b`)
	s = reEmail.ReplaceAllString(s, "[EMAIL REDACTED]")

	// IBAN-like (common) — mask fully (require length >= 15 to avoid ISINs)
	reIBAN := regexp.MustCompile(`\b[A-Z]{2}[0-9]{2}[0-9A-Z]{11,}\b`)
	s = reIBAN.ReplaceAllString(s, "[IBAN REDACTED]")

	// labeled account/depot numbers (keep other numeric amounts intact)
	reLabels := regexp.MustCompile(`(?im)^(Depotnummer|Direkt-Depot Nr\.|Direkt-Depot Nr|Abrechnungs-IBAN|Depotinhaber|Abrechnungs-IBAN|IBAN|Abrechnungs-IBAN:)\s*[:\-\s]*(.+)$`)
	s = reLabels.ReplaceAllString(s, "[REDACTED]")

	// obvious personal-name lines (very conservative)
	reNames := regexp.MustCompile(`(?m)^([A-Z][a-z]+\s+[A-Z][a-z]+).*$`)
	s = reNames.ReplaceAllStringFunc(s, func(m string) string {
		if strings.Contains(m, " ") {
			return "[NAME REDACTED]"
		}
		return m
	})

	// phone numbers (require leading + or labeled 'Tel.')
	rePhone := regexp.MustCompile(`(?m)(?:Tel\.|Tel:)?\s*(?:\+)[0-9][0-9 \-\(\)]{6,}`)
	s = rePhone.ReplaceAllString(s, "[PHONE REDACTED]")

	// Restore ISIN placeholders
	for k, v := range savedISIN {
		s = strings.ReplaceAll(s, k, v)
	}

	return s
}

// getLocalSubs loads local substitution mappings from .pdfdump.local(.json)
// in the repository root or the user's home directory. Returns empty map
// if none found.
func getLocalSubs() map[string]string {
	localSubsOnce.Do(func() {
		localSubs = map[string]string{}
		candidates := []string{".pdfdump.local.json", ".pdfdump.local"}

		cwd, _ := os.Getwd()
		paths := []string{cwd}
		if h, err := os.UserHomeDir(); err == nil {
			paths = append(paths, h)
		}

		for _, dir := range paths {
			for _, name := range candidates {
				p := filepath.Join(dir, name)
				b, err := os.ReadFile(p)
				if err != nil {
					continue
				}

				var m map[string]string
				if json.Unmarshal(b, &m) == nil {
					localSubs = m
					return
				}

				// fallback: simple `from => to` lines
				lines := strings.Split(string(b), "\n")
				for _, ln := range lines {
					ln = strings.TrimSpace(ln)
					if ln == "" || strings.HasPrefix(ln, "#") {
						continue
					}
					parts := strings.SplitN(ln, "=>", 2)
					if len(parts) == 2 {
						k := strings.TrimSpace(parts[0])
						v := strings.TrimSpace(parts[1])
						localSubs[k] = v
					}
				}
				return
			}
		}
	})
	return localSubs
}
