-- name: SaveUser :one
INSERT INTO users(email, password) VALUES ($1, $2)
RETURNING id, email, username, verified, created_at, updated_at;

-- name: VerifyUser :one
UPDATE users SET verified = true, updated_at = now() WHERE email = $1
RETURNING id, verified, updated_at;

-- name: UpdateUsername :one
UPDATE users SET username = $1, updated_at = now() WHERE id = $2
RETURNING id, username, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password, username, verified, created_at, updated_at FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserById :one
SELECT id, email, username, verified, created_at, updated_at FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT id, email, username, verified, created_at, updated_at FROM users WHERE username = $1 LIMIT 1;

-- name: ResetPassword :one
UPDATE users SET password = $1, updated_at = now() WHERE email = $2
RETURNING id, email, password, updated_at;

