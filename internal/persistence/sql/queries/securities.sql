-- name: CreateSecurity :one
INSERT INTO
    securities (id, display_name, underlying_security_id, asset_kind, person_id)
VALUES
    (?, ?, ?, ?, ?) RETURNING *;

-- name: GetSecurityVisibleTo :one
-- A shared security, or a personal asset of this person.
SELECT
    *
FROM
    securities
WHERE
    id = ?
    AND (person_id IS NULL OR person_id = ?);

-- name: GetSecurity :one
SELECT
    *
FROM
    securities
WHERE
    id = ?;

-- name: ListSecurities :many
-- The shared, listed instruments.
SELECT
    *
FROM
    securities
WHERE
    person_id IS NULL
ORDER BY
    display_name;

-- name: ListAssetsForPerson :many
-- The person's own assets.
SELECT
    *
FROM
    securities
WHERE
    person_id = ?
ORDER BY
    display_name;

-- name: UpdateSecurity :one
UPDATE securities
SET
    display_name = ?,
    underlying_security_id = ?,
    asset_kind = ?
WHERE
    id = ?
    AND (person_id IS NULL OR person_id = ?) RETURNING *;

-- name: DeleteSecurity :execrows
DELETE FROM securities
WHERE
    id = ?
    AND (person_id IS NULL OR person_id = ?);

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

-- name: FindSecurityByIdentifierVisibleTo :one
-- A shared security or one of this person's own assets carrying the
-- identifier.
SELECT
    s.*
FROM
    securities s
    JOIN security_identifiers si ON si.security_id = s.id
WHERE
    si.kind = ?
    AND si.value = ?
    AND (s.person_id IS NULL OR s.person_id = ?);

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

-- name: GetListingVisibleTo :one
-- A listing of a shared security or of this person's own asset.
SELECT
    l.*
FROM
    listings l
    JOIN securities s ON s.id = l.security_id
WHERE
    l.id = ?
    AND (s.person_id IS NULL OR s.person_id = ?);

-- name: ListListings :many
SELECT
    *
FROM
    listings
WHERE
    security_id = ?
ORDER BY
    ticker;

-- name: ListListingsWithProvider :many
-- Listings worth fetching quotes for: those with a provider, except of
-- securities that were held once and are sold out or redeemed. Quote feeds
-- drop those (a matured bond is gone from ING's site) and a fresh price
-- would not change any figure. A security with no transactions at all is
-- fetched: it was just created, and a purchase is probably on its way.
SELECT
    *
FROM
    listings
WHERE
    quote_provider IS NOT NULL
    AND security_id NOT IN (
        SELECT
            security_id
        FROM
            transactions
        WHERE
            security_id IS NOT NULL
        GROUP BY
            security_id
        HAVING
            SUM(
                CASE
                    WHEN type IN ('BUY', 'DELIVERY_INBOUND') THEN units
                    WHEN type IN ('SELL', 'DELIVERY_OUTBOUND') THEN - units
                    ELSE 0
                END
            ) <= 0.000001
    )
ORDER BY
    security_id,
    ticker;

-- name: DeleteListings :exec
DELETE FROM listings
WHERE
    security_id = ?;

-- name: UpdateListing :one
UPDATE listings
SET
    exchange = ?,
    ticker = ?,
    currency = ?,
    quote_provider = ?
WHERE
    id = ?
    AND security_id = ? RETURNING *;

-- name: DeleteListing :execrows
DELETE FROM listings
WHERE
    id = ?
    AND security_id = ?;

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

-- name: GetLatestQuoteBefore :one
SELECT
    *
FROM
    quotes
WHERE
    listing_id = ?
    AND time <= ?
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

-- name: DeleteQuote :execrows
DELETE FROM quotes
WHERE
    listing_id = ?
    AND time = ?;
