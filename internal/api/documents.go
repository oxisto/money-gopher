package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// DocumentResolver resolves the Document GraphQL type.
type DocumentResolver struct {
	db  *persistence.DB
	doc *persistence.Document
}

// StagedTransactionResolver resolves the StagedTransaction GraphQL type.
type StagedTransactionResolver struct {
	db *persistence.DB
	tx *persistence.StagedTransaction
}

// Documents resolves Query.documents.
func (r *RootResolver) Documents(ctx context.Context, args struct{ State *string }) ([]*DocumentResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var docs []*persistence.Document
	if args.State != nil {
		docs, err = r.db.ListDocumentsByState(ctx, persistence.ListDocumentsByStateParams{
			PersonID: person.ID,
			State:    *args.State,
		})
	} else {
		docs, err = r.db.ListDocuments(ctx, person.ID)
	}
	if err != nil {
		return nil, err
	}

	resolvers := make([]*DocumentResolver, len(docs))
	for i, doc := range docs {
		resolvers[i] = &DocumentResolver{db: r.db, doc: doc}
	}
	return resolvers, nil
}

// DeleteDocument resolves Mutation.deleteDocument.
func (r *RootResolver) DeleteDocument(ctx context.Context, args struct{ ID graphql.ID }) (graphql.ID, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return "", err
	}

	n, err := r.db.DeleteDocument(ctx, persistence.DeleteDocumentParams{ID: string(args.ID), PersonID: person.ID})
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", fmt.Errorf("document %q not found", args.ID)
	}
	return args.ID, nil
}

// ConfirmImportInput is the input type for Mutation.confirmImport.
type ConfirmImportInput struct {
	PortfolioID   *graphql.ID
	CashAccountID *graphql.ID
	Transactions  []StagedTransactionConfirmInput
}

// StagedTransactionConfirmInput carries per-transaction overrides within a
// confirmImport call.
type StagedTransactionConfirmInput struct {
	ID            graphql.ID
	SecurityID    *graphql.ID
	PortfolioID   *graphql.ID
	CashAccountID *graphql.ID
	CashDelta     *int32
}

