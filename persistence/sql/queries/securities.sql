-- name: GetSecurity :one
SELECT
    *
FROM
    securities
WHERE
    id = ?;

-- name: ListSecurities :many
SELECT
    *
FROM
    securities
ORDER BY
    id;

-- name: CreateSecurity :one
INSERT INTO
    securities (id, display_name, quote_provider)
VALUES
    (?, ?, ?) RETURNING *;

-- name: UpdateSecurity :one
UPDATE securities
SET
    display_name = ?,
    quote_provider = ?
WHERE
    id = ? RETURNING *;

-- name: DeleteListedSecurity :one
DELETE FROM listed_securities
WHERE
    security_id = ?
    AND ticker = ? RETURNING *;

-- name: UpsertListedSecurity :one
INSERT INTO
    listed_securities (
        security_id,
        ticker,
        currency,
        latest_quote,
        latest_quote_timestamp
    )
VALUES
    (?, ?, ?, ?, ?) ON CONFLICT (security_id, ticker) DO
UPDATE
SET
    ticker = excluded.ticker,
    currency = excluded.currency,
    latest_quote = excluded.latest_quote,
    latest_quote_timestamp = excluded.latest_quote_timestamp RETURNING *;

-- name: ListListedSecuritiesBySecurityID :many
SELECT
    listed_securities.*
FROM
    listed_securities,
    securities
WHERE
    securities.id = listed_securities.security_id
    AND securities.id = ?;