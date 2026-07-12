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

// PortfolioResolver resolves the Portfolio GraphQL type.
type PortfolioResolver struct {
	db        *persistence.DB
	portfolio *persistence.Portfolio
}

// Portfolios resolves Query.portfolios.
func (r *RootResolver) Portfolios(ctx context.Context) ([]*PortfolioResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	portfolios, err := r.db.ListPortfolios(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*PortfolioResolver, 0, len(portfolios))
	for _, p := range portfolios {
		resolvers = append(resolvers, &PortfolioResolver{db: r.db, portfolio: p})
	}

	return resolvers, nil
}

// Portfolio resolves Query.portfolio.
func (r *RootResolver) Portfolio(ctx context.Context, args struct{ ID graphql.ID }) (*PortfolioResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	portfolio, err := r.db.GetPortfolio(ctx, persistence.GetPortfolioParams{
		ID:     string(args.ID),
		UserID: user.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &PortfolioResolver{db: r.db, portfolio: portfolio}, nil
}

// CreatePortfolioInput is the input for Mutation.createPortfolio.
type CreatePortfolioInput struct {
	DisplayName   string
	CashAccountID graphql.ID
}

// CreatePortfolio resolves Mutation.createPortfolio.
func (r *RootResolver) CreatePortfolio(ctx context.Context, args struct{ Input CreatePortfolioInput }) (*PortfolioResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Verify the cash account belongs to this user.
	if _, err = r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
		ID:     string(args.Input.CashAccountID),
		UserID: user.ID,
	}); errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("cash account %q not found", args.Input.CashAccountID)
	} else if err != nil {
		return nil, err
	}

	portfolio, err := r.db.CreatePortfolio(ctx, persistence.CreatePortfolioParams{
		ID:            uuid.NewString(),
		UserID:        user.ID,
		DisplayName:   args.Input.DisplayName,
		CashAccountID: string(args.Input.CashAccountID),
	})
	if err != nil {
		return nil, err
	}

	return &PortfolioResolver{db: r.db, portfolio: portfolio}, nil
}

// UpdatePortfolioInput is the input for Mutation.updatePortfolio.
type UpdatePortfolioInput struct {
	DisplayName *string
}

// UpdatePortfolio resolves Mutation.updatePortfolio.
func (r *RootResolver) UpdatePortfolio(ctx context.Context, args struct {
	ID    graphql.ID
	Input UpdatePortfolioInput
}) (*PortfolioResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	current, err := r.db.GetPortfolio(ctx, persistence.GetPortfolioParams{
		ID:     string(args.ID),
		UserID: user.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("portfolio %q not found", args.ID)
	} else if err != nil {
		return nil, err
	}

	displayName := current.DisplayName
	if args.Input.DisplayName != nil {
		displayName = *args.Input.DisplayName
	}

	portfolio, err := r.db.UpdatePortfolio(ctx, persistence.UpdatePortfolioParams{
		DisplayName: displayName,
		ID:          string(args.ID),
		UserID:      user.ID,
	})
	if err != nil {
		return nil, err
	}

	return &PortfolioResolver{db: r.db, portfolio: portfolio}, nil
}

// DeletePortfolio resolves Mutation.deletePortfolio.
func (r *RootResolver) DeletePortfolio(ctx context.Context, args struct{ ID graphql.ID }) (graphql.ID, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return "", err
	}

	rows, err := r.db.DeletePortfolio(ctx, persistence.DeletePortfolioParams{
		ID:     string(args.ID),
		UserID: user.ID,
	})
	if err != nil {
		return "", err
	}
	if rows == 0 {
		return "", fmt.Errorf("portfolio %q not found", args.ID)
	}

	return args.ID, nil
}

func (r *PortfolioResolver) ID() graphql.ID {
	return graphql.ID(r.portfolio.ID)
}

func (r *PortfolioResolver) DisplayName() string {
	return r.portfolio.DisplayName
}

// CashAccount resolves Portfolio.cashAccount.
func (r *PortfolioResolver) CashAccount(ctx context.Context) (*CashAccountResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	account, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
		ID:     r.portfolio.CashAccountID,
		UserID: user.ID,
	})
	if err != nil {
		return nil, err
	}

	return &CashAccountResolver{db: r.db, account: account}, nil
}

// Transactions resolves Portfolio.transactions, ordered by time.
func (r *PortfolioResolver) Transactions(ctx context.Context) ([]*TransactionResolver, error) {
	transactions, err := r.db.ListTransactionsByPortfolio(ctx, sql.NullString{String: r.portfolio.ID, Valid: true})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*TransactionResolver, 0, len(transactions))
	for _, t := range transactions {
		resolvers = append(resolvers, &TransactionResolver{db: r.db, transaction: t})
	}

	return resolvers, nil
}
