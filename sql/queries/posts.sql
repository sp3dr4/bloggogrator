-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, url, title, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserPosts :many
SELECT p.*
FROM posts p
  INNER JOIN feeds f ON p.feed_id = f.id
WHERE f.user_id = $1
ORDER BY p.published_at DESC
LIMIT $2;
