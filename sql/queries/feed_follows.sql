-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, user_id, feed_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFeedFollow :one
SELECT * FROM feed_follows
WHERE id = $1;

-- name: ListUserFeedFollows :many
SELECT * FROM feed_follows
WHERE user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE id = $1;
