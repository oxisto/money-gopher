// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: securities.sql

package persistence

import (
	"context"
	"database/sql"
)

const createSecurity = `-- name: CreateSecurity :one
INSERT INTO
    securities (id, display_name)
VALUES
    (?, ?) RETURNING id, display_name, quote_provider
`

type CreateSecurityParams struct {
	ID          string
	DisplayName string
}

func (q *Queries) CreateSecurity(ctx context.Context, arg CreateSecurityParams) (*Security, error) {
	row := q.db.QueryRowContext(ctx, createSecurity, arg.ID, arg.DisplayName)
	var i Security
	err := row.Scan(&i.ID, &i.DisplayName, &i.QuoteProvider)
	return &i, err
}

const deleteListedSecurity = `-- name: DeleteListedSecurity :one
DELETE FROM listed_securities
WHERE
    security_id = ?
    AND ticker = ? RETURNING security_id, ticker, currency, latest_quote, latest_quote_timestamp
`

type DeleteListedSecurityParams struct {
	SecurityID string
	Ticker     string
}

func (q *Queries) DeleteListedSecurity(ctx context.Context, arg DeleteListedSecurityParams) (*ListedSecurity, error) {
	row := q.db.QueryRowContext(ctx, deleteListedSecurity, arg.SecurityID, arg.Ticker)
	var i ListedSecurity
	err := row.Scan(
		&i.SecurityID,
		&i.Ticker,
		&i.Currency,
		&i.LatestQuote,
		&i.LatestQuoteTimestamp,
	)
	return &i, err
}

const getSecurity = `-- name: GetSecurity :one
SELECT
    id, display_name, quote_provider
FROM
    securities
WHERE
    id = ?
`

func (q *Queries) GetSecurity(ctx context.Context, id string) (*Security, error) {
	row := q.db.QueryRowContext(ctx, getSecurity, id)
	var i Security
	err := row.Scan(&i.ID, &i.DisplayName, &i.QuoteProvider)
	return &i, err
}

const listListedSecuritiesBySecurityID = `-- name: ListListedSecuritiesBySecurityID :many
SELECT
    listed_securities.security_id, listed_securities.ticker, listed_securities.currency, listed_securities.latest_quote, listed_securities.latest_quote_timestamp
FROM
    listed_securities,
    securities
WHERE
    securities.id = listed_securities.security_id
    AND securities.id = ?
`

func (q *Queries) ListListedSecuritiesBySecurityID(ctx context.Context, id string) ([]*ListedSecurity, error) {
	rows, err := q.db.QueryContext(ctx, listListedSecuritiesBySecurityID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ListedSecurity
	for rows.Next() {
		var i ListedSecurity
		if err := rows.Scan(
			&i.SecurityID,
			&i.Ticker,
			&i.Currency,
			&i.LatestQuote,
			&i.LatestQuoteTimestamp,
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

const listSecurities = `-- name: ListSecurities :many
SELECT
    id, display_name, quote_provider
FROM
    securities
ORDER BY
    id
`

func (q *Queries) ListSecurities(ctx context.Context) ([]*Security, error) {
	rows, err := q.db.QueryContext(ctx, listSecurities)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Security
	for rows.Next() {
		var i Security
		if err := rows.Scan(&i.ID, &i.DisplayName, &i.QuoteProvider); err != nil {
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

const updateSecurity = `-- name: UpdateSecurity :one
UPDATE securities
SET
    display_name = ?,
    quote_provider = ?
WHERE
    id = ? RETURNING id, display_name, quote_provider
`

type UpdateSecurityParams struct {
	DisplayName   string
	QuoteProvider sql.NullString
	ID            string
}

func (q *Queries) UpdateSecurity(ctx context.Context, arg UpdateSecurityParams) (*Security, error) {
	row := q.db.QueryRowContext(ctx, updateSecurity, arg.DisplayName, arg.QuoteProvider, arg.ID)
	var i Security
	err := row.Scan(&i.ID, &i.DisplayName, &i.QuoteProvider)
	return &i, err
}

const upsertListedSecurity = `-- name: UpsertListedSecurity :one
INSERT INTO
    listed_securities (security_id, ticker, currency)
VALUES
    (?, ?, ?) ON CONFLICT (security_id, ticker) DO
UPDATE
SET
    ticker = excluded.ticker,
    currency = excluded.currency RETURNING security_id, ticker, currency, latest_quote, latest_quote_timestamp
`

type UpsertListedSecurityParams struct {
	SecurityID string
	Ticker     string
	Currency   string
}

func (q *Queries) UpsertListedSecurity(ctx context.Context, arg UpsertListedSecurityParams) (*ListedSecurity, error) {
	row := q.db.QueryRowContext(ctx, upsertListedSecurity, arg.SecurityID, arg.Ticker, arg.Currency)
	var i ListedSecurity
	err := row.Scan(
		&i.SecurityID,
		&i.Ticker,
		&i.Currency,
		&i.LatestQuote,
		&i.LatestQuoteTimestamp,
	)
	return &i, err
}
