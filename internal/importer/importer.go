// Package importer implements the document import pipeline: an uploaded
// file is stored, its text extracted, the bank format detected by
// fingerprint, and the parsed transactions staged for user review. Nothing
// is committed to the books until the user confirms the import via the
// confirmImport GraphQL mutation.
package importer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	hash := contentHash(data)

	// The same bytes again, for the same person, is the same statement: a
	// second document would end in the same transactions twice.
	existing, err := s.DB.GetDocumentByHash(ctx, persistence.GetDocumentByHashParams{
		PersonID:    personID,
		ContentHash: sql.NullString{String: hash, Valid: true},
	})
	if err == nil {
		return nil, &DuplicateError{Existing: existing}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return s.DB.CreateDocument(ctx, persistence.CreateDocumentParams{
		ID:          uuid.NewString(),
		PersonID:    personID,
		Filename:    filename,
		ContentType: contentType,
		Data:        data,
		ContentHash: sql.NullString{String: hash, Valid: true},
	})
}

// DuplicateError is returned by Upload for a document whose bytes are
// already stored for the person.
type DuplicateError struct {
	Existing *persistence.Document
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("already imported on %s as %s", e.Existing.CreatedAt.Format("2006-01-02"), e.Existing.Filename)
}

// contentHash is the hex SHA-256 of a document's bytes.
func contentHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// BackfillHashes computes the content hash of documents stored before
// hashes existed, so that they take part in duplicate detection. It is
// cheap to call when there is nothing to do; moneyd runs it at startup.
func (s *Service) BackfillHashes(ctx context.Context) (int, error) {
	docs, err := s.DB.ListDocumentsWithoutHash(ctx)
	if err != nil {
		return 0, err
	}
	for _, doc := range docs {
		if err := s.DB.SetDocumentHash(ctx, persistence.SetDocumentHashParams{
			ContentHash: sql.NullString{String: contentHash(doc.Data), Valid: true},
			ID:          doc.ID,
		}); err != nil {
			return 0, err
		}
	}
	return len(docs), nil
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
			security, err := s.DB.FindSecurityByIdentifierVisibleTo(ctx, persistence.FindSecurityByIdentifierVisibleToParams{
				Kind:     "ISIN",
				Value:    tx.ISIN,
				PersonID: sql.NullString{String: personID, Valid: true},
			})
			if err == nil {
				securityID = sql.NullString{String: security.ID, Valid: true}
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}

		_, err = s.DB.CreateStagedTransaction(ctx, persistence.CreateStagedTransactionParams{
			ID:            uuid.NewString(),
			DocumentID:    doc.ID,
			Type:          tx.Type,
			Time:          tx.Time,
			Units:         tx.Units,
			Price:         tx.Price,
			Fees:          tx.Fees,
			Taxes:         tx.Taxes,
			CashDelta:     tx.CashDelta,
			Currency:      tx.Currency,
			SecurityHint:  nullStr(tx.SecurityHint),
			Isin:          nullStr(tx.ISIN),
			SecurityID:    securityID,
			PriceCurrency: nullStr(tx.PriceCurrency),
			FxRate:        tx.FxRate,
			SecurityKind:  nullStr(tx.SecurityKind),
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
