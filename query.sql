-- name: GetSecurity :one
SELECT * FROM securities
WHERE id = ?;

-- name: ListSecurities :many
SELECT * FROM securities
ORDER BY id;
