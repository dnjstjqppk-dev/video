-- name: ListCommentsByVideo :many
SELECT c.id, c.video_id, c.user_id, c.parent_id, c.content, c.like_count, c.reply_count,
       c.is_pinned, c.is_hearted, c.created_at,
       p.display_name, p.avatar_url, ch.handle as channel_handle
FROM comments c
JOIN users u ON u.id = c.user_id
LEFT JOIN user_profiles p ON p.user_id = u.id
LEFT JOIN channels ch ON ch.owner_id = u.id
WHERE c.video_id = $1 AND c.parent_id IS NULL AND c.is_deleted = FALSE
ORDER BY c.is_pinned DESC, c.like_count DESC, c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateComment :one
INSERT INTO comments (video_id, user_id, parent_id, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteComment :exec
UPDATE comments
SET is_deleted = TRUE, content = '[Deleted Comment]'
WHERE id = $1;

