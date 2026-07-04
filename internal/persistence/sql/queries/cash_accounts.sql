-- name: CreateCashAccount :one
INSERT INTO
    cash_accounts (id, user_id, display_name, currency)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: GetCashAccount :one
SELECT
    *
FROM
    cash_accounts
WHERE
    id = ?
    AND user_id = ?;

-- name: ListCashAccounts :many
SELECT
    *
FROM
    cash_accounts
WHERE
    user_id = ?
ORDER BY
    display_name;

-- name: UpdateCashAccount :one
UPDATE cash_accounts
SET
    display_name = ?
WHERE
    id = ?
    AND user_id = ? RETURNING *;

-- name: DeleteCashAccount :execrows
DELETE FROM cash_accounts
WHERE
    id = ?
    AND user_id = ?;

-- name: GetCashAccountBalance :one
SELECT
    CAST(COALESCE(SUM(cash_delta), 0) AS INTEGER) AS balance
FROM
    transactions
WHERE
    cash_account_id = ?;
