-- name: GetChannelByID :one
SELECT * FROM channels WHERE id = $1;

-- name: GetChannelByHandle :one
SELECT * FROM channels WHERE handle = $1;

-- name: CreateChannel :one
INSERT INTO channels (owner_id, handle, name, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateChannelSubscriberCount :exec
UPDATE channels
SET subscriber_count = (SELECT COUNT(*) FROM subscriptions WHERE channel_id = $1)
WHERE id = $1;

