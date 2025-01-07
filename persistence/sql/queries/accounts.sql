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

-- name: DeleteAccount :one
DELETE FROM accounts
WHERE
    id = ? RETURNING *;

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

-- name: CreateTransaction :one
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
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: ListTransactionsByAccountID :many
SELECT
    *
FROM
    transactions
WHERE
    source_account_id = sqlc.arg ('account_id')
    OR destination_account_id = sqlc.arg ('account_id');