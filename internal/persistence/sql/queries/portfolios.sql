-- name: CreatePortfolio :one
INSERT INTO
    portfolios (id, user_id, display_name, cash_account_id)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: GetPortfolio :one
SELECT
    *
FROM
    portfolios
WHERE
    id = ?
    AND user_id = ?;

-- name: ListPortfolios :many
SELECT
    *
FROM
    portfolios
WHERE
    user_id = ?
ORDER BY
    display_name;

-- name: UpdatePortfolio :one
UPDATE portfolios
SET
    display_name = ?
WHERE
    id = ?
    AND user_id = ? RETURNING *;

-- name: DeletePortfolio :execrows
DELETE FROM portfolios
WHERE
    id = ?
    AND user_id = ?;
