-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetSession :one
SELECT *
FROM sessions
WHERE id = ? AND expires_at > now();

-- name: UpdateSessionPerson :exec
UPDATE sessions SET person_id = ? WHERE id = ?;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= now();
