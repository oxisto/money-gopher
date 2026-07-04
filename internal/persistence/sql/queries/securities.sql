-- name: CreateSecurity :one
INSERT INTO
    securities (id, display_name)
VALUES
    (?, ?) RETURNING *;

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
    display_name;

-- name: UpdateSecurity :one
UPDATE securities
SET
    display_name = ?
WHERE
    id = ? RETURNING *;

-- name: DeleteSecurity :execrows
DELETE FROM securities
WHERE
    id = ?;

-- name: CreateSecurityIdentifier :exec
INSERT INTO
    security_identifiers (security_id, kind, value)
VALUES
    (?, ?, ?);

-- name: ListSecurityIdentifiers :many
SELECT
    *
FROM
    security_identifiers
WHERE
    security_id = ?
ORDER BY
    kind,
    value;

-- name: DeleteSecurityIdentifiers :exec
DELETE FROM security_identifiers
WHERE
    security_id = ?;

-- name: FindSecurityByIdentifier :one
SELECT
    s.*
FROM
    securities s
    JOIN security_identifiers si ON si.security_id = s.id
WHERE
    si.kind = ?
    AND si.value = ?;

-- name: CreateListing :one
INSERT INTO
    listings (
        id,
        security_id,
        exchange,
        ticker,
        currency,
        quote_provider
    )
VALUES
    (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: ListListings :many
SELECT
    *
FROM
    listings
WHERE
    security_id = ?
ORDER BY
    ticker;

-- name: DeleteListings :exec
DELETE FROM listings
WHERE
    security_id = ?;

-- name: CreateQuote :exec
INSERT INTO
    quotes (listing_id, time, price)
VALUES
    (?, ?, ?) ON CONFLICT (listing_id, time) DO
UPDATE
SET
    price = excluded.price;

-- name: GetLatestQuote :one
SELECT
    *
FROM
    quotes
WHERE
    listing_id = ?
ORDER BY
    time DESC
LIMIT
    1;

-- name: ListQuotes :many
SELECT
    *
FROM
    quotes
WHERE
    listing_id = ?
    AND time >= ?
    AND time <= ?
ORDER BY
    time;
