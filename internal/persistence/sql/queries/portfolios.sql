-- name: CreatePortfolio :one
INSERT INTO
    portfolios (id, person_id, display_name, cash_account_id, purpose)
VALUES
    (?, ?, ?, ?, ?) RETURNING *;

-- name: GetAssetsPortfolio :one
SELECT
    *
FROM
    portfolios
WHERE
    person_id = ?
    AND purpose = 'ASSETS';

-- name: GetPortfolio :one
SELECT
    *
FROM
    portfolios
WHERE
    id = ?
    AND person_id = ?;

-- name: ListPortfolios :many
SELECT
    *
FROM
    portfolios
WHERE
    person_id = ?
ORDER BY
    display_name;

-- name: UpdatePortfolio :one
UPDATE portfolios
SET
    display_name = ?,
    cash_account_id = ?
WHERE
    id = ?
    AND person_id = ? RETURNING *;

-- name: DeletePortfolio :execrows
DELETE FROM portfolios
WHERE
    id = ?
    AND person_id = ?;

-- name: GetPortfolioHoldingSecurity :one
-- The person's portfolio that last took the security in: where a rent or
-- dividend schedule books its income.
SELECT
    portfolios.*
FROM
    portfolios
    JOIN transactions ON transactions.portfolio_id = portfolios.id
WHERE
    portfolios.person_id = ?
    AND transactions.security_id = ?
    AND transactions.type IN ('BUY', 'DELIVERY_INBOUND')
ORDER BY
    transactions.time DESC
LIMIT
    1;
