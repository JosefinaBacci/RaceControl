-- Sanctions (US15, US25, US26). A sanction is its identity in sanctions plus an
-- append-only chain of versions (000035); the reads go through the
-- sanction_version_details and sanction_notices views (000037).

-- name: ListPenaltyTypes :many
SELECT id, code, name, unit, max_value
FROM penalty_types
ORDER BY id;

-- name: GetPenaltyTypeByCode :one
SELECT id, code, name, unit, max_value
FROM penalty_types
WHERE code = $1;

-- A team filter matches the accountable team, so it also finds the sanctions of
-- the team's drivers. The days of the range are already resolved to instants.
-- name: ListCurrentSanctions :many
SELECT *
FROM sanction_version_details
WHERE is_current
  AND (sqlc.narg(driver_id)::bigint IS NULL OR driver_id = sqlc.narg(driver_id)::bigint)
  AND (sqlc.narg(team_id)::bigint IS NULL OR accountable_team_id = sqlc.narg(team_id)::bigint)
  AND (sqlc.narg(category_code)::text IS NULL OR category_code = sqlc.narg(category_code)::text)
  AND (sqlc.narg(is_revoked)::boolean IS NULL OR (action = 'revoked') = sqlc.narg(is_revoked)::boolean)
  AND (sqlc.narg(from_at)::timestamptz IS NULL OR issued_at >= sqlc.narg(from_at)::timestamptz)
  AND (sqlc.narg(until_at)::timestamptz IS NULL OR issued_at < sqlc.narg(until_at)::timestamptz)
ORDER BY issued_at DESC, sanction_id DESC;

-- name: ListSanctionVersions :many
SELECT *
FROM sanction_version_details
WHERE sanction_id = $1
ORDER BY changed_at, version_id;

-- Amendments and revocations of one sanction queue behind this lock, so each one
-- reads the version it replaces and none is lost.
-- name: LockSanction :one
SELECT id FROM sanctions WHERE id = $1 FOR UPDATE;

-- name: GetCurrentSanctionVersion :one
SELECT version_id, action, accountable_team_id, penalty_type_id, penalty_value, reason, event_id
FROM current_sanctions
WHERE id = $1;

-- name: ActiveDriverExists :one
SELECT EXISTS (SELECT 1 FROM drivers WHERE id = $1 AND is_active);

-- name: ActiveTeamExists :one
SELECT EXISTS (SELECT 1 FROM teams WHERE id = $1 AND is_active);

-- An event can only frame a sanction of its own championship.
-- name: ActiveEventInTeamCategory :one
SELECT EXISTS (
    SELECT 1
    FROM calendar_events ce
    JOIN teams t ON t.category_id = ce.category_id
    WHERE ce.id = sqlc.arg(event_id) AND ce.is_active AND t.id = sqlc.arg(team_id)
);

-- The accountable team is set by a trigger (000035, RC011 for a teamless driver).
-- name: CreateSanction :one
INSERT INTO sanctions (driver_id, team_id)
VALUES (sqlc.narg(driver_id), sqlc.narg(team_id))
RETURNING id, accountable_team_id;

-- The recipients of the version are written by a trigger (000036).
-- name: CreateSanctionVersion :one
INSERT INTO sanction_versions (sanction_id, action, penalty_type_id, penalty_value, reason, event_id, note, actor_user_id)
VALUES (sqlc.arg(sanction_id), sqlc.arg(action), sqlc.arg(penalty_type_id), sqlc.narg(penalty_value),
        sqlc.arg(reason), sqlc.narg(event_id), sqlc.narg(note), sqlc.arg(actor_user_id))
RETURNING id;

-- A team sees the version that stands of every sanction it was sent.
-- name: ListTeamSanctionNotices :many
SELECT sqlc.embed(details), n.acknowledged_at, acknowledger.username AS acknowledged_by
FROM sanction_notices n
JOIN sanction_version_details details ON details.version_id = n.version_id
LEFT JOIN users acknowledger ON acknowledger.id = n.acknowledged_by
WHERE n.team_id = sqlc.arg(team_id)
  AND n.is_current
  AND (sqlc.narg(version_id)::bigint IS NULL OR n.version_id = sqlc.narg(version_id)::bigint)
ORDER BY details.changed_at DESC, details.version_id DESC;

-- The team comes from the user, and a version the team was not sent (RC013) or
-- that was replaced (RC014) is refused (000036). Confirming again keeps the
-- first confirmation.
-- name: AcknowledgeSanctionVersion :exec
INSERT INTO sanction_acknowledgements (sanction_version_id, acknowledged_by)
VALUES (sqlc.arg(version_id), sqlc.arg(acknowledged_by))
ON CONFLICT (sanction_version_id, team_id) DO NOTHING;

-- Every version of every sanction with each recipient team and its confirmation,
-- for the FIA's follow-up.
-- name: ListSanctionAcknowledgements :many
SELECT sqlc.embed(details),
       actor.username AS changed_by,
       t.id AS recipient_team_id, t.code AS recipient_team_code, t.name AS recipient_team_name,
       n.acknowledged_at,
       acknowledger.username AS acknowledged_by
FROM sanction_notices n
JOIN sanction_version_details details ON details.version_id = n.version_id
JOIN users actor ON actor.id = details.actor_user_id
JOIN teams t ON t.id = n.team_id
LEFT JOIN users acknowledger ON acknowledger.id = n.acknowledged_by
ORDER BY details.issued_at DESC, details.sanction_id DESC,
         details.changed_at DESC, details.version_id DESC, t.name, t.id;
