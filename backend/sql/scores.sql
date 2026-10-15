-- Scores belong to a race (score_entries_race_fkey, 000016): an event without a
-- race has no grid. The audit trail (one score_change_log version plus its items)
-- is written by the service in the same transaction; the schema only demands an
-- actor through the guard of 000023.

-- name: GetRaceChampionship :one
SELECT c.id AS category_id, c.code AS category_code, c.name AS category_name,
       ro.season_id
FROM races r
JOIN rounds ro ON ro.id = r.round_id
JOIN categories c ON c.id = ro.category_id
WHERE r.event_id = $1;

-- The grid of a race: every driver whose team on the race date runs in its
-- championship, with the number they hold there and the points already loaded
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
   AND sd.season_id = (SELECT max(sd2.season_id)
                       FROM season_drivers sd2
                       WHERE sd2.driver_id = d.id AND sd2.category_id = ro.category_id)
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
-- rows through (000023). set_config(..., true) reverts at COMMIT/ROLLBACK.
-- name: SetScoreActor :exec
SELECT set_config('racecontrol.actor_id', sqlc.arg(actor_id)::text, true);

-- Driver standings of a championship: the general classification is derived from
-- the entries, so it reflects every change with no recalculation to run.
-- name: ListDriverStandings :many
SELECT d.id AS driver_id, d.first_name, d.last_name, d.country_code,
       t.id AS team_id, t.name AS team_name,
       SUM(se.points)::float8 AS points
FROM score_entries se
JOIN races r ON r.event_id = se.event_id
JOIN rounds ro ON ro.id = r.round_id
JOIN seasons s ON s.id = ro.season_id
JOIN categories c ON c.id = ro.category_id
JOIN drivers d ON d.id = se.driver_id
JOIN teams t ON t.id = se.team_id
WHERE c.code = sqlc.arg(category_code) AND s.year = sqlc.arg(season_year)
GROUP BY d.id, d.first_name, d.last_name, d.country_code, t.id, t.name
ORDER BY points DESC, d.last_name, d.first_name;
