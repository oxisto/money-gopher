-- name: GetPortfolio :one
SELECT
    *
FROM
    portfolios
WHERE
    id = ?;

-- name: ListPortfolios :many
SELECT
    *
FROM
    portfolios
ORDER BY
    id;

-- name: ListPortfolioEventsByPortfolioID :many
SELECT
    *
FROM
    portfolio_events
WHERE
    portfolio_id = ?;

-- name: ListAccounts :many
SELECT
    *
FROM
    accounts
ORDER BY
    id;

-- name: GetAccount :one
SELECT
    *
FROM
    accounts
WHERE
    id = ?;

-- name: GetBankAccount :one
SELECT
    *
FROM
    bank_accounts
WHERE
    id = ?;

-- name: CreateAccount :one
INSERT INTO
    accounts (id, display_name, type, reference_account_id)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: CreateBankAccount :one
INSERT INTO
    bank_accounts (id, display_name)
VALUES
    (?, ?) RETURNING *;

-- name: CreatePortfolio :one
INSERT INTO
    portfolios (id, display_name, bank_account_id)
VALUES
    (?, ?, ?) RETURNING *;