// ConfirmImport resolves Mutation.confirmImport. It commits all staged
// transactions from the document as real transactions and marks the document
// as IMPORTED. The global defaults in input apply to every staged transaction;
// per-transaction overrides take precedence.
func (r *RootResolver) ConfirmImport(ctx context.Context, args struct {
	DocumentID graphql.ID
	Input      ConfirmImportInput
}) ([]*TransactionResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	doc, err := r.db.GetDocument(ctx, persistence.GetDocumentParams{
		ID:       string(args.DocumentID),
		PersonID: person.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("document %q not found", args.DocumentID)
	} else if err != nil {
		return nil, err
	}

	staged, err := r.db.ListStagedTransactions(ctx, doc.ID)
	if err != nil {
		return nil, err
	}

	overrides := make(map[string]StagedTransactionConfirmInput, len(args.Input.Transactions))
	for _, ov := range args.Input.Transactions {
		overrides[string(ov.ID)] = ov
	}

	source := "import"
	if doc.DetectedBank.Valid {
		source = "import:" + doc.DetectedBank.String
	}

	var created []*TransactionResolver
	for _, stx := range staged {
		ov := overrides[stx.ID]

		d := txData{
			Type:      stx.Type,
			Time:      stx.Time,
			Units:     stx.Units,
			Fees:      stx.Fees,
			Taxes:     stx.Taxes,
			CashDelta: stx.CashDelta,
			Currency:  stx.Currency,
		}
		if stx.Price.Valid {
			price := stx.Price.Int64
			d.Price = &price
		}

		// Apply portfolio default only for types that use a portfolio; pure cash
		// events (DEPOSIT_CASH, WITHDRAW_CASH, INTEREST, etc.) must not have one.
		if isCashOnlyType(stx.Type) {
			d.PortfolioID = idPtr(ov.PortfolioID)
		} else {
			d.PortfolioID = firstID(ov.PortfolioID, args.Input.PortfolioID)
		}
		d.CashAccountID = firstID(ov.CashAccountID, args.Input.CashAccountID)

		if ov.SecurityID != nil {
			s := string(*ov.SecurityID)
			d.SecurityID = &s
		} else if stx.SecurityID.Valid {
			d.SecurityID = &stx.SecurityID.String
		} else if stx.Isin.Valid && stx.Isin.String != "" {
			// Live ISIN lookup for documents processed before the security existed.
			sec, err := r.db.FindSecurityByIdentifier(ctx, persistence.FindSecurityByIdentifierParams{
				Kind:  "ISIN",
				Value: stx.Isin.String,
			})
			if err == nil {
				d.SecurityID = &sec.ID
			} else if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
		}

		// Derive cashDelta when not staged and not overridden.
		if ov.CashDelta != nil {
			d.CashDelta = int64(*ov.CashDelta)
		} else if stx.CashDelta == 0 {
			if delta, ok := deriveCashDelta(&d); ok {
				d.CashDelta = delta
			}
		}

		if err := d.validate(ctx, r.db, person); err != nil {
			return nil, fmt.Errorf("staged transaction %q: %w", stx.ID, err)
		}

		tx, err := r.db.CreateTransaction(ctx, persistence.CreateTransactionParams{
			ID:            uuid.NewString(),
			Type:          d.Type,
			Time:          d.Time,
			PortfolioID:   nullString(d.PortfolioID),
			SecurityID:    nullString(d.SecurityID),
			CashAccountID: nullString(d.CashAccountID),
			Units:         d.Units,
			Price:         nullInt64(d.Price),
			Fees:          d.Fees,
			Taxes:         d.Taxes,
			CashDelta:     d.CashDelta,
			Currency:      d.Currency,
			Source:        source,
		})
		if err != nil {
			return nil, err
		}

		created = append(created, &TransactionResolver{db: r.db, transaction: tx})
	}

	if _, err := r.db.UpdateDocumentState(ctx, persistence.UpdateDocumentStateParams{
		State:           "IMPORTED",
		DetectedBank:    doc.DetectedBank,
		ExtractedText:   doc.ExtractedText,
		SettlementIban:  doc.SettlementIban,
		TransactionDate: doc.TransactionDate,
		ID:              doc.ID,
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// firstID returns a *string from the first non-nil graphql.ID pointer.
func firstID(ids ...*graphql.ID) *string {
	for _, id := range ids {
		if id != nil {
			s := string(*id)
			return &s
		}
	}
	return nil
}

// isCashOnlyType reports whether a transaction type is a pure cash event (no
// portfolio or security). These types must not receive a portfolioID default.
func isCashOnlyType(typ string) bool {
	switch typ {
	case "INTEREST", "DEPOSIT_CASH", "WITHDRAW_CASH", "ACCOUNT_FEES", "TAX_REFUND":
		return true
	}
	return false
}

// Document field resolvers.

func (r *DocumentResolver) ID() graphql.ID      { return graphql.ID(r.doc.ID) }
func (r *DocumentResolver) Filename() string    { return r.doc.Filename }
func (r *DocumentResolver) ContentType() string { return r.doc.ContentType }
func (r *DocumentResolver) State() string       { return r.doc.State }

func (r *DocumentResolver) DetectedBank() *string {
	if !r.doc.DetectedBank.Valid {
		return nil
	}
	return &r.doc.DetectedBank.String
}

func (r *DocumentResolver) Error() *string {
	if !r.doc.Error.Valid {
		return nil
	}
	return &r.doc.Error.String
}

func (r *DocumentResolver) CreatedAt() graphql.Time {
	return graphql.Time{Time: r.doc.CreatedAt}
}

func (r *DocumentResolver) TransactionDate() *graphql.Time {
	if !r.doc.TransactionDate.Valid {
		return nil
	}
	return &graphql.Time{Time: r.doc.TransactionDate.Time}
}

func (r *DocumentResolver) StagedTransactions(ctx context.Context) ([]*StagedTransactionResolver, error) {
	txs, err := r.db.ListStagedTransactions(ctx, r.doc.ID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*StagedTransactionResolver, len(txs))
	for i, tx := range txs {
		resolvers[i] = &StagedTransactionResolver{db: r.db, tx: tx}
	}
	return resolvers, nil
}

// SuggestedCashAccount resolves Document.suggestedCashAccount by looking up
// the settlement IBAN extracted from the document against all cash accounts
// belonging to the current user.
func (r *DocumentResolver) SuggestedCashAccount(ctx context.Context) (*CashAccountResolver, error) {
	if !r.doc.SettlementIban.Valid || r.doc.SettlementIban.String == "" {
		return nil, nil
	}
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	account, err := r.db.GetCashAccountByIBAN(ctx, persistence.GetCashAccountByIBANParams{
		Iban:     sql.NullString{String: r.doc.SettlementIban.String, Valid: true},
		PersonID: person.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &CashAccountResolver{db: r.db, account: account}, nil
}

// StagedTransaction field resolvers.

func (r *StagedTransactionResolver) ID() graphql.ID { return graphql.ID(r.tx.ID) }
func (r *StagedTransactionResolver) Type() string   { return r.tx.Type }

func (r *StagedTransactionResolver) Time() graphql.Time {
	return graphql.Time{Time: r.tx.Time}
}

func (r *StagedTransactionResolver) Units() float64 { return r.tx.Units }

func (r *StagedTransactionResolver) Price() *Money {
	if !r.tx.Price.Valid {
		return nil
	}
	m := money(r.tx.Price.Int64, r.tx.Currency)
	return &m
}

func (r *StagedTransactionResolver) Fees() Money {
	return money(r.tx.Fees, r.tx.Currency)
}

func (r *StagedTransactionResolver) Taxes() Money {
	return money(r.tx.Taxes, r.tx.Currency)
}

func (r *StagedTransactionResolver) CashDelta() Money {
	return money(r.tx.CashDelta, r.tx.Currency)
}

func (r *StagedTransactionResolver) Currency() string { return r.tx.Currency }

func (r *StagedTransactionResolver) SecurityHint() *string {
	if !r.tx.SecurityHint.Valid {
		return nil
	}
	return &r.tx.SecurityHint.String
}

func (r *StagedTransactionResolver) Isin() *string {
	if !r.tx.Isin.Valid {
		return nil
	}
	return &r.tx.Isin.String
}

func (r *StagedTransactionResolver) Security(ctx context.Context) (*SecurityResolver, error) {
	if r.tx.SecurityID.Valid {
		security, err := r.db.GetSecurity(ctx, r.tx.SecurityID.String)
		if err != nil {
			return nil, err
		}
		return &SecurityResolver{db: r.db, security: security}, nil
	}
	// Fall back to a live ISIN lookup so that securities created after this
	// document was processed are still matched.
	if !r.tx.Isin.Valid || r.tx.Isin.String == "" {
		return nil, nil
	}
	security, err := r.db.FindSecurityByIdentifier(ctx, persistence.FindSecurityByIdentifierParams{
		Kind:  "ISIN",
		Value: r.tx.Isin.String,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &SecurityResolver{db: r.db, security: security}, nil
}
