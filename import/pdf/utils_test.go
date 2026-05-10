package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func resetLocalSubsCache() {
	localSubs = nil
	localSubsOnce = sync.Once{}
}

func TestSanitizeText_AppliesLocalJSONAndPreservesISIN(t *testing.T) {
	tmp := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// write JSON substitutions
	jsonFile := filepath.Join(tmp, ".pdfdump.local.json")
	if err := os.WriteFile(jsonFile, []byte(`{
	"My Name": "Max Mustermann",
	"My Street": "Musterstraße 1",
	"0987654321": "1234567890"
}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	resetLocalSubsCache()

	in := "ISIN DE000BASF111\nName: My Name\nAddress: My Street\nAccount: 0987654321\nEmail: test@example.com"
	out := SanitizeText(in)

	if !strings.Contains(out, "Musterstraße 1") {
		t.Fatalf("expected address substituted, got %q", out)
	}
	if !strings.Contains(out, "DE000BASF111") {
		t.Fatalf("expected ISIN preserved, got %q", out)
	}
	if !strings.Contains(out, "[EMAIL REDACTED]") {
		t.Fatalf("expected email redacted, got %q", out)
	}
	if !strings.Contains(out, "1234567890") {
		t.Fatalf("expected account substituted, got %q", out)
	}
}

func TestSanitizeText_AppliesPlainFallbackSubstitutions(t *testing.T) {
	tmp := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// write plain .pdfdump.local
	plain := filepath.Join(tmp, ".pdfdump.local")
	if err := os.WriteFile(plain, []byte(`# plain substitutions
Foo Account => BAR123
Mühlfeldweg 22 => Musterstraße 1
`), 0o644); err != nil {
		t.Fatalf("write plain: %v", err)
	}

	resetLocalSubsCache()

	in := "Account: Foo Account\nAddress: Mühlfeldweg 22"
	out := SanitizeText(in)

	if !strings.Contains(out, "BAR123") {
		t.Fatalf("expected plain substitution applied, got %q", out)
	}
	if !strings.Contains(out, "Musterstraße 1") {
		t.Fatalf("expected plain address substitution applied, got %q", out)
	}
}

func TestExtractTextPagesFromPDFBytes_SplitsPages(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "internal", "testdata", "flatex-verkauf-2026-01.pdf")
	b, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	pages, err := ExtractTextPagesFromPDFBytes(b)
	if err != nil {
		t.Fatalf("extract pages: %v", err)
	}
	if len(pages) == 0 {
		t.Fatalf("expected >=1 page, got 0")
	}
	// if there are multiple pages, ensure they are not identical
	if len(pages) > 1 && pages[0] == pages[1] {
		t.Fatalf("expected different content on page 1 and 2")
	}
}

func TestExtractTextFromPDFBytes_UsesPageSeparator(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "internal", "testdata", "flatex-verkauf-2026-01.pdf")
	b, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	s, err := ExtractTextFromPDFBytes(b)
	if err != nil {
		t.Fatalf("extract text: %v", err)
	}
	if !strings.Contains(s, "\n---\n") {
		t.Fatalf("expected page separator '---' in ExtractTextFromPDFBytes output")
	}
}
