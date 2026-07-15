-- name: CreateCashAccount :one
INSERT INTO
    cash_accounts (id, person_id, display_name, currency, iban)
VALUES
    (?, ?, ?, ?, ?) RETURNING *;

-- name: GetCashAccountByIBAN :one
SELECT
    *
FROM
    cash_accounts
WHERE
    iban = ?
    AND person_id = ?;

-- name: GetCashAccount :one
SELECT
    *
FROM
    cash_accounts
WHERE
    id = ?
    AND person_id = ?;

-- name: ListCashAccounts :many
SELECT
    *
FROM
    cash_accounts
WHERE
    person_id = ?
ORDER BY
    display_name;

-- name: UpdateCashAccount :one
UPDATE cash_accounts
SET
    display_name = ?
WHERE
    id = ?
    AND person_id = ? RETURNING *;

-- name: DeleteCashAccount :execrows
DELETE FROM cash_accounts
WHERE
    id = ?
    AND person_id = ?;

-- name: GetCashAccountBalance :one
SELECT
    CAST(COALESCE(SUM(cash_delta), 0) AS INTEGER) AS balance
FROM
    transactions
WHERE
    cash_account_id = ?;
