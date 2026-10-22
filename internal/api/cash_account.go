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
	"github.com/oxisto/money-gopher/internal/persistence"
)

// CashAccountResolver resolves the CashAccount GraphQL type.
type CashAccountResolver struct {
	db      *persistence.DB
	account *persistence.CashAccount
}

// CashAccounts resolves Query.cashAccounts.
func (r *RootResolver) CashAccounts(ctx context.Context) ([]*CashAccountResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	accounts, err := r.db.ListCashAccounts(ctx, person.ID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*CashAccountResolver, 0, len(accounts))
	for _, a := range accounts {
		resolvers = append(resolvers, &CashAccountResolver{db: r.db, account: a})
	}

	return resolvers, nil
}

// CashAccount resolves Query.cashAccount.
func (r *RootResolver) CashAccount(ctx context.Context, args struct{ ID graphql.ID }) (*CashAccountResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	account, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
		ID:       string(args.ID),
		PersonID: person.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &CashAccountResolver{db: r.db, account: account}, nil
}

// CreateCashAccountInput is the input for Mutation.createCashAccount.
type CreateCashAccountInput struct {
	DisplayName     string
	Kind            *string
	Currency        string
	IBAN            *string
	Reference       *string
	DeclaredBalance *int32
}

// kindColumn is the stored form of an AccountKind: NULL for CASH, so
// every account from before kinds existed reads as one.
func kindColumn(kind *string) sql.NullString {
	if kind == nil || *kind == "CASH" {
		return sql.NullString{}
	}
	return sql.NullString{String: *kind, Valid: true}
}

// kindOf is the AccountKind of a stored account.
func kindOf(account *persistence.CashAccount) string {
	if account.Kind.Valid {
		return account.Kind.String
	}
	return "CASH"
}

// declared is the estimate columns for a balance declared now.
func declared(balance int32) (sql.NullInt64, sql.NullTime) {
	return sql.NullInt64{Int64: int64(balance), Valid: true}, sql.NullTime{Time: time.Now().UTC(), Valid: true}
}

// CreateCashAccount resolves Mutation.createCashAccount.
func (r *RootResolver) CreateCashAccount(ctx context.Context, args struct{ Input CreateCashAccountInput }) (*CashAccountResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := persistence.CreateCashAccountParams{
		ID:          uuid.NewString(),
		PersonID:    person.ID,
		DisplayName: args.Input.DisplayName,
		Currency:    args.Input.Currency,
		Iban:        normalizedIBAN(args.Input.IBAN),
		Kind:        kindColumn(args.Input.Kind),
		Reference:   referenceColumn(args.Input.Reference),
	}
	if args.Input.DeclaredBalance != nil {
		params.EstimatedBalance, params.EstimatedAt = declared(*args.Input.DeclaredBalance)
	}
	account, err := r.db.CreateCashAccount(ctx, params)
	if err != nil {
		return nil, err
	}

	return &CashAccountResolver{db: r.db, account: account}, nil
}

// UpdateCashAccountInput is the input for Mutation.updateCashAccount.
type UpdateCashAccountInput struct {
	DisplayName     *string
	Kind            *string
	IBAN            *string
	Reference       *string
	Estimated       *bool
	DeclaredBalance *int32
}

// UpdateCashAccount resolves Mutation.updateCashAccount.
func (r *RootResolver) UpdateCashAccount(ctx context.Context, args struct {
	ID    graphql.ID
	Input UpdateCashAccountInput
}) (*CashAccountResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	current, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
		ID:       string(args.ID),
		PersonID: person.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("cash account %q not found", args.ID)
	} else if err != nil {
		return nil, err
	}

	displayName := current.DisplayName
	if args.Input.DisplayName != nil {
		displayName = *args.Input.DisplayName
	}
	iban := current.Iban
	if args.Input.IBAN != nil {
		iban = normalizedIBAN(args.Input.IBAN) // "" clears it
	}
	kind := current.Kind
	if args.Input.Kind != nil {
		kind = kindColumn(args.Input.Kind)
	}
	reference := current.Reference
	if args.Input.Reference != nil {
		reference = referenceColumn(args.Input.Reference) // "" clears it
	}
	// A declaration (re)sets the estimate as of now; estimated: false
	// returns the account to its ledger; estimated: true without a balance,
	// now or earlier, has nothing to declare.
	balance, at := current.EstimatedBalance, current.EstimatedAt
	if args.Input.DeclaredBalance != nil {
		balance, at = declared(*args.Input.DeclaredBalance)
	}
	if args.Input.Estimated != nil {
		if !*args.Input.Estimated {
			balance, at = sql.NullInt64{}, sql.NullTime{}
		} else if !balance.Valid {
			return nil, errors.New("an estimated account needs a declared balance")
		}
	}

	account, err := r.db.UpdateCashAccount(ctx, persistence.UpdateCashAccountParams{
		DisplayName:      displayName,
		Iban:             iban,
		Kind:             kind,
		EstimatedBalance: balance,
		EstimatedAt:      at,
		Reference:        reference,
		ID:               string(args.ID),
		PersonID:         person.ID,
	})
	if err != nil {
		return nil, err
	}

	return &CashAccountResolver{db: r.db, account: account}, nil
}

