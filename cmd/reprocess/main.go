// reprocess bulk-uploads PDFs from a directory and/or re-runs the import
// pipeline on every non-imported document already in the DB.
//
//	go run ./cmd/reprocess -db money.db [-dir /path/to/pdfs]
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
)

func main() {
	dbPath := flag.String("db", "moneyd-data", "path to the lightsql data directory (must match moneyd's -db; moneyd must not be running, lightsql has no multi-process locking yet)")
	dir := flag.String("dir", "", "directory of PDFs to upload before reprocessing")
	flag.Parse()

	if err := run(*dbPath, *dir); err != nil {
		slog.Error("reprocess failed", "err", err)
		os.Exit(1)
	}
}

func run(dbPath, dir string) error {
	db, err := persistence.OpenDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()

	_, person, err := auth.EnsureDevUser(ctx, db)
	if err != nil {
		return err
	}

	svc := importer.NewService(db)

	// Optional: upload all files from a directory tree first.
	if dir != "" {
		uploaded := 0
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				slog.Warn("read failed", "file", d.Name(), "err", err)
				return nil
			}
			ct := http.DetectContentType(data)
			if _, err := svc.Upload(ctx, person.ID, d.Name(), ct, data); err != nil {
				slog.Warn("upload failed", "file", d.Name(), "err", err)
				return nil
			}
			uploaded++
			return nil
		})
		if err != nil {
			return fmt.Errorf("walk dir: %w", err)
		}
		slog.Info("uploaded", "count", uploaded, "dir", dir)
	}

	docs, err := db.ListDocuments(ctx, person.ID)
	if err != nil {
		return err
	}

	ok, failed, already := 0, 0, 0
	for _, doc := range docs {
		if doc.State == "IMPORTED" {
			already++
			continue
		}
		if err := svc.Process(ctx, doc.ID, person.ID); err != nil {
			slog.Warn("skipped", "file", doc.Filename, "err", err)
			failed++
		} else {
			slog.Info("ok", "file", doc.Filename)
			ok++
		}
	}

	fmt.Printf("done: %d ok, %d failed/skipped, %d already imported\n", ok, failed, already)
	return nil
}
