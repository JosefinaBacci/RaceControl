-- The calendar is read through the view that joins each event with its child
-- (000021), so the list and the detail share a single definition.

-- name: ListCalendarEvents :many
SELECT *
FROM calendar_events
WHERE is_active
  AND (sqlc.narg(category_code)::text IS NULL OR category_code = sqlc.narg(category_code)::text)
  AND (sqlc.narg(kind)::text IS NULL OR kind = sqlc.narg(kind)::text)
  AND (
      sqlc.narg(period)::text IS NULL
      OR (sqlc.narg(period)::text = 'past' AND ends_at < now())
      OR (sqlc.narg(period)::text = 'upcoming' AND ends_at >= now())
  )
ORDER BY starts_at;

-- name: GetCalendarEvent :one
SELECT *
FROM calendar_events
WHERE id = $1;

-- name: ListCircuits :many
SELECT id, name, locality, country_code
FROM circuits
WHERE is_active
ORDER BY name;

-- name: ListSeasons :many
SELECT id, year
FROM seasons
ORDER BY year DESC;

-- name: SeasonYear :one
SELECT year FROM seasons WHERE id = $1;

-- name: CategoryIDByCode :one
SELECT id FROM categories WHERE code = $1;

-- name: CircuitExists :one
SELECT EXISTS (SELECT 1 FROM circuits WHERE id = $1 AND is_active);

-- name: CreateEvent :one
INSERT INTO events (kind, starts_at, ends_at)
VALUES ($1, $2, $3)
RETURNING id;

-- name: CreateRound :one
INSERT INTO rounds (season_id, category_id, number, circuit_id, gp_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: CreateRace :exec
INSERT INTO races (event_id, round_id)
VALUES ($1, $2);

-- name: CreateTyreTest :exec
INSERT INTO tyre_tests (event_id, season_id, category_id, circuit_id)
VALUES ($1, $2, $3, $4);

-- name: CreateTechnicalControl :exec
INSERT INTO technical_controls (event_id, control_type, season_id, category_id)
VALUES ($1, $2, $3, $4);

-- name: UpdateEventTimes :exec
UPDATE events
SET starts_at = $2, ends_at = $3
WHERE id = $1;

-- name: UpdateRound :exec
UPDATE rounds
SET season_id = $2, category_id = $3, number = $4, circuit_id = $5, gp_name = $6
WHERE id = $1;

-- name: UpdateTyreTest :exec
UPDATE tyre_tests
SET season_id = $2, category_id = $3, circuit_id = $4
WHERE event_id = $1;

-- name: UpdateTechnicalControl :exec
UPDATE technical_controls
SET season_id = $2, category_id = $3, control_type = $4
WHERE event_id = $1;

-- name: DeactivateEvent :exec
UPDATE events
SET deactivated_at = now()
WHERE id = $1;

-- name: DeactivateRound :exec
UPDATE rounds
SET deactivated_at = now()
WHERE id = $1;
