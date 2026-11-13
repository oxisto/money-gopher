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
        transfer_id,
        schedule_id,
        scheduled_for,
        note)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

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
-- Every transaction the document describes: imported from it, or confirmed by it.
SELECT
    transactions.*
FROM
    transactions
    JOIN transaction_documents ON transaction_documents.transaction_id = transactions.id
WHERE
    transaction_documents.document_id = ?
ORDER BY
    transactions.time;

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
    fx_rate = ?,
    note = ?
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
    currency = ?,
    note = ?
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

-- name: ListTransactionsByScheduleOccurrence :many
-- The transaction(s) a schedule booked for one occurrence: both legs of a transfer.
SELECT
    *
FROM
    transactions
WHERE
    schedule_id = ?
    AND scheduled_for = ?;

-- name: ListTransactionsBySchedule :many
SELECT
    *
FROM
    transactions
WHERE
    schedule_id = ?
ORDER BY
    scheduled_for;

-- name: ListConfirmableTransactions :many
-- What a statement row may confirm on this account: a schedule's booking
-- or a transfer leg, whether or not another document describes it already.
SELECT
    *
FROM
    transactions
WHERE
    cash_account_id = ?
    AND type = ?
    AND (
        schedule_id IS NOT NULL
        OR transfer_id IS NOT NULL
    );

-- name: ConfirmTransaction :one
-- A statement row confirms a booking: the bank's date and amounts win.
UPDATE transactions
SET
    time = ?,
    fees = ?,
    taxes = ?,
    cash_delta = ?,
    source = ?
WHERE
    id = ? RETURNING *;

-- name: UnlinkScheduleTransactions :exec
-- Before a schedule is deleted: its bookings stay, as plain transactions.
UPDATE transactions
SET
    schedule_id = NULL,
    scheduled_for = NULL
WHERE
    schedule_id = ?;

-- name: ListAdoptableTransactions :many
-- Transactions a new schedule may adopt as its own occurrences: on the
-- account, of the type and amount, in the window, not yet a schedule's.
SELECT
    *
FROM
    transactions
WHERE
    cash_account_id = ?
    AND type = ?
    AND cash_delta = ?
    AND schedule_id IS NULL
    AND time >= ?
    AND time <= ?
ORDER BY
    time;

-- name: LinkTransactionToSchedule :exec
UPDATE transactions
SET
    schedule_id = ?,
    scheduled_for = ?
WHERE
    id = ?;
