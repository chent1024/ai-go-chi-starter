-- name: CreateExample :one
INSERT INTO examples (id, name)
VALUES ($1, $2)
RETURNING id, name, created_at, updated_at;

-- name: GetExample :one
SELECT id, name, created_at, updated_at
FROM examples
WHERE id = $1;

-- name: ListExamples :many
SELECT id, name, created_at, updated_at
FROM examples
ORDER BY created_at DESC, id DESC
LIMIT 100;
