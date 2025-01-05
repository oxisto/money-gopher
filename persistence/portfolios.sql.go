// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: portfolios.sql

package persistence

import (
	"context"
)

const createBankAccount = `-- name: CreateBankAccount :one
INSERT INTO
    bank_accounts (id, display_name)
VALUES
    (?, ?) RETURNING id, display_name
`

type CreateBankAccountParams struct {
	ID          string
	DisplayName string
}

func (q *Queries) CreateBankAccount(ctx context.Context, arg CreateBankAccountParams) (*BankAccount, error) {
	row := q.db.QueryRowContext(ctx, createBankAccount, arg.ID, arg.DisplayName)
	var i BankAccount
	err := row.Scan(&i.ID, &i.DisplayName)
	return &i, err
}

const createPortfolio = `-- name: CreatePortfolio :one
INSERT INTO
    portfolios (id, display_name, bank_account_id)
VALUES
    (?, ?, ?) RETURNING id, display_name, bank_account_id
`

type CreatePortfolioParams struct {
	ID            string
	DisplayName   string
	BankAccountID string
}

func (q *Queries) CreatePortfolio(ctx context.Context, arg CreatePortfolioParams) (*Portfolio, error) {
	row := q.db.QueryRowContext(ctx, createPortfolio, arg.ID, arg.DisplayName, arg.BankAccountID)
	var i Portfolio
	err := row.Scan(&i.ID, &i.DisplayName, &i.BankAccountID)
	return &i, err
}

const getBankAccount = `-- name: GetBankAccount :one
SELECT
    id, display_name
FROM
    bank_accounts
WHERE
    id = ?
`

func (q *Queries) GetBankAccount(ctx context.Context, id string) (*BankAccount, error) {
	row := q.db.QueryRowContext(ctx, getBankAccount, id)
	var i BankAccount
	err := row.Scan(&i.ID, &i.DisplayName)
	return &i, err
}

const getPortfolio = `-- name: GetPortfolio :one
SELECT
    id, display_name, bank_account_id
FROM
    portfolios
WHERE
    id = ?
`

func (q *Queries) GetPortfolio(ctx context.Context, id string) (*Portfolio, error) {
	row := q.db.QueryRowContext(ctx, getPortfolio, id)
	var i Portfolio
	err := row.Scan(&i.ID, &i.DisplayName, &i.BankAccountID)
	return &i, err
}

const listPortfolioEventsByPortfolioID = `-- name: ListPortfolioEventsByPortfolioID :many
SELECT
    id, type, time, portfolio_id, security_id, amount, price, fees, taxes
FROM
    portfolio_events
WHERE
    portfolio_id = ?
`

func (q *Queries) ListPortfolioEventsByPortfolioID(ctx context.Context, portfolioID string) ([]*PortfolioEvent, error) {
	rows, err := q.db.QueryContext(ctx, listPortfolioEventsByPortfolioID, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PortfolioEvent
	for rows.Next() {
		var i PortfolioEvent
		if err := rows.Scan(
			&i.ID,
			&i.Type,
			&i.Time,
			&i.PortfolioID,
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

const listPortfolios = `-- name: ListPortfolios :many
SELECT
    id, display_name, bank_account_id
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
		if err := rows.Scan(&i.ID, &i.DisplayName, &i.BankAccountID); err != nil {
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
