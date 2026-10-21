-- name: CreateFxRate :exec
INSERT INTO
    fx_rates (currency, time, rate)
VALUES
    (?, ?, ?) ON CONFLICT (currency, time) DO
UPDATE
SET
    rate = excluded.rate;

-- name: GetFxRateBefore :one
SELECT
    *
FROM
    fx_rates
WHERE
    currency = ?
    AND time <= ?
ORDER BY
    time DESC
LIMIT
    1;

-- name: CountFxRates :one
SELECT
    COUNT(*)
FROM
    fx_rates
WHERE
    currency = ?;

-- name: ListListingCurrencies :many
SELECT DISTINCT
    currency
FROM
    listings
ORDER BY
    currency;

-- name: ListTransactionPriceCurrencies :many
SELECT DISTINCT
    price_currency
FROM
    transactions
WHERE
    price_currency IS NOT NULL
ORDER BY
    price_currency;
