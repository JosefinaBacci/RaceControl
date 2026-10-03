-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, role, team_id, is_active, deactivated_at, created_at, updated_at
FROM users
WHERE username = lower($1);

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

-- name: GetAccount :one
SELECT u.id, u.username, u.email, u.role, u.team_id, t.name AS team_name,
       u.is_active, u.deactivated_at, u.created_at, u.updated_at
FROM users u
LEFT JOIN teams t ON t.id = u.team_id
WHERE u.id = $1;

-- name: ListAccounts :many
SELECT u.id, u.username, u.email, u.role, u.team_id, t.name AS team_name,
       u.is_active, u.deactivated_at, u.created_at, u.updated_at
FROM users u
LEFT JOIN teams t ON t.id = u.team_id
WHERE (sqlc.narg(search)::text IS NULL
       OR u.username LIKE '%' || sqlc.narg(search) || '%'
       OR u.email LIKE '%' || sqlc.narg(search) || '%')
  AND (sqlc.narg(role)::text IS NULL OR u.role = sqlc.narg(role))
  AND (sqlc.narg(is_active)::boolean IS NULL OR u.is_active = sqlc.narg(is_active))
ORDER BY u.username;

-- name: UpdateAccount :exec
UPDATE users
SET email = lower(sqlc.narg(email)), role = sqlc.arg(role), team_id = sqlc.narg(team_id), password_hash = sqlc.arg(password_hash)
WHERE id = sqlc.arg(id);

-- name: DeactivateUser :exec
UPDATE users
SET deactivated_at = now()
WHERE id = $1 AND deactivated_at IS NULL;

-- name: ReactivateUser :exec
UPDATE users
SET deactivated_at = NULL
WHERE id = $1;
