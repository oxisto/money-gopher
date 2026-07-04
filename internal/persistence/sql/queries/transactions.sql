-- name: CreateTransaction :one
INSERT INTO
    transactions (
        id,
        type,
        time,
        portfolio_id,
        security_id,
        cash_account_id,
        units,
        price,
        fees,
        taxes,
        cash_delta,
        currency,
        source
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetTransaction :one
SELECT
    *
FROM
    transactions
WHERE
    id = ?;

-- name: ListTransactionsByPortfolio :many
SELECT
    *
FROM
    transactions
WHERE
    portfolio_id = ?
ORDER BY
    time;

-- name: ListTransactionsByCashAccount :many
SELECT
    *
FROM
    transactions
WHERE
    cash_account_id = ?
ORDER BY
    time;

-- name: UpdateTransaction :one
UPDATE transactions
SET
    type = ?,
    time = ?,
    portfolio_id = ?,
    security_id = ?,
    cash_account_id = ?,
    units = ?,
    price = ?,
    fees = ?,
    taxes = ?,
    cash_delta = ?,
    currency = ?
WHERE
    id = ? RETURNING *;

-- name: DeleteTransaction :execrows
DELETE FROM transactions
WHERE
    id = ?;
