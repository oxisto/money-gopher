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
        source,
        document_id,
        price_currency,
        fx_rate,
        transfer_id
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

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
    time DESC;

-- name: ListTransactionsByCashAccount :many
SELECT
    *
FROM
    transactions
WHERE
    cash_account_id = ?
ORDER BY
    time DESC;

-- name: ListTransactionsByDocument :many
SELECT
    *
FROM
    transactions
WHERE
    document_id = ?
ORDER BY
    time;

-- name: UpdateTransactionFromImport :one
-- ConfirmImport's reconciliation update; also refreshes FX fields, unlike UpdateTransaction.
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
    currency = ?,
    price_currency = ?,
    fx_rate = ?
WHERE
    id = ? RETURNING *;

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

-- name: GetTransferCounterpart :one
SELECT
    *
FROM
    transactions
WHERE
    transfer_id = ?
    AND id <> ?;

-- name: DeleteTransfer :execrows
DELETE FROM transactions
WHERE
    transfer_id = ?;
