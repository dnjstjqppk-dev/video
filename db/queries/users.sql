-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserProfile :one
SELECT u.id, u.email, u.username, u.role, u.subscription_tier, u.is_verified,
       p.display_name, p.bio, p.avatar_url, p.banner_url, p.language
FROM users u
LEFT JOIN user_profiles p ON p.user_id = u.id
WHERE u.id = $1;

-- name: CreateUser :one
INSERT INTO users (email, username, password_hash, role, subscription_tier)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateUserProfile :one
INSERT INTO user_profiles (user_id, display_name, bio, avatar_url, language)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListUsers :many
SELECT u.id, u.email, u.username, u.role, u.subscription_tier, u.is_active, u.created_at,
       p.display_name, p.avatar_url
FROM users u
LEFT JOIN user_profiles p ON p.user_id = u.id
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUserStatus :one
UPDATE users
SET is_active = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

