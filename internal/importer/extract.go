package importer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Extractor turns an uploaded file into plain text for the bank parsers.
type Extractor interface {
	// Supports reports whether the extractor handles the content type.
	Supports(contentType string) bool
	// Extract returns the plain text of the file.
	Extract(ctx context.Context, data []byte) (string, error)
}

// PDFToText extracts text from PDFs by shelling out to poppler's pdftotext
// with layout preservation, which keeps the column structure bank parsers
// fingerprint on. The binary's presence is checked at startup and surfaced
// as a per-document error, not a crash.
type PDFToText struct {
	// Path to the pdftotext binary; empty if it was not found.
	Path string
}

// NewPDFToText locates pdftotext in PATH.
func NewPDFToText() *PDFToText {
	path, err := exec.LookPath("pdftotext")
	if err != nil {
		return &PDFToText{}
	}
	return &PDFToText{Path: path}
}

func (p *PDFToText) Supports(contentType string) bool {
	return strings.HasPrefix(contentType, "application/pdf")
}

func (p *PDFToText) Extract(ctx context.Context, data []byte) (string, error) {
	if p.Path == "" {
		return "", fmt.Errorf("pdftotext is not installed; install poppler to import PDFs")
	}

	// pdftotext wants a file, not stdin.
	tmp, err := os.CreateTemp("", "money-gopher-*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	tmp.Close()

	var out, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, p.Path, "-layout", tmp.Name(), "-")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext failed: %w: %s", err, stderr.String())
	}

	return out.String(), nil
}

// PlainText passes text-based files (CSV, plain text) through unchanged.
type PlainText struct{}

func (PlainText) Supports(contentType string) bool {
	return strings.HasPrefix(contentType, "text/") ||
		strings.HasPrefix(contentType, "application/csv")
}

func (PlainText) Extract(_ context.Context, data []byte) (string, error) {
	return string(data), nil
}
