package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/finance"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// TransactionResolver resolves the Transaction GraphQL type.
type TransactionResolver struct {
	db          *persistence.DB
	transaction *persistence.Transaction
}

// txData is the merged, validated set of transaction fields shared by the
// create and update mutations.
type txData struct {
	Type          string
	Time          time.Time
	PortfolioID   *string
	SecurityID    *string
	CashAccountID *string
	Units         float64
	Price         persistence.Decimal
	Fees          int64
	Taxes         int64
	CashDelta     int64
	Currency      string
	// PriceCurrency and FxRate: see StagedTransaction's doc comments in
	// internal/importer/parser.go — same meaning here.
	PriceCurrency string
	FxRate        persistence.Decimal
}

// deriveCashDelta computes the signed cash effect for transaction types where
// it follows from units, price, fees and taxes. The second return value is
// false for pure cash events, whose delta must be provided explicitly.
func deriveCashDelta(d *txData) (int64, bool) {
	// d.Price is major units, in PriceCurrency when set or Currency
	// otherwise; gross converts it to Currency minor units (cents), like
	// CashDelta, Fees and Taxes already are.
	gross := d.Price.MulFloat(d.Units)
	if d.FxRate.Valid {
		gross = gross.Div(d.FxRate)
	}
	grossMinor := gross.MulInt(100).RoundToInt64()

	switch d.Type {
	case "BUY":
		return -(grossMinor + d.Fees + d.Taxes), true
	case "SELL", "DIVIDEND":
		return grossMinor - d.Fees - d.Taxes, true
	case "DELIVERY_INBOUND", "DELIVERY_OUTBOUND":
		return 0, true
	default:
		return 0, false
	}
}

// validate checks the internal consistency of a transaction and that all
// referenced entities exist and belong to the user.
func (d *txData) validate(ctx context.Context, db *persistence.DB, person *persistence.Person) error {
	switch d.Type {
	case "BUY", "SELL", "DIVIDEND":
		if d.PortfolioID == nil || d.SecurityID == nil {
			return fmt.Errorf("%s requires a portfolio and a security", d.Type)
		}
		if d.CashAccountID == nil {
			return fmt.Errorf("%s requires a cash account to settle against", d.Type)
		}
		if d.Type != "DIVIDEND" && (d.Units <= 0 || !d.Price.Valid) {
			return fmt.Errorf("%s requires positive units and a price", d.Type)
		}
	case "DELIVERY_INBOUND", "DELIVERY_OUTBOUND":
		if d.PortfolioID == nil || d.SecurityID == nil {
			return fmt.Errorf("%s requires a portfolio and a security", d.Type)
		}
		if d.Units <= 0 {
			return fmt.Errorf("%s requires positive units", d.Type)
		}
	case "INTEREST", "DEPOSIT_CASH", "WITHDRAW_CASH", "ACCOUNT_FEES", "TAX_REFUND":
		if d.CashAccountID == nil {
			return fmt.Errorf("%s requires a cash account", d.Type)
		}
		if d.PortfolioID != nil || d.SecurityID != nil {
			return fmt.Errorf("%s is a pure cash event and cannot reference a portfolio or security", d.Type)
		}
	}

	switch d.Type {
	case "DEPOSIT_CASH":
		if d.CashDelta <= 0 {
			return fmt.Errorf("%s must increase the cash balance", d.Type)
		}
	case "WITHDRAW_CASH", "ACCOUNT_FEES":
		if d.CashDelta >= 0 {
			return fmt.Errorf("%s must decrease the cash balance", d.Type)
		}
	}
	// TAX_REFUND deliberately allows either sign: a retroactive
	// capital-gains-tax adjustment (a loss offset applied after the fact)
	// can net to an additional small charge as easily as a refund — see
	// ingKontoTaxLabel's callers in ing_konto.go for a real example.
	if d.Type == "TAX_REFUND" && d.CashDelta == 0 {
		return fmt.Errorf("%s must not be zero", d.Type)
	}

	if d.PortfolioID != nil {
		_, err := db.GetPortfolio(ctx, persistence.GetPortfolioParams{ID: *d.PortfolioID, PersonID: person.ID})
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("portfolio %q not found", *d.PortfolioID)
		} else if err != nil {
			return err
		}
	}
	if d.CashAccountID != nil {
		account, err := db.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: *d.CashAccountID, PersonID: person.ID})
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("cash account %q not found", *d.CashAccountID)
		} else if err != nil {
			return err
		}
		// Interest is earned on a cash or savings account and charged on a
		// loan, where it is what makes the balance more negative.
		if d.Type == "INTEREST" {
			if kindOf(account) == "LOAN" && d.CashDelta >= 0 {
				return errors.New("INTEREST on a loan is interest charged and must decrease the balance")
			}
			if kindOf(account) != "LOAN" && d.CashDelta <= 0 {
				return errors.New("INTEREST must increase the cash balance")
			}
		}
	}
	if d.SecurityID != nil {
		_, err := db.GetSecurityVisibleTo(ctx, persistence.GetSecurityVisibleToParams{ID: *d.SecurityID, PersonID: ownedBy(person)})
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("security %q not found", *d.SecurityID)
		} else if err != nil {
			return err
		}
	}

	return nil
}

