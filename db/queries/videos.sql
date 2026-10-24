-- name: GetVideoByID :one
SELECT v.*, c.name as channel_name, c.handle as channel_handle, c.avatar_url as channel_avatar_url, c.subscriber_count as channel_subscriber_count
FROM videos v
JOIN channels c ON c.id = v.channel_id
WHERE v.id = $1 AND v.status != 'DELETED';

-- name: ListHomeFeedVideos :many
SELECT v.id, v.channel_id, v.title, v.description, v.video_type, v.visibility, v.status,
       v.duration_seconds, v.thumbnail_url, v.hls_master_url, v.view_count, v.like_count,
       v.category, v.published_at,
       c.name as channel_name, c.handle as channel_handle, c.avatar_url as channel_avatar_url
FROM videos v
JOIN channels c ON c.id = v.channel_id
WHERE v.visibility = 'public' AND v.status = 'READY'
ORDER BY v.published_at DESC
LIMIT $1 OFFSET $2;

-- name: ListShortsFeed :many
SELECT v.id, v.channel_id, v.title, v.description, v.duration_seconds, v.thumbnail_url,
       v.hls_master_url, v.view_count, v.like_count, v.comment_count,
       c.name as channel_name, c.handle as channel_handle, c.avatar_url as channel_avatar_url
FROM videos v
JOIN channels c ON c.id = v.channel_id
WHERE v.video_type = 'shorts' AND v.visibility = 'public' AND v.status = 'READY'
ORDER BY v.published_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateVideo :one
INSERT INTO videos (channel_id, title, description, video_type, visibility, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateVideoReady :one
UPDATE videos
SET status = 'READY',
    duration_seconds = $2,
    hls_master_url = $3,
    thumbnail_url = $4,
    published_at = COALESCE(published_at, NOW()),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: IncrementVideoViewCount :exec
UPDATE videos
SET view_count = view_count + 1
WHERE id = $1;

