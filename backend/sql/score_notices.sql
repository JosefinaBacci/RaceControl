-- Score notices are derived from the change log by the score_notices view
-- (000025): a team is notified of every change whose items touched it, newest
-- first. The race is embedded from calendar_events, like in the results.

-- name: ListTeamScoreNotices :many
SELECT sqlc.embed(ce),
       n.change_log_id, n.changed_at, n.is_publication, n.acknowledged_at,
       acknowledger.username AS acknowledged_by
FROM score_notices n
JOIN calendar_events ce ON ce.id = n.event_id
LEFT JOIN users acknowledger ON acknowledger.id = n.acknowledged_by
WHERE n.team_id = sqlc.arg(team_id)
  AND (sqlc.narg(change_log_id)::bigint IS NULL OR n.change_log_id = sqlc.narg(change_log_id)::bigint)
ORDER BY n.changed_at DESC, n.change_log_id DESC;

-- name: ListTeamScoreChangeItems :many
SELECT i.change_log_id, i.driver_id, d.first_name, d.last_name,
       i.old_points, i.new_points
FROM score_change_log_items i
JOIN drivers d ON d.id = i.driver_id
WHERE i.team_id = sqlc.arg(team_id)
  AND (sqlc.narg(change_log_id)::bigint IS NULL OR i.change_log_id = sqlc.narg(change_log_id)::bigint)
ORDER BY i.change_log_id, d.last_name, d.first_name, i.driver_id;

-- The team is derived from the user by a trigger that also refuses a change the
-- team is not part of (000023, RC005). Acknowledging again keeps the first
-- confirmation, so the call is idempotent.
-- name: AcknowledgeScoreChange :exec
INSERT INTO score_acknowledgements (change_log_id, acknowledged_by)
VALUES (sqlc.arg(change_log_id), sqlc.arg(acknowledged_by))
ON CONFLICT (change_log_id, team_id) DO NOTHING;

-- name: ListScoreChangeAcknowledgements :many
SELECT sqlc.embed(ce),
       n.change_log_id, n.changed_at, n.is_publication,
       actor.username AS changed_by,
       t.id AS team_id, t.code AS team_code, t.name AS team_name,
       n.acknowledged_at,
       acknowledger.username AS acknowledged_by
FROM score_notices n
JOIN calendar_events ce ON ce.id = n.event_id
JOIN users actor ON actor.id = n.actor_user_id
JOIN teams t ON t.id = n.team_id
LEFT JOIN users acknowledger ON acknowledger.id = n.acknowledged_by
ORDER BY n.changed_at DESC, n.change_log_id DESC, t.name, t.id;
