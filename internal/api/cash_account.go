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

// CashAccountResolver resolves the CashAccount GraphQL type.
type CashAccountResolver struct {
	db      *persistence.DB
	account *persistence.CashAccount
}

// CashAccounts resolves Query.cashAccounts.
func (r *RootResolver) CashAccounts(ctx context.Context) ([]*CashAccountResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	accounts, err := r.db.ListCashAccounts(ctx, user.ID)
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
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	account, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
		ID:     string(args.ID),
		UserID: user.ID,
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
	DisplayName string
	Currency    string
	IBAN        *string
}

// CreateCashAccount resolves Mutation.createCashAccount.
func (r *RootResolver) CreateCashAccount(ctx context.Context, args struct{ Input CreateCashAccountInput }) (*CashAccountResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	account, err := r.db.CreateCashAccount(ctx, persistence.CreateCashAccountParams{
		ID:          uuid.NewString(),
		UserID:      user.ID,
		DisplayName: args.Input.DisplayName,
		Currency:    args.Input.Currency,
		Iban:        nullString(args.Input.IBAN),
	})
	if err != nil {
		return nil, err
	}

	return &CashAccountResolver{db: r.db, account: account}, nil
}

// UpdateCashAccountInput is the input for Mutation.updateCashAccount.
type UpdateCashAccountInput struct {
	DisplayName *string
}

// UpdateCashAccount resolves Mutation.updateCashAccount.
func (r *RootResolver) UpdateCashAccount(ctx context.Context, args struct {
	ID    graphql.ID
	Input UpdateCashAccountInput
}) (*CashAccountResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	current, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
		ID:     string(args.ID),
		UserID: user.ID,
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

	account, err := r.db.UpdateCashAccount(ctx, persistence.UpdateCashAccountParams{
		DisplayName: displayName,
		ID:          string(args.ID),
		UserID:      user.ID,
	})
	if err != nil {
		return nil, err
	}

	return &CashAccountResolver{db: r.db, account: account}, nil
}

// DeleteCashAccount resolves Mutation.deleteCashAccount.
func (r *RootResolver) DeleteCashAccount(ctx context.Context, args struct{ ID graphql.ID }) (graphql.ID, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return "", err
	}

	rows, err := r.db.DeleteCashAccount(ctx, persistence.DeleteCashAccountParams{
		ID:     string(args.ID),
		UserID: user.ID,
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

// Balance resolves CashAccount.balance as the sum of all cash deltas.
func (r *CashAccountResolver) Balance(ctx context.Context) (Money, error) {
	balance, err := r.db.GetCashAccountBalance(ctx, sql.NullString{String: r.account.ID, Valid: true})
	if err != nil {
		return Money{}, err
	}

	return money(balance, r.account.Currency), nil
}

// Transactions resolves CashAccount.transactions, ordered by time.
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
