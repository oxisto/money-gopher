-- name: CreatePerson :one
INSERT INTO persons (id, display_name)
VALUES (?, ?)
RETURNING *;

-- name: GetPerson :one
SELECT * FROM persons WHERE id = ?;

-- name: ListPersonsForUser :many
SELECT p.*
FROM persons p
JOIN user_person_access upa ON upa.person_id = p.id
WHERE upa.user_id = ?
ORDER BY p.display_name;

-- name: GrantPersonAccess :exec
INSERT INTO user_person_access (user_id, person_id)
VALUES (?, ?)
ON CONFLICT DO NOTHING;

-- name: CheckPersonAccess :one
SELECT COUNT(*) FROM user_person_access
WHERE user_id = ? AND person_id = ?;
