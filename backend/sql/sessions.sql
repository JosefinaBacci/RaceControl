-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, absolute_expires_at, ip, user_agent)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetActiveSessionByTokenHash :one
SELECT s.id AS session_id, s.last_used_at, u.id AS user_id, u.username, u.role, u.team_id
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token_hash = sqlc.arg(token_hash)
  AND s.revoked_at IS NULL
  AND s.absolute_expires_at > sqlc.arg(now)
  AND s.last_used_at > sqlc.arg(idle_cutoff)
  AND u.is_active;

-- name: TouchSession :exec
UPDATE sessions
SET last_used_at = $2
WHERE id = $1;

-- name: RevokeSessionByTokenHash :exec
UPDATE sessions
SET revoked_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeUserSessions :exec
UPDATE sessions
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;
