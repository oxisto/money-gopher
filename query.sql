-- name: GetSecurity :one
SELECT * FROM securities
WHERE id = ?;

-- name: ListSecurities :many
SELECT * FROM securities
ORDER BY id;

-- name: CreateSecurity :one
INSERT INTO securities (id, display_name)
VALUES (?, ?)
RETURNING *;
