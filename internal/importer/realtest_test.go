//go:build realtest

package importer

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestING_RealPDFs(t *testing.T) {
	bases := []string{
		"/Users/oxisto/Library/CloudStorage/OneDrive-Personal/Documents/Finanzen/ING/Kauf",
		"/Users/oxisto/Library/CloudStorage/OneDrive-Personal/Documents/Finanzen/ING/Verkauf",
		"/Users/oxisto/Library/CloudStorage/OneDrive-Personal/Documents/Finanzen/ING/Dividenden",
		"/Users/oxisto/Library/CloudStorage/OneDrive-Personal/Documents/Finanzen/ING/Zinsen",
		"/Users/oxisto/Library/CloudStorage/OneDrive-Personal/Documents/Finanzen/ING Christian",
	}

	ok, errs := 0, 0
	errMsgs := map[string]int{}
	parser := ING{}

	for _, base := range bases {
		pdfs, _ := filepath.Glob(filepath.Join(base, "*.pdf"))
		for _, pdf := range pdfs {
			out, err := exec.Command("pdftotext", "-layout", pdf, "-").Output()
			if err != nil {
				continue
			}
			text := string(out)
			if !parser.Matches(text) {
				continue
			}
			_, parseErr := parser.Parse(text)
			if parseErr != nil {
				errs++
				errMsgs[parseErr.Error()]++
				t.Logf("FAIL %s: %v", filepath.Base(pdf), parseErr)
			} else {
				ok++
			}
		}
	}

	t.Logf("Parsed OK: %d  Errors: %d", ok, errs)
	for msg, count := range errMsgs {
		short := msg
		if idx := strings.Index(msg, `"`); idx >= 0 {
			if end := strings.Index(msg[idx+1:], `"`); end >= 0 {
				short = msg[:idx+1] + msg[idx+1:][:end+1]
			}
		}
		t.Logf("  %3d  %s", count, short)
	}
}
