-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = ?;

-- name: GetUserByIdentity :one
SELECT
    *
FROM
    users
WHERE
    issuer = ?
    AND subject = ?;

-- name: CreateUser :one
INSERT INTO
    users (id, issuer, subject, display_name)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: ListUsers :many
SELECT
    *
FROM
    users
ORDER BY
    display_name;
