-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, role, team_id, is_active, deactivated_at, created_at, updated_at
FROM users
WHERE username = lower($1);

-- name: GetUserByID :one
SELECT id, username, email, password_hash, role, team_id, is_active, deactivated_at, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, role, team_id)
VALUES (lower(sqlc.arg(username)), lower(sqlc.narg(email)), sqlc.arg(password_hash), sqlc.arg(role), sqlc.narg(team_id))
RETURNING id, username, email, password_hash, role, team_id, is_active, deactivated_at, created_at, updated_at;

-- name: CreateUserIfAbsent :exec
INSERT INTO users (username, email, password_hash, role, team_id)
VALUES (lower(sqlc.arg(username)), lower(sqlc.narg(email)), sqlc.arg(password_hash), sqlc.arg(role), sqlc.narg(team_id))
ON CONFLICT (username) DO NOTHING;

-- name: UpdateUserPasswordHash :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;
