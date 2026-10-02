-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, role, team_id, is_active, deactivated_at, created_at, updated_at
FROM users
WHERE username = lower($1);

-- name: GetUserByID :one
SELECT id, username, email, password_hash, role, team_id, is_active, deactivated_at, created_at, updated_at
FROM users
WHERE id = $1;