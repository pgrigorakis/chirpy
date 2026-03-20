-- name: CreateRefreshTokens :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, user_id)
VALUES (
    $1,
    now(),
    now(),
    $2,		
    $3
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1
AND revoked_at IS NULL
AND expires_at > NOW();

-- name: GetUserFromRefreshToken :one
SELECT user_id FROM refresh_tokens
WHERE token = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET
updated_at=$1,
revoked_at=$2
WHERE token=$3;
