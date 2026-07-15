// Package importer implements the document import pipeline: an uploaded
// file is stored, its text extracted, the bank format detected by
// fingerprint, and the parsed transactions staged for user review. Nothing
// is committed to the books until the user confirms the import via the
// confirmImport GraphQL mutation.
package importer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// Document states as stored in documents.state.
const (
	StateUploaded = "UPLOADED"
	StateParsed   = "PARSED"
	StateFailed   = "FAILED"
	StateSkipped  = "SKIPPED"
	StateImported = "IMPORTED"
)

// ErrSkipped is returned by parsers for documents that are intentionally not
// imported (Vorabpauschale, Storno). The document is marked SKIPPED, not FAILED.
var ErrSkipped = errors.New("skipped")

// Service runs the import pipeline.
type Service struct {
	DB         *persistence.DB
	Extractors []Extractor
	Parsers    []Parser
}

// NewService returns a Service with all built-in extractors and parsers.
func NewService(db *persistence.DB) *Service {
	return &Service{
		DB:         db,
		Extractors: []Extractor{NewPDFToText(), PlainText{}},
		Parsers:    DefaultParsers(),
	}
}

// Upload stores a new document and returns it; call Process afterwards
// (typically asynchronously) to run the pipeline.
func (s *Service) Upload(ctx context.Context, personID, filename, contentType string, data []byte) (*persistence.Document, error) {
	return s.DB.CreateDocument(ctx, persistence.CreateDocumentParams{
		ID:          uuid.NewString(),
		PersonID:    personID,
		Filename:    filename,
		ContentType: contentType,
		Data:        data,
	})
}

// Process runs extraction, detection and parsing for a document and stores
// the staged transactions. Failures are recorded on the document itself, so
// the UI can surface them; the returned error mirrors that.
func (s *Service) Process(ctx context.Context, documentID string, personID string) error {
	doc, err := s.DB.GetDocument(ctx, persistence.GetDocumentParams{ID: documentID, PersonID: personID})
	if err != nil {
		return err
	}

	text, bank, staged, err := s.run(ctx, doc)
	if err != nil {
		state := StateFailed
		if errors.Is(err, ErrSkipped) {
			state = StateSkipped
		}
		_, uerr := s.DB.UpdateDocumentState(ctx, persistence.UpdateDocumentStateParams{
			State:         state,
			DetectedBank:  nullStr(bank),
			ExtractedText: nullStr(text),
			Error:         nullStr(err.Error()),
			ID:            doc.ID,
		})
		return errors.Join(err, uerr)
	}

	// Reprocessing replaces earlier staged rows instead of duplicating them.
	if err := s.DB.DeleteStagedTransactions(ctx, doc.ID); err != nil {
		return err
	}

	for _, tx := range staged {
		securityID := sql.NullString{}
		if tx.ISIN != "" {
			security, err := s.DB.FindSecurityByIdentifier(ctx, persistence.FindSecurityByIdentifierParams{
				Kind:  "ISIN",
				Value: tx.ISIN,
			})
			if err == nil {
				securityID = sql.NullString{String: security.ID, Valid: true}
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}

		_, err = s.DB.CreateStagedTransaction(ctx, persistence.CreateStagedTransactionParams{
			ID:           uuid.NewString(),
			DocumentID:   doc.ID,
			Type:         tx.Type,
			Time:         tx.Time,
			Units:        tx.Units,
			Price:        nullInt64(tx.Price),
			Fees:         tx.Fees,
			Taxes:        tx.Taxes,
			CashDelta:    tx.CashDelta,
			Currency:     tx.Currency,
			SecurityHint: nullStr(tx.SecurityHint),
			Isin:         nullStr(tx.ISIN),
			SecurityID:   securityID,
		})
		if err != nil {
			return err
		}
	}

	// Extract per-document fields from the first staged transaction.
	var settlementIBAN string
	var txDate sql.NullTime
	if len(staged) > 0 {
		settlementIBAN = staged[0].SettlementIBAN
		txDate = sql.NullTime{Time: staged[0].Time, Valid: true}
	}

	_, err = s.DB.UpdateDocumentState(ctx, persistence.UpdateDocumentStateParams{
		State:           StateParsed,
		DetectedBank:    nullStr(bank),
		ExtractedText:   nullStr(text),
		SettlementIban:  nullStr(settlementIBAN),
		TransactionDate: txDate,
		ID:              doc.ID,
	})

	return err
}

// run is the pure pipeline: extract → detect → parse.
func (s *Service) run(ctx context.Context, doc *persistence.Document) (text string, bank string, staged []StagedTransaction, err error) {
	var extractor Extractor
	for _, e := range s.Extractors {
		if e.Supports(doc.ContentType) {
			extractor = e
			break
		}
	}
	if extractor == nil {
		return "", "", nil, fmt.Errorf("unsupported content type %q", doc.ContentType)
	}

	text, err = extractor.Extract(ctx, doc.Data)
	if err != nil {
		return "", "", nil, err
	}

	parser, ok := detect(s.Parsers, text)
	if !ok {
		return text, "", nil, fmt.Errorf("no parser recognizes this document")
	}

	staged, err = parser.Parse(text)
	if err != nil {
		return text, parser.Name(), nil, err
	}
	if len(staged) == 0 {
		return text, parser.Name(), nil, fmt.Errorf("no transactions found in document")
	}

	return text, parser.Name(), staged, nil
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}