// DeleteCashAccount resolves Mutation.deleteCashAccount.
func (r *RootResolver) DeleteCashAccount(ctx context.Context, args struct{ ID graphql.ID }) (graphql.ID, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return "", err
	}

	rows, err := r.db.DeleteCashAccount(ctx, persistence.DeleteCashAccountParams{
		ID:       string(args.ID),
		PersonID: person.ID,
	})
	if err != nil {
		return "", err
	}
	if rows == 0 {
		return "", fmt.Errorf("cash account %q not found", args.ID)
	}

	return args.ID, nil
}

func (r *CashAccountResolver) ID() graphql.ID      { return graphql.ID(r.account.ID) }
func (r *CashAccountResolver) DisplayName() string { return r.account.DisplayName }
func (r *CashAccountResolver) Currency() string    { return r.account.Currency }
func (r *CashAccountResolver) IBAN() *string {
	if r.account.Iban.Valid {
		return &r.account.Iban.String
	}
	return nil
}

func (r *CashAccountResolver) Reference() *string {
	if r.account.Reference.Valid {
		return &r.account.Reference.String
	}
	return nil
}

func (r *CashAccountResolver) Kind() string    { return kindOf(r.account) }
func (r *CashAccountResolver) Estimated() bool { return r.account.EstimatedBalance.Valid }

func (r *CashAccountResolver) DeclaredBalance() *Money {
	if !r.account.EstimatedBalance.Valid {
		return nil
	}
	m := money(r.account.EstimatedBalance.Int64, r.account.Currency)
	return &m
}

func (r *CashAccountResolver) DeclaredAt() *graphql.Time {
	if !r.account.EstimatedAt.Valid {
		return nil
	}
	return &graphql.Time{Time: r.account.EstimatedAt.Time}
}

// LedgerBalance resolves CashAccount.ledgerBalance: the sum of all cash
// deltas, whatever the account's mode.
func (r *CashAccountResolver) LedgerBalance(ctx context.Context) (Money, error) {
	balance, err := r.db.GetCashAccountBalance(ctx, sql.NullString{String: r.account.ID, Valid: true})
	if err != nil {
		return Money{}, err
	}
	return money(balance, r.account.Currency), nil
}

// Balance resolves CashAccount.balance: what the owner declared for an
// estimated account, the ledger sum otherwise.
func (r *CashAccountResolver) Balance(ctx context.Context) (Money, error) {
	if r.account.EstimatedBalance.Valid {
		return money(r.account.EstimatedBalance.Int64, r.account.Currency), nil
	}
	return r.LedgerBalance(ctx)
}

// Transactions resolves CashAccount.transactions, newest first.
func (r *CashAccountResolver) Transactions(ctx context.Context) ([]*TransactionResolver, error) {
	transactions, err := r.db.ListTransactionsByCashAccount(ctx, sql.NullString{String: r.account.ID, Valid: true})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*TransactionResolver, 0, len(transactions))
	for _, t := range transactions {
		resolvers = append(resolvers, &TransactionResolver{db: r.db, transaction: t})
	}

	return resolvers, nil
}

// referenceColumn stores an account or contract number without spaces, the
// way a parser reads it off a statement ("7 038 222 803" → "7038222803").
func referenceColumn(ref *string) sql.NullString {
	if ref == nil {
		return sql.NullString{}
	}
	s := strings.Join(strings.Fields(*ref), "")
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// normalizedIBAN is nullString for an IBAN: stored without spaces, upper
// case, so a statement's IBAN finds it. Empty stays NULL.
func normalizedIBAN(iban *string) sql.NullString {
	if iban == nil || persistence.NormalizeIBAN(*iban) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: persistence.NormalizeIBAN(*iban), Valid: true}
}
