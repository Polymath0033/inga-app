-- name: SaveLink :one
INSERT INTO links (user_id, title, url, description, category,remind_at, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, title, url, description, category, remind_at, tags, created_at, updated_at;

-- name: GetLinksByUserID :many
SELECT id, user_id, title, url, description, category, remind_at, tags, created_at, updated_at
FROM links
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetLinkByID :one
SELECT id, user_id, title, url, description, category, remind_at, tags, created_at, updated_at
FROM links
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: UpdateLink :one
UPDATE links
SET title = COALESCE(NULLIF($2, ''), title),
    url = COALESCE(NULLIF($3, ''), url),
    description = COALESCE(NULLIF($4, ''), description),
    category = COALESCE(NULLIF($5, ''), category),
    remind_at = COALESCE($6, remind_at),
    tags = COALESCE($7, tags),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, user_id, title, url, description, category, remind_at, tags, created_at, updated_at;

-- name: DeleteLink :exec
DELETE FROM links WHERE id = $1 AND user_id = $2;