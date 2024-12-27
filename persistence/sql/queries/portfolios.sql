-- name: GetPortfolio :one
SELECT
    *
FROM
    portfolios
WHERE
    id = ?;

-- name: ListPortfolios :many
SELECT
    *
FROM
    portfolios
ORDER BY
    id;

-- name: CreateBankAccount :one
INSERT INTO
    bank_accounts (id, display_name)
VALUES
    (?, ?) RETURNING *;