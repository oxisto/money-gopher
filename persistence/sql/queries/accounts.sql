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

-- name: CreateAccount :one
INSERT INTO
    accounts (id, display_name, type, reference_account_id)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: DeleteAccount :one
DELETE FROM accounts
WHERE
    id = ? RETURNING *;

-- name: CreatePortfolio :one
INSERT INTO
    portfolios (id, display_name)
VALUES
    (?, ?) RETURNING *;

-- name: UpdatePortfolio :one
UPDATE portfolios
SET
    display_name = ?
WHERE
    id = ? RETURNING *;

-- name: AddAccountToPortfolio :exec
INSERT INTO
    portfolio_accounts (portfolio_id, account_id)
VALUES
    (?, ?);

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

-- name: ListAccountsByPortfolioID :many
SELECT
    accounts.*
FROM
    accounts
    JOIN portfolio_accounts ON accounts.id = portfolio_accounts.account_id
WHERE
    portfolio_accounts.portfolio_id = ?;

-- name: ListTransactionsByAccountID :many
SELECT
    *
FROM
    transactions
WHERE
    source_account_id = sqlc.arg ('account_id')
    OR destination_account_id = sqlc.arg ('account_id');

-- name: ListTransactionsByPortfolioID :many
SELECT
    *
FROM
    transactions
WHERE
    source_account_id IN (
        SELECT
            account_id
        FROM
            portfolio_accounts
        WHERE
            portfolio_accounts.portfolio_id = sqlc.arg ('portfolio_id')
    )
    OR destination_account_id IN (
        SELECT
            account_id
        FROM
            portfolio_accounts
        WHERE
            portfolio_accounts.portfolio_id = sqlc.arg ('portfolio_id')
    );