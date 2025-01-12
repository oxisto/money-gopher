// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: accounts.sql

package persistence

import (
	"context"
	"time"

	currency "github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/portfolio/accounts"
	"github.com/oxisto/money-gopher/portfolio/events"
)

const addAccountToPortfolio = `-- name: AddAccountToPortfolio :exec
INSERT INTO
    portfolio_accounts (portfolio_id, account_id)
VALUES
    (?, ?)
`

type AddAccountToPortfolioParams struct {
	PortfolioID string
	AccountID   string
}

func (q *Queries) AddAccountToPortfolio(ctx context.Context, arg AddAccountToPortfolioParams) error {
	_, err := q.db.ExecContext(ctx, addAccountToPortfolio, arg.PortfolioID, arg.AccountID)
	return err
}

const createAccount = `-- name: CreateAccount :one
INSERT INTO
    accounts (id, display_name, type, reference_account_id)
VALUES
    (?, ?, ?, ?) RETURNING id, display_name, type, reference_account_id
`

type CreateAccountParams struct {
	ID                 string
	DisplayName        string
	Type               accounts.AccountType
	ReferenceAccountID *int64
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (*Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount,
		arg.ID,
		arg.DisplayName,
		arg.Type,
		arg.ReferenceAccountID,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.DisplayName,
		&i.Type,
		&i.ReferenceAccountID,
	)
	return &i, err
}

const createPortfolio = `-- name: CreatePortfolio :one
INSERT INTO
    portfolios (id, display_name)
VALUES
    (?, ?) RETURNING id, display_name
`

type CreatePortfolioParams struct {
	ID          string
	DisplayName string
}

func (q *Queries) CreatePortfolio(ctx context.Context, arg CreatePortfolioParams) (*Portfolio, error) {
	row := q.db.QueryRowContext(ctx, createPortfolio, arg.ID, arg.DisplayName)
	var i Portfolio
	err := row.Scan(&i.ID, &i.DisplayName)
	return &i, err
}

const createTransaction = `-- name: CreateTransaction :one
INSERT INTO
    transactions (
        id,
        source_account_id,
        destination_account_id,
        time,
        type,
        security_id,
        amount,
        price,
        fees,
        taxes
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id, source_account_id, destination_account_id, time, type, security_id, amount, price, fees, taxes
`

type CreateTransactionParams struct {
	ID                   string
	SourceAccountID      *string
	DestinationAccountID *string
	Time                 time.Time
	Type                 events.PortfolioEventType
	SecurityID           *string
	Amount               float64
	Price                *currency.Currency
	Fees                 *currency.Currency
	Taxes                *currency.Currency
}

func (q *Queries) CreateTransaction(ctx context.Context, arg CreateTransactionParams) (*Transaction, error) {
	row := q.db.QueryRowContext(ctx, createTransaction,
		arg.ID,
		arg.SourceAccountID,
		arg.DestinationAccountID,
		arg.Time,
		arg.Type,
		arg.SecurityID,
		arg.Amount,
		arg.Price,
		arg.Fees,
		arg.Taxes,
	)
	var i Transaction
	err := row.Scan(
		&i.ID,
		&i.SourceAccountID,
		&i.DestinationAccountID,
		&i.Time,
		&i.Type,
		&i.SecurityID,
		&i.Amount,
		&i.Price,
		&i.Fees,
		&i.Taxes,
	)
	return &i, err
}

const deleteAccount = `-- name: DeleteAccount :one
DELETE FROM accounts
WHERE
    id = ? RETURNING id, display_name, type, reference_account_id
`

func (q *Queries) DeleteAccount(ctx context.Context, id string) (*Account, error) {
	row := q.db.QueryRowContext(ctx, deleteAccount, id)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.DisplayName,
		&i.Type,
		&i.ReferenceAccountID,
	)
	return &i, err
}

const getAccount = `-- name: GetAccount :one
SELECT
    id, display_name, type, reference_account_id
FROM
    accounts
WHERE
    id = ?
`

func (q *Queries) GetAccount(ctx context.Context, id string) (*Account, error) {
	row := q.db.QueryRowContext(ctx, getAccount, id)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.DisplayName,
		&i.Type,
		&i.ReferenceAccountID,
	)
	return &i, err
}

const getPortfolio = `-- name: GetPortfolio :one
SELECT
    id, display_name
FROM
    portfolios
WHERE
    id = ?
`

