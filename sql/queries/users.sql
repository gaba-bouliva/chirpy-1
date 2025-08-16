-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    $1, NOW(), NOW(), $2
)
RETURNING *;

-- name: GetUser :one
SELECT id, created_at, updated_at, email 
FROM users
WHERE id = $1;