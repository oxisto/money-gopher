package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	pdf "github.com/ledongthuc/pdf"
)

var (
	flagOut      string
	flagSanitize bool
	localSubs     map[string]string
	localSubsOnce sync.Once
)

func init() {
	flag.StringVar(&flagOut, "out", "", "output file or directory for extracted text (default stdout)")
	flag.BoolVar(&flagSanitize, "sanitize", false, "sanitize personal data (mask IBANs, names, emails, account numbers)")
}

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
				if b, err := os.ReadFile(p); err == nil {
					var m map[string]string
					if json.Unmarshal(b, &m) == nil {
						localSubs = m
						return
					}
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
		}
	})
	return localSubs
}

func sanitizeText(s string) string {
	// apply user local substitutions first (if present)
	if subs := getLocalSubs(); len(subs) > 0 {
		for from, to := range subs {
			if from == "" {
				continue
			}
			s = strings.ReplaceAll(s, from, to)
		}
	}

	// Preserve ISINs (not PII): temporarily replace with placeholders.
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
	reNames := regexp.MustCompile(`(?m)^(Christian Banse|Veronika Banse|[A-Z][a-z]+\s+[A-Z][a-z]+).*$`)
	s = reNames.ReplaceAllStringFunc(s, func(m string) string {
		// only redact if line contains a space and letters (avoid removing titles)
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

func extractTextFromPDF(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	st, _ := f.Stat()
	r, err := pdf.NewReader(f, st.Size())
	if err != nil {
		return "", err
	}

	var b strings.Builder
	if rdr, err := r.GetPlainText(); err == nil {
		bb, _ := io.ReadAll(rdr)
		b.Write(bb)
		return b.String(), nil
	}

	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		if txt, err := p.GetPlainText(nil); err == nil {
			b.WriteString(txt)
		}
	}

	return b.String(), nil
}

func writeOutput(outPath, content string) error {
	if outPath == "" {
		fmt.Print(content)
		return nil
	}

	// if outPath is a directory, write to <dir>/<basename>.txt
	info, err := os.Stat(outPath)
	if err == nil && info.IsDir() {
		return os.WriteFile(filepath.Join(outPath, "output.txt"), []byte(content), 0o644)
	}

	// otherwise create parent dirs and write file
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(content), 0o644)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("usage: pdfdump [flags] <pdf-path> [<pdf-path> ...]\nFlags:\n  -out <file|dir>   write output to file or directory\n  -sanitize         mask personal data in extracted text")
	}

	for _, p := range args {
		text, err := extractTextFromPDF(p)
		if err != nil {
			log.Fatalf("extract %s: %v", p, err)
		}

		if flagSanitize {
			text = sanitizeText(text)
		}

		if flagOut != "" {
			// if multiple inputs and out is a directory, write each to basename.txt
			if info, err := os.Stat(flagOut); err == nil && info.IsDir() {
				outFile := filepath.Join(flagOut, strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))+".txt")
				if err := os.WriteFile(outFile, []byte(text), 0o644); err != nil {
					log.Fatalf("write %s: %v", outFile, err)
				}
				fmt.Printf("wrote %s\n", outFile)
				continue
			}

			// single-file output
			if err := writeOutput(flagOut, text); err != nil {
				log.Fatalf("write %s: %v", flagOut, err)
			}
			fmt.Printf("wrote %s\n", flagOut)
			continue
		}

		// default: stdout
		fmt.Printf("----- %s -----\n%s\n", p, text)
	}
}