func (q *Queries) GetPortfolio(ctx context.Context, id string) (*Portfolio, error) {
	row := q.db.QueryRowContext(ctx, getPortfolio, id)
	var i Portfolio
	err := row.Scan(&i.ID, &i.DisplayName)
	return &i, err
}

const listAccounts = `-- name: ListAccounts :many
SELECT
    id, display_name, type, reference_account_id
FROM
    accounts
ORDER BY
    id
`

func (q *Queries) ListAccounts(ctx context.Context) ([]*Account, error) {
	rows, err := q.db.QueryContext(ctx, listAccounts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.DisplayName,
			&i.Type,
			&i.ReferenceAccountID,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAccountsByPortfolioID = `-- name: ListAccountsByPortfolioID :many
SELECT
    accounts.id, accounts.display_name, accounts.type, accounts.reference_account_id
FROM
    accounts
    JOIN portfolio_accounts ON accounts.id = portfolio_accounts.account_id
WHERE
    portfolio_accounts.portfolio_id = ?
`

func (q *Queries) ListAccountsByPortfolioID(ctx context.Context, portfolioID string) ([]*Account, error) {
	rows, err := q.db.QueryContext(ctx, listAccountsByPortfolioID, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.DisplayName,
			&i.Type,
			&i.ReferenceAccountID,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPortfolios = `-- name: ListPortfolios :many
SELECT
    id, display_name
FROM
    portfolios
ORDER BY
    id
`

func (q *Queries) ListPortfolios(ctx context.Context) ([]*Portfolio, error) {
	rows, err := q.db.QueryContext(ctx, listPortfolios)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Portfolio
	for rows.Next() {
		var i Portfolio
		if err := rows.Scan(&i.ID, &i.DisplayName); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTransactionsByAccountID = `-- name: ListTransactionsByAccountID :many
SELECT
    id, source_account_id, destination_account_id, time, type, security_id, amount, price, fees, taxes
FROM
    transactions
WHERE
    source_account_id = ?1
    OR destination_account_id = ?1
`

func (q *Queries) ListTransactionsByAccountID(ctx context.Context, accountID *string) ([]*Transaction, error) {
	rows, err := q.db.QueryContext(ctx, listTransactionsByAccountID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Transaction
	for rows.Next() {
		var i Transaction
		if err := rows.Scan(
			&i.ID,
			&i.SourceAccountID,
			&i.DestinationAccountID,
			&i.Time,
			&i.Type,
			&i.SecurityID,
			&i.Amount,
			&i.Price,
			&i.Fees,
			&i.Taxes,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTransactionsByPortfolioID = `-- name: ListTransactionsByPortfolioID :many
SELECT
    id, source_account_id, destination_account_id, time, type, security_id, amount, price, fees, taxes
FROM
    transactions
WHERE
    source_account_id IN (
        SELECT
            account_id
        FROM
            portfolio_accounts
        WHERE
            portfolio_accounts.portfolio_id = ?1
    )
    OR destination_account_id IN (
        SELECT
            account_id
        FROM
            portfolio_accounts
        WHERE
            portfolio_accounts.portfolio_id = ?1
    )
`

func (q *Queries) ListTransactionsByPortfolioID(ctx context.Context, portfolioID string) ([]*Transaction, error) {
	rows, err := q.db.QueryContext(ctx, listTransactionsByPortfolioID, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Transaction
	for rows.Next() {
		var i Transaction
		if err := rows.Scan(
			&i.ID,
			&i.SourceAccountID,
			&i.DestinationAccountID,
			&i.Time,
			&i.Type,
			&i.SecurityID,
			&i.Amount,
			&i.Price,
			&i.Fees,
			&i.Taxes,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updatePortfolio = `-- name: UpdatePortfolio :one
UPDATE portfolios
SET
    display_name = ?
WHERE
    id = ? RETURNING id, display_name
`

type UpdatePortfolioParams struct {
	DisplayName string
	ID          string
}

func (q *Queries) UpdatePortfolio(ctx context.Context, arg UpdatePortfolioParams) (*Portfolio, error) {
	row := q.db.QueryRowContext(ctx, updatePortfolio, arg.DisplayName, arg.ID)
	var i Portfolio
	err := row.Scan(&i.ID, &i.DisplayName)
	return &i, err
}
