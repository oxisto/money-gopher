package importer

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

const sampleCSV = `Date,Type,Security,ISIN,Units,Price,Fees,Taxes,CashDelta,Currency
2026-01-03,BUY,iShares Core MSCI World,IE00B4L5Y983,10,100.50,1.50,0,,EUR
2026-02-01,DIVIDEND,iShares Core MSCI World,IE00B4L5Y983,,,0,3.25,12.00,EUR
2026-03-01,DEPOSIT_CASH,,,,,0,0,500.00,EUR
`

func newTestService(t *testing.T) (*Service, string, context.Context) {
	t.Helper()

	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, person, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatalf("EnsureDevUser() error = %v", err)
	}

	return NewService(db), person.ID, context.Background()
}

func TestCSV_Parse(t *testing.T) {
	if !(CSV{}).Matches(sampleCSV) {
		t.Fatal("CSV fingerprint did not match its own header")
	}
	if (CSV{}).Matches("Kontoauszug Nr. 7\nsome bank prose") {
		t.Fatal("CSV fingerprint matched a non-CSV document")
	}

	staged, err := (CSV{}).Parse(sampleCSV)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(staged) != 3 {
		t.Fatalf("staged = %d, want 3", len(staged))
	}

	buy := staged[0]
	if buy.Type != "BUY" || buy.Units != 10 || buy.Price.StringPrec(2) != "100.50" || buy.Fees != 150 {
		t.Errorf("buy = %+v", buy)
	}
	if buy.ISIN != "IE00B4L5Y983" || buy.SecurityHint != "iShares Core MSCI World" {
		t.Errorf("buy hints = %+v", buy)
	}

	dividend := staged[1]
	if dividend.Type != "DIVIDEND" || dividend.CashDelta != 1200 || dividend.Taxes != 325 {
		t.Errorf("dividend = %+v", dividend)
	}

	deposit := staged[2]
	if deposit.Type != "DEPOSIT_CASH" || deposit.CashDelta != 50000 || deposit.ISIN != "" {
		t.Errorf("deposit = %+v", deposit)
	}
}

