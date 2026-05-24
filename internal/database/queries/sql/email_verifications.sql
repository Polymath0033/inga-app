-- name: SaveVerification :one
INSERT INTO email_verifications (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token, expires_at, used, created_at;

-- name: GetVerificationByToken :one
SELECT id, user_id, token, expires_at, used, created_at
FROM email_verifications
WHERE token = $1
LIMIT 1;

-- name: MarkVerificationUsed :exec
UPDATE email_verifications SET used = true WHERE token = $1;

