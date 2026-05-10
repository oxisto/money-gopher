package main

// Command-line PDF → text extractor (same behavior as tools/pdfdump).
// This lives under /cmd so it can be installed with `go install ./cmd/pdfdump`.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/oxisto/money-gopher/import/pdf"
)

var (
	flagOut      string
	flagSanitize bool
	flagPages    bool
)

func init() {
	flag.StringVar(&flagOut, "out", "", "output file or directory for extracted text (default stdout)")
	flag.BoolVar(&flagSanitize, "sanitize", false, "sanitize personal data (mask IBANs, names, emails, account numbers)")
	flag.BoolVar(&flagPages, "pages", false, "write individual pages separately (useful for page-local parsing)")
}

func extractTextFromPDF(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return pdf.ExtractTextFromPDFBytes(b)
}

func writeOutput(outPath, content string) error {
	if outPath == "" {
		fmt.Print(content)
		return nil
	}

	info, err := os.Stat(outPath)
	if err == nil && info.IsDir() {
		return os.WriteFile(filepath.Join(outPath, "output.txt"), []byte(content), 0o644)
	}

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
		if flagPages {
			b, err := os.ReadFile(p)
			if err != nil {
				log.Fatalf("read %s: %v", p, err)
			}
			pages, err := pdf.ExtractTextPagesFromPDFBytes(b)
			if err != nil {
				log.Fatalf("extract pages %s: %v", p, err)
			}

			for i, pg := range pages {
				if flagSanitize {
					pg = pdf.SanitizeText(pg)
				}

				// if out is a directory, write each page separately
				if flagOut != "" {
					if info, err := os.Stat(flagOut); err == nil && info.IsDir() {
						outFile := filepath.Join(flagOut, strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))+"-page-"+fmt.Sprint(i+1)+".txt")
						if err := os.WriteFile(outFile, []byte(pg), 0o644); err != nil {
							log.Fatalf("write %s: %v", outFile, err)
						}
						fmt.Printf("wrote %s\n", outFile)
						continue
					}

					// single-file output: append pages separated by headers
					sep := "\n----- PAGE " + fmt.Sprint(i+1) + " -----\n"
					if err := os.WriteFile(flagOut, []byte(sep+pg), 0o644); err != nil {
						log.Fatalf("write %s: %v", flagOut, err)
					}
					continue
				}

				// default: print page with header to stdout
				fmt.Printf("----- %s (page %d) -----\n%s\n", p, i+1, pg)
			}

			continue
		}

		text, err := extractTextFromPDF(p)
		if err != nil {
			log.Fatalf("extract %s: %v", p, err)
		}

		if flagSanitize {
			text = pdf.SanitizeText(text)
		}

		if flagOut != "" {
			if info, err := os.Stat(flagOut); err == nil && info.IsDir() {
				outFile := filepath.Join(flagOut, strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))+".txt")
				if err := os.WriteFile(outFile, []byte(text), 0o644); err != nil {
					log.Fatalf("write %s: %v", outFile, err)
				}
				fmt.Printf("wrote %s\n", outFile)
				continue
			}

			if err := writeOutput(flagOut, text); err != nil {
				log.Fatalf("write %s: %v", flagOut, err)
			}
			fmt.Printf("wrote %s\n", flagOut)
			continue
		}

		fmt.Printf("----- %s -----\n%s\n", p, text)
	}
}