// TestService_Upload_Duplicate: the same bytes twice is the same statement,
// whatever the file is called; another person may still have their own
// copy, and documents stored before hashes existed join in once hashed.
func TestService_Upload_Duplicate(t *testing.T) {
	svc, personID, ctx := newTestService(t)

	first, err := svc.Upload(ctx, personID, "statement.pdf", "application/pdf", []byte("%PDF-1.4 the statement"))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	_, err = svc.Upload(ctx, personID, "statement (1).pdf", "application/pdf", []byte("%PDF-1.4 the statement"))
	var duplicate *DuplicateError
	if !errors.As(err, &duplicate) {
		t.Fatalf("second Upload() error = %v, want a DuplicateError", err)
	}
	if duplicate.Existing.ID != first.ID {
		t.Errorf("duplicate of %s, want %s", duplicate.Existing.ID, first.ID)
	}
	if msg := duplicate.Error(); !strings.Contains(msg, "statement.pdf") {
		t.Errorf("message %q should name the earlier file", msg)
	}

	// A different statement is fine.
	if _, err := svc.Upload(ctx, personID, "statement.pdf", "application/pdf", []byte("%PDF-1.4 another statement")); err != nil {
		t.Errorf("Upload() of other content error = %v", err)
	}

	// Another person's copy of the same statement is theirs to import.
	other, err := svc.DB.CreatePerson(ctx, persistence.CreatePersonParams{ID: "other", DisplayName: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Upload(ctx, other.ID, "statement.pdf", "application/pdf", []byte("%PDF-1.4 the statement")); err != nil {
		t.Errorf("Upload() for another person error = %v", err)
	}

	// A document from before hashes existed is caught once backfilled.
	if _, err := svc.DB.CreateDocument(ctx, persistence.CreateDocumentParams{
		ID: "old", PersonID: personID, Filename: "old.pdf", ContentType: "application/pdf", Data: []byte("%PDF-1.4 an old one"),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Upload(ctx, personID, "old again.pdf", "application/pdf", []byte("%PDF-1.4 an old one")); err != nil {
		t.Fatalf("Upload() before backfill error = %v, want none (not hashed yet)", err)
	}
	if n, err := svc.BackfillHashes(ctx); err != nil || n != 1 {
		t.Fatalf("BackfillHashes() = %d, %v; want 1, nil", n, err)
	}
	if _, err := svc.Upload(ctx, personID, "old once more.pdf", "application/pdf", []byte("%PDF-1.4 an old one")); !errors.As(err, &duplicate) {
		t.Errorf("Upload() after backfill error = %v, want a DuplicateError", err)
	}
}

func TestService_Process(t *testing.T) {
	svc, userID, ctx := newTestService(t)

	// A security with the matching ISIN already exists; the pipeline should
	// link it.
	security, err := svc.DB.CreateSecurity(ctx, persistence.CreateSecurityParams{
		ID: "sec-msci", DisplayName: "iShares Core MSCI World",
	})
	if err != nil {
		t.Fatalf("CreateSecurity() error = %v", err)
	}
	err = svc.DB.CreateSecurityIdentifier(ctx, persistence.CreateSecurityIdentifierParams{
		SecurityID: security.ID, Kind: "ISIN", Value: "IE00B4L5Y983",
	})
	if err != nil {
		t.Fatalf("CreateSecurityIdentifier() error = %v", err)
	}

	doc, err := svc.Upload(ctx, userID, "transactions.csv", "text/csv", []byte(sampleCSV))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if doc.State != StateUploaded {
		t.Errorf("state after upload = %q, want %q", doc.State, StateUploaded)
	}

	if err := svc.Process(ctx, doc.ID, userID); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	processed, err := svc.DB.GetDocument(ctx, persistence.GetDocumentParams{ID: doc.ID, PersonID: userID})
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if processed.State != StateParsed {
		t.Fatalf("state = %q (error: %s), want %q", processed.State, processed.Error.String, StateParsed)
	}
	if processed.DetectedBank.String != "csv" {
		t.Errorf("detected bank = %q, want csv", processed.DetectedBank.String)
	}

	staged, err := svc.DB.ListStagedTransactions(ctx, doc.ID)
	if err != nil {
		t.Fatalf("ListStagedTransactions() error = %v", err)
	}
	if len(staged) != 3 {
		t.Fatalf("staged rows = %d, want 3", len(staged))
	}
	if got := staged[0].SecurityID.String; got != security.ID {
		t.Errorf("matched security = %q, want %q", got, security.ID)
	}

	// Reprocessing must not duplicate the staged rows.
	if err := svc.Process(ctx, doc.ID, userID); err != nil {
		t.Fatalf("second Process() error = %v", err)
	}
	staged, _ = svc.DB.ListStagedTransactions(ctx, doc.ID)
	if len(staged) != 3 {
		t.Errorf("staged rows after reprocess = %d, want 3", len(staged))
	}
}

func TestService_ProcessFailure(t *testing.T) {
	svc, userID, ctx := newTestService(t)

	// A text document no parser understands must fail visibly, not silently.
	doc, err := svc.Upload(ctx, userID, "notes.txt", "text/plain", []byte("Dear diary, today I bought stocks."))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	if err := svc.Process(ctx, doc.ID, userID); err == nil {
		t.Fatal("Process() succeeded for garbage input")
	}

	processed, _ := svc.DB.GetDocument(ctx, persistence.GetDocumentParams{ID: doc.ID, PersonID: userID})
	if processed.State != StateFailed {
		t.Errorf("state = %q, want %q", processed.State, StateFailed)
	}
	if !strings.Contains(processed.Error.String, "no parser") {
		t.Errorf("error = %q, want a 'no parser' message", processed.Error.String)
	}
}
