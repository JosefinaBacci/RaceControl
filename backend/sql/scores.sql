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
