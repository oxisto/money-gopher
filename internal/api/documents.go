package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
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

// ReprocessDocument resolves Mutation.reprocessDocument. Re-runs extraction
// and parsing with the current importer code, replacing the document's
// staged transactions — never its already-committed ones; see
// Document.transactions.
func (r *RootResolver) ReprocessDocument(ctx context.Context, args struct{ ID graphql.ID }) (*DocumentResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	svc := importer.NewService(r.db)
	if err := svc.Process(ctx, string(args.ID), person.ID); err != nil {
		return nil, err
	}

	doc, err := r.db.GetDocument(ctx, persistence.GetDocumentParams{ID: string(args.ID), PersonID: person.ID})
	if err != nil {
		return nil, err
	}
	return &DocumentResolver{db: r.db, doc: doc}, nil
}

// ConfirmImportInput is the input type for Mutation.confirmImport.
type ConfirmImportInput struct {
	PortfolioID               *graphql.ID
	CashAccountID             *graphql.ID
	DepositsFromCashAccountID *graphql.ID
	Transactions              []StagedTransactionConfirmInput
}

// StagedTransactionConfirmInput carries per-transaction overrides within a
// confirmImport call.
type StagedTransactionConfirmInput struct {
	ID                        graphql.ID
	AssetID                   *graphql.ID
	PortfolioID               *graphql.ID
	CashAccountID             *graphql.ID
	CashDelta                 *int32
	TransferFromCashAccountID *graphql.ID
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

	// Everything from here on commits together. Without this, a failure
	// partway through left some transactions committed while the document
	// stayed STAGED. The resolvers keep r.db, not q: they outlive the
	// transaction and read through the pool.
	var created []*TransactionResolver
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		created = nil // Tx retries on write conflicts
		// Re-processing (see reprocessDocument) regenerates staged transactions
		// for a document that may already have committed ones from an earlier
		// confirm — reprocessDocument never touches those. Matching each staged
		// row against them here (by type/time/units/cashDelta, i.e. "the same
		// real-world event") lets confirming safely reconcile a transaction
		// in place — picking up whatever the newer parser now gets right, e.g.
		// FX precision or a security match — instead of creating a duplicate.
		existingForDoc, err := q.ListTransactionsByDocument(ctx, sql.NullString{String: doc.ID, Valid: true})
		if err != nil {
			return err
		}
		consumed := make(map[string]bool, len(existingForDoc))
		matchExisting := func(d *txData) *persistence.Transaction {
			for _, ex := range existingForDoc {
				if consumed[ex.ID] {
					continue
				}
				if ex.Type == d.Type && ex.Time.Equal(d.Time) && ex.Units == d.Units && ex.CashDelta == d.CashDelta {
					consumed[ex.ID] = true
					return ex
				}
			}
			return nil
		}

		overrides := make(map[string]StagedTransactionConfirmInput, len(args.Input.Transactions))
		for _, ov := range args.Input.Transactions {
			overrides[string(ov.ID)] = ov
		}

		source := "import"
		if doc.DetectedBank.Valid {
			source = "import:" + doc.DetectedBank.String
		}

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
			d.Price = stx.Price
			if stx.PriceCurrency.Valid {
				d.PriceCurrency = stx.PriceCurrency.String
			}
			d.FxRate = stx.FxRate

			// Apply portfolio default only for types that use a portfolio; pure cash
			// events (DEPOSIT_CASH, WITHDRAW_CASH, INTEREST, etc.) must not have one.
			if isCashOnlyType(stx.Type) {
				d.PortfolioID = idPtr(ov.PortfolioID)
			} else {
				d.PortfolioID = firstID(ov.PortfolioID, args.Input.PortfolioID)
			}
			d.CashAccountID = firstID(ov.CashAccountID, args.Input.CashAccountID)

			if ov.AssetID != nil {
				s := string(*ov.AssetID)
				d.SecurityID = &s
			} else if stx.SecurityID.Valid {
				d.SecurityID = &stx.SecurityID.String
			} else if stx.Isin.Valid && stx.Isin.String != "" {
				// Live ISIN lookup for documents processed before the security existed.
				sec, err := q.FindSecurityByIdentifierVisibleTo(ctx, persistence.FindSecurityByIdentifierVisibleToParams{
					Kind:     "ISIN",
					Value:    stx.Isin.String,
					PersonID: ownedBy(person),
				})
				if err == nil {
					d.SecurityID = &sec.ID
				} else if !errors.Is(err, sql.ErrNoRows) {
					return err
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
				return fmt.Errorf("staged transaction %q: %w", stx.ID, err)
			}

			// A deposit "paid from" one of the person's accounts is a
			// transfer: the deposit here plus a withdrawal there, linked.
			var leg *txData
			if from := firstID(ov.TransferFromCashAccountID, args.Input.DepositsFromCashAccountID); from != nil && d.Type == "DEPOSIT_CASH" {
				if d.CashAccountID != nil && *from == *d.CashAccountID {
					return fmt.Errorf("staged transaction %q: a deposit cannot come from the account it lands on", stx.ID)
				}
				leg = &txData{Type: "WITHDRAW_CASH", Time: d.Time, CashAccountID: from, CashDelta: -d.CashDelta, Currency: d.Currency}
				if err := leg.validate(ctx, r.db, person); err != nil {
					return fmt.Errorf("staged transaction %q, paid from: %w", stx.ID, err)
				}
			}

			// A statement row a schedule already booked (same account,
			// type and amount, within a week) confirms that booking: the
			// bank's date and amounts win, nothing is created twice.
			if d.CashAccountID != nil {
				candidates, err := q.ListUnconfirmedScheduledTransactions(ctx, persistence.ListUnconfirmedScheduledTransactionsParams{
					CashAccountID: nullString(d.CashAccountID),
					Type:          d.Type,
				})
				if err != nil {
					return err
				}
				var hit *persistence.Transaction
				for _, c := range candidates {
					if consumed[c.ID] || c.CashDelta != d.CashDelta {
						continue
					}
					if diff := c.Time.Sub(d.Time); diff < -7*24*time.Hour || diff > 7*24*time.Hour {
						continue
					}
					hit = c
					break
				}
				if hit != nil {
					consumed[hit.ID] = true
					tx, err := q.ConfirmScheduledTransaction(ctx, persistence.ConfirmScheduledTransactionParams{
						Time:       d.Time,
						Fees:       d.Fees,
						Taxes:      d.Taxes,
						CashDelta:  d.CashDelta,
						DocumentID: sql.NullString{String: doc.ID, Valid: true},
						Source:     source,
						ID:         hit.ID,
					})
					if err != nil {
						return err
					}
					if err := mirrorOntoCounterpart(ctx, q, hit, d.Time, -d.CashDelta); err != nil {
						return err
					}
					created = append(created, &TransactionResolver{db: r.db, transaction: tx})
					continue
				}
			}

			if match := matchExisting(&d); match != nil {
				tx, err := q.UpdateTransactionFromImport(ctx, persistence.UpdateTransactionFromImportParams{
					Type:          d.Type,
					Time:          d.Time,
					PortfolioID:   nullString(d.PortfolioID),
					SecurityID:    nullString(d.SecurityID),
					CashAccountID: nullString(d.CashAccountID),
					Units:         d.Units,
					Price:         d.Price,
					Fees:          d.Fees,
					Taxes:         d.Taxes,
					CashDelta:     d.CashDelta,
					Currency:      d.Currency,
					PriceCurrency: sql.NullString{String: d.PriceCurrency, Valid: d.PriceCurrency != ""},
					FxRate:        d.FxRate,
					ID:            match.ID,
				})
				if err != nil {
					return err
				}
				// A re-confirm keeps a transfer's other leg in step; it
				// does not turn a plain deposit into a transfer after the
				// fact (delete and import again for that).
				if leg != nil {
					if err := mirrorOntoCounterpart(ctx, q, match, leg.Time, leg.CashDelta); err != nil {
						return err
					}
				}
				created = append(created, &TransactionResolver{db: r.db, transaction: tx})
				continue
			}

			var transferID sql.NullString
			if leg != nil {
				transferID = sql.NullString{String: uuid.NewString(), Valid: true}
			}
			tx, err := q.CreateTransaction(ctx, persistence.CreateTransactionParams{
				ID:            uuid.NewString(),
				TransferID:    transferID,
				Type:          d.Type,
				Time:          d.Time,
				PortfolioID:   nullString(d.PortfolioID),
				SecurityID:    nullString(d.SecurityID),
				CashAccountID: nullString(d.CashAccountID),
				Units:         d.Units,
				Price:         d.Price,
				Fees:          d.Fees,
				Taxes:         d.Taxes,
				CashDelta:     d.CashDelta,
				Currency:      d.Currency,
				Source:        source,
				DocumentID:    sql.NullString{String: doc.ID, Valid: true},
				PriceCurrency: sql.NullString{String: d.PriceCurrency, Valid: d.PriceCurrency != ""},
				FxRate:        d.FxRate,
				Note:          stx.Note,
			})
			if err != nil {
				return err
			}
			if leg != nil {
				if _, err := q.CreateTransaction(ctx, persistence.CreateTransactionParams{
					ID:            uuid.NewString(),
					TransferID:    transferID,
					Type:          leg.Type,
					Time:          leg.Time,
					CashAccountID: nullString(leg.CashAccountID),
					CashDelta:     leg.CashDelta,
					Currency:      leg.Currency,
					Source:        source,
					DocumentID:    sql.NullString{String: doc.ID, Valid: true},
				}); err != nil {
					return err
				}
			}

			created = append(created, &TransactionResolver{db: r.db, transaction: tx})
		}

		if _, err := q.UpdateDocumentState(ctx, persistence.UpdateDocumentStateParams{
			State:           "IMPORTED",
			DetectedBank:    doc.DetectedBank,
			ExtractedText:   doc.ExtractedText,
			SettlementIban:  doc.SettlementIban,
			TransactionDate: doc.TransactionDate,
			ID:              doc.ID,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
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

// mirrorOntoCounterpart moves a transfer's other leg to the given time and
// amount, when tx is a leg; a plain transaction is left alone.
func mirrorOntoCounterpart(ctx context.Context, q *persistence.Queries, tx *persistence.Transaction, at time.Time, cashDelta int64) error {
	if !tx.TransferID.Valid {
		return nil
	}
	other, err := q.GetTransferCounterpart(ctx, persistence.GetTransferCounterpartParams{TransferID: tx.TransferID, ID: tx.ID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		return err
	}
	_, err = q.UpdateTransactionFromImport(ctx, persistence.UpdateTransactionFromImportParams{
		Type: other.Type, Time: at, CashAccountID: other.CashAccountID,
		CashDelta: cashDelta, Currency: other.Currency, Price: other.Price, FxRate: other.FxRate,
		PriceCurrency: other.PriceCurrency, Units: other.Units, Fees: other.Fees, Taxes: other.Taxes,
		PortfolioID: other.PortfolioID, SecurityID: other.SecurityID, ID: other.ID,
	})
	return err
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
		Iban:     sql.NullString{String: persistence.NormalizeIBAN(r.doc.SettlementIban.String), Valid: true},
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

// Transactions resolves Document.transactions — already-committed
// transactions that came from this document; see the schema doc comment
// for why this matters before re-confirming a reprocessed document.
func (r *DocumentResolver) Transactions(ctx context.Context) ([]*TransactionResolver, error) {
	transactions, err := r.db.ListTransactionsByDocument(ctx, sql.NullString{String: r.doc.ID, Valid: true})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*TransactionResolver, 0, len(transactions))
	for _, t := range transactions {
		resolvers = append(resolvers, &TransactionResolver{db: r.db, transaction: t})
	}
	return resolvers, nil
}

// StagedTransaction field resolvers.

func (r *StagedTransactionResolver) ID() graphql.ID { return graphql.ID(r.tx.ID) }
func (r *StagedTransactionResolver) Type() string   { return r.tx.Type }

func (r *StagedTransactionResolver) Time() graphql.Time {
	return graphql.Time{Time: r.tx.Time}
}

func (r *StagedTransactionResolver) Units() float64 { return r.tx.Units }

func (r *StagedTransactionResolver) Price() *Price {
	if !r.tx.Price.Valid {
		return nil
	}
	p := price(r.tx.Price, priceCurrencyOf(r.tx.PriceCurrency, r.tx.Currency))
	return &p
}

// FxRate resolves StagedTransaction.fxRate — 1 unit of Currency expressed
// in the Price's own currency, when they differ (see the schema doc
// comment on price_currency in migrations/0005_fx_rate.sql).
func (r *StagedTransactionResolver) FxRate() *Decimal {
	if !r.tx.FxRate.Valid {
		return nil
	}
	d := decimal(r.tx.FxRate)
	return &d
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

// CalculatedCashDelta resolves StagedTransaction.calculatedCashDelta — see
// its schema doc comment for why this is worth showing next to the parsed
// CashDelta. Only meaningful when there's an actual per-unit Price to
// reconcile against: deriveCashDelta's formula is gross = price × units,
// and a cash-only DIVIDEND/INTEREST document (no per-share price parsed at
// all) would otherwise show a bogus "mismatch" against every single one.
func (r *StagedTransactionResolver) CalculatedCashDelta() *Money {
	if !r.tx.Price.Valid {
		return nil
	}
	delta, ok := deriveCashDelta(&txData{
		Type:          r.tx.Type,
		Units:         r.tx.Units,
		Price:         r.tx.Price,
		Fees:          r.tx.Fees,
		Taxes:         r.tx.Taxes,
		PriceCurrency: priceCurrencyOf(r.tx.PriceCurrency, ""),
		FxRate:        r.tx.FxRate,
	})
	if !ok {
		return nil
	}
	m := money(delta, r.tx.Currency)
	return &m
}

func (r *StagedTransactionResolver) Currency() string { return r.tx.Currency }

func (r *StagedTransactionResolver) AssetHint() *string {
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

func (r *StagedTransactionResolver) Note() *string {
	if !r.tx.Note.Valid {
		return nil
	}
	return &r.tx.Note.String
}

func (r *StagedTransactionResolver) AccountHint() *string {
	if !r.tx.AccountHint.Valid {
		return nil
	}
	return &r.tx.AccountHint.String
}

func (r *StagedTransactionResolver) AssetKindHint() *string {
	if !r.tx.SecurityKind.Valid {
		return nil
	}
	return &r.tx.SecurityKind.String
}

func (r *StagedTransactionResolver) Asset(ctx context.Context) (*AssetResolver, error) {
	if r.tx.SecurityID.Valid {
		security, err := r.db.GetSecurity(ctx, r.tx.SecurityID.String)
		if err != nil {
			return nil, err
		}
		return &AssetResolver{db: r.db, security: security}, nil
	}
	// Fall back to a live ISIN lookup so that securities created after this
	// document was processed are still matched.
	if !r.tx.Isin.Valid || r.tx.Isin.String == "" {
		return nil, nil
	}
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	security, err := r.db.FindSecurityByIdentifierVisibleTo(ctx, persistence.FindSecurityByIdentifierVisibleToParams{
		Kind:     "ISIN",
		Value:    r.tx.Isin.String,
		PersonID: ownedBy(person),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &AssetResolver{db: r.db, security: security}, nil
}
