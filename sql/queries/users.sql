-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2
)
RETURNING *;

-- name: GetUserByID :one
SELECT * from users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * from users WHERE email = $1 LIMIT 1;

-- name: GetAllUsers :many
SELECT * from users;

-- name: DeleteUsers :exec
DELETE FROM users;