// getOwnedTransaction fetches a transaction and verifies that it belongs to
// the user via its portfolio or cash account. It returns sql.ErrNoRows for
// both a missing and a foreign transaction, so callers cannot tell the two
// apart.
func (r *RootResolver) getOwnedTransaction(ctx context.Context, person *persistence.Person, id string) (*persistence.Transaction, error) {
	transaction, err := r.db.GetTransaction(ctx, id)
	if err != nil {
		return nil, err
	}

	if transaction.PortfolioID.Valid {
		_, err = r.db.GetPortfolio(ctx, persistence.GetPortfolioParams{
			ID:       transaction.PortfolioID.String,
			PersonID: person.ID,
		})
	} else {
		_, err = r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
			ID:       transaction.CashAccountID.String,
			PersonID: person.ID,
		})
	}
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// Transaction resolves Query.transaction.
func (r *RootResolver) Transaction(ctx context.Context, args struct{ ID graphql.ID }) (*TransactionResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	transaction, err := r.getOwnedTransaction(ctx, person, string(args.ID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &TransactionResolver{db: r.db, transaction: transaction}, nil
}

// CreateTransactionInput is the input for Mutation.createTransaction.
type CreateTransactionInput struct {
	Type          string
	Time          graphql.Time
	PortfolioID   *graphql.ID
	AssetID       *graphql.ID
	CashAccountID *graphql.ID
	Units         *float64
	Price         *Decimal
	Fees          *int32
	Taxes         *int32
	CashDelta     *int32
	Currency      string
	Note          *string
}

// CreateTransaction resolves Mutation.createTransaction.
func (r *RootResolver) CreateTransaction(ctx context.Context, args struct{ Input CreateTransactionInput }) (*TransactionResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	in := args.Input
	d := txData{
		Type: in.Type,
		// Times are stored in UTC so that their string representation in
		// SQLite always sorts chronologically.
		Time:          in.Time.Time.UTC(),
		PortfolioID:   idPtr(in.PortfolioID),
		SecurityID:    idPtr(in.AssetID),
		CashAccountID: idPtr(in.CashAccountID),
		Fees:          int64Value(in.Fees),
		Taxes:         int64Value(in.Taxes),
		Currency:      in.Currency,
	}
	if in.Units != nil {
		d.Units = *in.Units
	}
	if in.Price != nil {
		d.Price = persistence.Decimal(*in.Price)
	}

	if in.CashDelta != nil {
		d.CashDelta = int64(*in.CashDelta)
	} else {
		delta, ok := deriveCashDelta(&d)
		if !ok {
			return nil, fmt.Errorf("%s requires an explicit cashDelta", d.Type)
		}
		d.CashDelta = delta
	}

	if err := d.validate(ctx, r.db, person); err != nil {
		return nil, err
	}

	transaction, err := r.db.CreateTransaction(ctx, persistence.CreateTransactionParams{
		ID:            uuid.NewString(),
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
		Source:        "MANUAL",
		Note:          noteColumn(ptrStr(args.Input.Note)),
	})
	if err != nil {
		return nil, err
	}

	return &TransactionResolver{db: r.db, transaction: transaction}, nil
}

// UpdateTransactionInput is the input for Mutation.updateTransaction.
type UpdateTransactionInput struct {
	Type          *string
	Time          *graphql.Time
	AssetID       *graphql.ID
	CashAccountID *graphql.ID
	Units         *float64
	Price         *Decimal
	Fees          *int32
	Taxes         *int32
	CashDelta     *int32
	Currency      *string
	Note          *string
}

// UpdateTransaction resolves Mutation.updateTransaction.
func (r *RootResolver) UpdateTransaction(ctx context.Context, args struct {
	ID    graphql.ID
	Input UpdateTransactionInput
}) (*TransactionResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	current, err := r.getOwnedTransaction(ctx, person, string(args.ID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("transaction %q not found", args.ID)
	} else if err != nil {
		return nil, err
	}

	// Merge the provided fields into the current values.
	in := args.Input
	if current.TransferID.Valid && (in.Type != nil || in.AssetID != nil || in.Units != nil || in.Price != nil || in.Fees != nil || in.Taxes != nil) {
		return nil, errors.New("a transfer leg only takes time, cash account, cash delta and currency; delete the transfer to change more")
	}
	d := txData{
		Type:          current.Type,
		Time:          current.Time,
		PortfolioID:   stringPtr(current.PortfolioID),
		SecurityID:    stringPtr(current.SecurityID),
		CashAccountID: stringPtr(current.CashAccountID),
		Units:         current.Units,
		Fees:          current.Fees,
		Taxes:         current.Taxes,
		CashDelta:     current.CashDelta,
		Currency:      current.Currency,
		Price:         current.Price,
	}

	if in.Type != nil {
		d.Type = *in.Type
	}
	if in.Time != nil {
		d.Time = in.Time.Time.UTC()
	}
	if in.AssetID != nil {
		d.SecurityID = idPtr(in.AssetID)
	}
	if in.CashAccountID != nil {
		d.CashAccountID = idPtr(in.CashAccountID)
	}
	if in.Units != nil {
		d.Units = *in.Units
	}
	if in.Price != nil {
		d.Price = persistence.Decimal(*in.Price)
	}
	if in.Fees != nil {
		d.Fees = int64(*in.Fees)
	}
	if in.Taxes != nil {
		d.Taxes = int64(*in.Taxes)
	}
	if in.Currency != nil {
		d.Currency = *in.Currency
	}

	// The cash delta follows an explicit value if given; otherwise it is
	// re-derived when any field it depends on changed.
	if in.CashDelta != nil {
		d.CashDelta = int64(*in.CashDelta)
	} else if in.Type != nil || in.Units != nil || in.Price != nil || in.Fees != nil || in.Taxes != nil {
		if delta, ok := deriveCashDelta(&d); ok {
			d.CashDelta = delta
		}
	}

	if err := d.validate(ctx, r.db, person); err != nil {
		return nil, err
	}

	params := persistence.UpdateTransactionParams{
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
		Note:          current.Note,
		ID:            string(args.ID),
	}
	if args.Input.Note != nil {
		params.Note = noteColumn(*args.Input.Note)
	}
	if !current.TransferID.Valid {
		transaction, err := r.db.UpdateTransaction(ctx, params)
		if err != nil {
			return nil, err
		}
		return &TransactionResolver{db: r.db, transaction: transaction}, nil
	}

	// The other leg of a transfer follows: same moment, same currency, the
	// mirrored amount, its own account untouched.
	other, err := r.db.GetTransferCounterpart(ctx, persistence.GetTransferCounterpartParams{TransferID: current.TransferID, ID: current.ID})
	if err != nil {
		return nil, err
	}
	var transaction *persistence.Transaction
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		var err error
		if transaction, err = q.UpdateTransaction(ctx, params); err != nil {
			return err
		}
		_, err = q.UpdateTransaction(ctx, persistence.UpdateTransactionParams{
			Type:          other.Type,
			Time:          d.Time,
			PortfolioID:   other.PortfolioID,
			SecurityID:    other.SecurityID,
			CashAccountID: other.CashAccountID,
			Units:         other.Units,
			Price:         other.Price,
			Fees:          other.Fees,
			Taxes:         other.Taxes,
			CashDelta:     -d.CashDelta,
			Currency:      d.Currency,
			Note:          other.Note,
			ID:            other.ID,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &TransactionResolver{db: r.db, transaction: transaction}, nil
}

// DeleteTransaction resolves Mutation.deleteTransaction.
func (r *RootResolver) DeleteTransaction(ctx context.Context, args struct{ ID graphql.ID }) (graphql.ID, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return "", err
	}

	// The ownership check guards the unscoped delete.
	current, err := r.getOwnedTransaction(ctx, person, string(args.ID))
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("transaction %q not found", args.ID)
	} else if err != nil {
		return "", err
	}

	// A transfer is one event: its two legs go together.
	if current.TransferID.Valid {
		if _, err := r.db.DeleteTransfer(ctx, current.TransferID); err != nil {
			return "", err
		}
		return args.ID, nil
	}
	if _, err := r.db.DeleteTransaction(ctx, string(args.ID)); err != nil {
		return "", err
	}

	return args.ID, nil
}

func (r *TransactionResolver) ID() graphql.ID {
	return graphql.ID(r.transaction.ID)
}

func (r *TransactionResolver) Type() string {
	return r.transaction.Type
}

func (r *TransactionResolver) Time() graphql.Time {
	return graphql.Time{Time: r.transaction.Time}
}

// Portfolio resolves Transaction.portfolio.
func (r *TransactionResolver) Portfolio(ctx context.Context) (*PortfolioResolver, error) {
	if !r.transaction.PortfolioID.Valid {
		return nil, nil
	}

	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	portfolio, err := r.db.GetPortfolio(ctx, persistence.GetPortfolioParams{
		ID:       r.transaction.PortfolioID.String,
		PersonID: person.ID,
	})
	if err != nil {
		return nil, err
	}

	return &PortfolioResolver{db: r.db, portfolio: portfolio}, nil
}

// Security resolves Transaction.asset.
func (r *TransactionResolver) Asset(ctx context.Context) (*AssetResolver, error) {
	if !r.transaction.SecurityID.Valid {
		return nil, nil
	}

	security, err := r.db.GetSecurity(ctx, r.transaction.SecurityID.String)
	if err != nil {
		return nil, err
	}

	return &AssetResolver{db: r.db, security: security}, nil
}

// CashAccount resolves Transaction.cashAccount.
func (r *TransactionResolver) CashAccount(ctx context.Context) (*CashAccountResolver, error) {
	if !r.transaction.CashAccountID.Valid {
		return nil, nil
	}

	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	account, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
		ID:       r.transaction.CashAccountID.String,
		PersonID: person.ID,
	})
	if err != nil {
		return nil, err
	}

	return &CashAccountResolver{db: r.db, account: account}, nil
}

func (r *TransactionResolver) Units() float64 {
	return r.transaction.Units
}

func (r *TransactionResolver) Price() *Price {
	if !r.transaction.Price.Valid {
		return nil
	}

	p := price(r.transaction.Price, priceCurrencyOf(r.transaction.PriceCurrency, r.transaction.Currency))
	return &p
}

// FxRate resolves Transaction.fxRate — see StagedTransactionResolver.FxRate.
func (r *TransactionResolver) FxRate() *Decimal {
	if !r.transaction.FxRate.Valid {
		return nil
	}
	d := decimal(r.transaction.FxRate)
	return &d
}

func (r *TransactionResolver) Fees() Money {
	return money(r.transaction.Fees, r.transaction.Currency)
}

func (r *TransactionResolver) Taxes() Money {
	return money(r.transaction.Taxes, r.transaction.Currency)
}

func (r *TransactionResolver) CashDelta() Money {
	return money(r.transaction.CashDelta, r.transaction.Currency)
}

// Gain resolves Transaction.gain: the realized profit/loss of a SELL,
// costed FIFO against the security's full transaction history (not just
// this portfolio's) — null for every other type.
func (r *TransactionResolver) Gain(ctx context.Context) (*Money, error) {
	if r.transaction.Type != "SELL" {
		return nil, nil
	}

	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	txs, err := allPersonTransactions(ctx, r.db, person.ID)
	if err != nil {
		return nil, err
	}

	amount, ok := finance.RealizedGainOf(txs, r.transaction.ID)
	if !ok {
		return nil, nil
	}

	m := money(amount, r.transaction.Currency)
	return &m, nil
}

func (r *TransactionResolver) Source() string {
	return r.transaction.Source
}

// Document resolves Transaction.document.
func (r *TransactionResolver) Document(ctx context.Context) (*DocumentResolver, error) {
	if !r.transaction.DocumentID.Valid {
		return nil, nil
	}

	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	doc, err := r.db.GetDocument(ctx, persistence.GetDocumentParams{
		ID:       r.transaction.DocumentID.String,
		PersonID: person.ID,
	})
	if err != nil {
		return nil, err
	}

	return &DocumentResolver{db: r.db, doc: doc}, nil
}

func idPtr(id *graphql.ID) *string {
	if id == nil {
		return nil
	}

	s := string(*id)
	return &s
}

func int64Value(i *int32) int64 {
	if i == nil {
		return 0
	}

	return int64(*i)
}

// noteColumn stores a note trimmed; empty means none.
func noteColumn(note string) sql.NullString {
	note = strings.TrimSpace(note)
	if note == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: note, Valid: true}
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Note resolves Transaction.note.
func (r *TransactionResolver) Note() *string {
	if !r.transaction.Note.Valid {
		return nil
	}
	return &r.transaction.Note.String
}
