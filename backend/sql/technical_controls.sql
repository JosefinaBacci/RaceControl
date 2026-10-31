-- name: LockTechnicalControl :one
-- FOR SHARE holds the control's days while a result is recorded: a concurrent
-- move or cancellation waits for the result and is then refused by the guard
-- (RC020) instead of leaving the result outside the event.
SELECT e.id AS event_id, e.starts_at, e.ends_at, e.is_active, tc.category_id
FROM events e
JOIN technical_controls tc ON tc.event_id = e.id
WHERE e.id = $1
FOR SHARE OF e;

-- name: GetTechnicalControl :one
SELECT e.id AS event_id, e.starts_at, e.ends_at, e.is_active, tc.control_type,
       s.year AS season_year, c.code AS category_code, c.name AS category_name
FROM events e
JOIN technical_controls tc ON tc.event_id = e.id
JOIN seasons s ON s.id = tc.season_id
JOIN categories c ON c.id = tc.category_id
WHERE e.id = $1;

-- name: GetGradedTeam :one
SELECT id, category_id, is_active FROM teams WHERE id = $1;

-- name: GetTechnicalControlResultSubject :one
SELECT id, event_id, team_id FROM technical_control_results WHERE id = $1;

-- name: InsertTechnicalControlResult :one
INSERT INTO technical_control_results (
    event_id, category_id, team_id, conducted_on, passed, notes,
    supersedes_id, correction_note, recorded_by
) VALUES (
    sqlc.arg(event_id), sqlc.arg(category_id), sqlc.arg(team_id), sqlc.arg(conducted_on),
    sqlc.arg(passed), sqlc.arg(notes), sqlc.narg(supersedes_id), sqlc.narg(correction_note),
    sqlc.arg(recorded_by)
)
RETURNING id;

-- name: ListTechnicalControlHistory :many
SELECT * FROM technical_control_current_results
WHERE (sqlc.narg(from_day)::date IS NULL OR conducted_on >= sqlc.narg(from_day)::date)
  AND (sqlc.narg(until_day)::date IS NULL OR conducted_on < sqlc.narg(until_day)::date)
  AND (sqlc.narg(control_type)::text IS NULL OR control_type = sqlc.narg(control_type)::text)
  AND (sqlc.narg(category_code)::text IS NULL OR category_code = sqlc.narg(category_code)::text)
  AND (sqlc.narg(team_id)::bigint IS NULL OR team_id = sqlc.narg(team_id)::bigint)
ORDER BY conducted_on DESC, event_id DESC, team_name;

-- name: ListTechnicalControlTypes :many
SELECT DISTINCT control_type FROM technical_control_current_results ORDER BY control_type;

-- name: ListTechnicalControlRevisions :many
SELECT * FROM technical_control_revisions WHERE event_id = $1 ORDER BY team_name, id;

-- name: GetTechnicalControlRevision :one
SELECT * FROM technical_control_revisions WHERE id = $1;

-- name: ListControlledTeams :many
-- The teams a control grades: the active ones of its championship, plus any
-- team already graded there even if it was deactivated since.
SELECT t.id, t.name, t.is_active
FROM teams t
JOIN technical_controls tc ON tc.category_id = t.category_id
WHERE tc.event_id = $1
  AND (t.is_active OR EXISTS (
      SELECT 1 FROM technical_control_results r WHERE r.event_id = tc.event_id AND r.team_id = t.id
  ))
ORDER BY t.name;
