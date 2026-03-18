-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),
    now(),
    now(),
    $1
)
RETURNING *;

-- name: GetUser :one
SELECT * from users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * from users WHERE id = $1 LIMIT 1;

-- name: GetAllUsers :many
SELECT * from users;

-- name: DeleteUsers :exec
DELETE FROM users;

