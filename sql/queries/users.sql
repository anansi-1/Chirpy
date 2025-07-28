-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email,hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id,created_at,updated_at,email;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, body, user_id;

-- name: GetAllChirps :many
SELECT id, created_at, updated_at, body, user_id 
FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpsByID :one
SELECT id, created_at, updated_at, body, user_id 
FROM chirps 
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, email,hashed_password
FROM users
WHERE email =$1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  token, created_at, updated_at, user_id, expires_at, revoked_at
)
VALUES (
  $1, NOW(), NOW(), $2, $3, NULL
)
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at;

-- name: GetRefreshToken :one
SELECT token, created_at, updated_at, user_id, expires_at, revoked_at 
FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1 AND revoked_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET email = $2,
    hashed_password = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id,created_at,updated_at,email;

-- name: DeleteChirpsByID :exec
DELETE FROM chirps
WHERE id = $1;