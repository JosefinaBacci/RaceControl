-- Scores belong to a race (score_entries_race_fkey, 000016): an event without a
-- race has no grid. The audit trail (one score_change_log version plus its items)
-- is written by the service in the same transaction; the schema only demands an
-- actor through the guard of 000026.

-- A cancelled race has no grid: it never ran, so it takes no results.
-- name: GetRaceChampionship :one
SELECT c.id AS category_id, c.code AS category_code, c.name AS category_name,
       ro.season_id, e.ends_at
FROM races r
JOIN events e ON e.id = r.event_id
JOIN rounds ro ON ro.id = r.round_id
JOIN categories c ON c.id = ro.category_id
WHERE r.event_id = $1 AND e.is_active AND ro.is_active;

-- Two edits of the same race take turns, so each one diffs against the scores
-- the previous one left and the audit records the real old points.
-- name: LockRaceScores :exec
SELECT 1 FROM races WHERE event_id = $1 FOR NO KEY UPDATE;

-- The grid of a race: every driver whose team on the race date runs in its
-- championship, with the number they hold in that season and the points already loaded
-- (NULL when the driver has no entry yet). The service diffs against this.
-- name: ListRaceEligibleDrivers :many
SELECT d.id, d.first_name, d.last_name, d.country_code,
       t.id AS team_id, t.name AS team_name, sd.number, se.points
FROM drivers d
JOIN races r ON r.event_id = sqlc.arg(event_id)
JOIN rounds ro ON ro.id = r.round_id
JOIN LATERAL (
    SELECT a.team_id
    FROM driver_team_assignments a
    WHERE a.driver_id = d.id
      AND a.effective_from <= (SELECT e.starts_at FROM events e WHERE e.id = r.event_id)
    ORDER BY a.effective_from DESC
    LIMIT 1
) a ON true
JOIN teams t ON t.id = a.team_id
LEFT JOIN season_drivers sd
    ON sd.driver_id = d.id
   AND sd.category_id = ro.category_id
   AND sd.season_id = ro.season_id
LEFT JOIN score_entries se ON se.event_id = r.event_id AND se.driver_id = d.id
WHERE d.is_active AND t.category_id = ro.category_id
ORDER BY t.name, d.last_name, d.first_name;

-- name: UpsertScoreEntry :exec
INSERT INTO score_entries (event_id, driver_id, team_id, points)
VALUES (sqlc.arg(event_id), sqlc.arg(driver_id), 0, sqlc.arg(points))
ON CONFLICT (event_id, driver_id) DO UPDATE SET points = EXCLUDED.points;

-- name: DeleteScoreEntry :exec
DELETE FROM score_entries
WHERE event_id = sqlc.arg(event_id) AND driver_id = sqlc.arg(driver_id);

-- name: CreateScoreChangeLog :one
INSERT INTO score_change_log (event_id, actor_user_id)
VALUES (sqlc.arg(event_id), sqlc.arg(actor_user_id))
RETURNING id;

-- name: CreateScoreChangeLogItem :exec
INSERT INTO score_change_log_items (change_log_id, driver_id, old_points, new_points)
VALUES (sqlc.arg(change_log_id), sqlc.arg(driver_id), sqlc.narg(old_points), sqlc.narg(new_points));

-- The actor travels as a transaction-local setting so the audit guard lets the
-- rows through (000026). set_config(..., true) reverts at COMMIT/ROLLBACK.
-- name: SetScoreActor :exec
SELECT set_config('racecontrol.actor_id', sqlc.arg(actor_id)::text, true);

-- Results and standings read the race_scores view (000024), so the position and
-- the championship of a result have a single definition. The race each row
-- belongs to is embedded from calendar_events (000021), which already composes
-- what identifies a race. Points leave the database as float8 only after they
-- were summed as numeric, so the totals are exact before the one conversion.

-- name: ListRaceResults :many
SELECT sqlc.embed(ce),
       rs.driver_id, rs.first_name, rs.last_name, rs.country_code AS driver_country_code,
       rs.driver_number, rs.team_id, rs.team_code, rs.team_name,
       rs.points::float8 AS points, rs.position
FROM race_scores rs
JOIN calendar_events ce ON ce.id = rs.event_id
WHERE (sqlc.narg(season_year)::integer IS NULL OR rs.season_year = sqlc.narg(season_year)::integer)
  AND (sqlc.narg(category_code)::text IS NULL OR rs.category_code = sqlc.narg(category_code)::text)
  AND (sqlc.narg(event_id)::bigint IS NULL OR rs.event_id = sqlc.narg(event_id)::bigint)
ORDER BY ce.starts_at, ce.id, rs.position, rs.last_name, rs.first_name, rs.driver_id;

-- A driver is listed under the team credited in their latest race of the
-- championship, the one they currently score for.
-- name: DriverStandings :many
SELECT driver_id, first_name, last_name, country_code, driver_number,
       (array_agg(team_id ORDER BY starts_at DESC))[1]::bigint AS team_id,
       (array_agg(team_code ORDER BY starts_at DESC))[1]::text AS team_code,
       (array_agg(team_name ORDER BY starts_at DESC))[1]::text AS team_name,
       sum(points)::float8 AS points,
       count(*)::integer AS races,
       RANK() OVER (ORDER BY sum(points) DESC)::integer AS position
FROM race_scores
WHERE season_year = sqlc.arg(season_year) AND category_code = sqlc.arg(category_code)
GROUP BY driver_id, first_name, last_name, country_code, driver_number
ORDER BY position, last_name, first_name, driver_id;

-- name: TeamStandings :many
SELECT team_id, team_code, team_name,
       sum(points)::float8 AS points,
       RANK() OVER (ORDER BY sum(points) DESC)::integer AS position
FROM race_scores
WHERE season_year = sqlc.arg(season_year) AND category_code = sqlc.arg(category_code)
GROUP BY team_id, team_code, team_name
ORDER BY position, team_name, team_id;
