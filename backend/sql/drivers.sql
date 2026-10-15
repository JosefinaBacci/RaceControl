-- The lineup of a team: the drivers whose latest assignment points at that team
-- (000007, 000022). A release makes the latest assignment teamless, so the driver
-- drops out of the list. The number is the one the driver holds in the team
-- category for the latest season (000011, 000019); a scalar subquery keeps it
-- nullable for a driver without a number yet.

-- name: ListTeamDrivers :many
SELECT d.id, d.first_name, d.last_name, d.license_number, d.country_code,
       a.is_starter, sd.number
FROM drivers d
JOIN LATERAL (
    SELECT a.is_starter, a.team_id
    FROM driver_team_assignments a
    WHERE a.driver_id = d.id AND a.effective_from <= now()
    ORDER BY a.effective_from DESC
    LIMIT 1
) a ON a.team_id = sqlc.arg(team_id)
JOIN teams t ON t.id = a.team_id
LEFT JOIN season_drivers sd
    ON sd.driver_id = d.id
   AND sd.category_id = t.category_id
   AND sd.season_id = (SELECT max(sd2.season_id)
                       FROM season_drivers sd2
                       WHERE sd2.driver_id = d.id AND sd2.category_id = t.category_id)
WHERE d.is_active
ORDER BY a.is_starter DESC, d.last_name, d.first_name;

-- name: GetTeamDriver :one
SELECT d.id, d.first_name, d.last_name, d.license_number, d.country_code,
       a.is_starter, sd.number
FROM drivers d
JOIN LATERAL (
    SELECT a.is_starter, a.team_id
    FROM driver_team_assignments a
    WHERE a.driver_id = d.id AND a.effective_from <= now()
    ORDER BY a.effective_from DESC
    LIMIT 1
) a ON a.team_id = sqlc.arg(team_id)
JOIN teams t ON t.id = a.team_id
LEFT JOIN season_drivers sd
    ON sd.driver_id = d.id
   AND sd.category_id = t.category_id
   AND sd.season_id = (SELECT max(sd2.season_id)
                       FROM season_drivers sd2
                       WHERE sd2.driver_id = d.id AND sd2.category_id = t.category_id)
WHERE d.id = sqlc.arg(driver_id) AND d.is_active;

-- The public lineup (US12): every driver whose latest assignment points at a team,
-- with the team they belong to and their number there. A released driver has no
-- team in their latest assignment, so the join leaves them out. Search runs over
-- the full name and the category filter runs over the team's championship.

-- name: ListActiveDrivers :many
SELECT d.id, d.first_name, d.last_name, d.country_code,
       a.is_starter, t.id AS team_id, t.name AS team_name,
       c.code AS category_code, c.name AS category_name, sd.number
FROM drivers d
JOIN LATERAL (
    SELECT a.is_starter, a.team_id
    FROM driver_team_assignments a
    WHERE a.driver_id = d.id AND a.effective_from <= now()
    ORDER BY a.effective_from DESC
    LIMIT 1
) a ON true
JOIN teams t ON t.id = a.team_id
JOIN categories c ON c.id = t.category_id
LEFT JOIN season_drivers sd
    ON sd.driver_id = d.id
   AND sd.category_id = c.id
   AND sd.season_id = (SELECT max(sd2.season_id)
                       FROM season_drivers sd2
                       WHERE sd2.driver_id = d.id AND sd2.category_id = c.id)
WHERE d.is_active
  AND (sqlc.narg(category_code)::text IS NULL OR c.code = sqlc.narg(category_code)::text)
  AND (sqlc.narg(search)::text IS NULL
       OR (d.first_name || ' ' || d.last_name) ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY c.id, t.name, a.is_starter DESC, d.last_name, d.first_name;

-- name: GetActiveDriver :one
SELECT d.id, d.first_name, d.last_name, d.country_code,
       a.is_starter, t.id AS team_id, t.name AS team_name,
       c.code AS category_code, c.name AS category_name, sd.number
FROM drivers d
JOIN LATERAL (
    SELECT a.is_starter, a.team_id
    FROM driver_team_assignments a
    WHERE a.driver_id = d.id AND a.effective_from <= now()
    ORDER BY a.effective_from DESC
    LIMIT 1
) a ON true
JOIN teams t ON t.id = a.team_id
JOIN categories c ON c.id = t.category_id
LEFT JOIN season_drivers sd
    ON sd.driver_id = d.id
   AND sd.category_id = c.id
   AND sd.season_id = (SELECT max(sd2.season_id)
                       FROM season_drivers sd2
                       WHERE sd2.driver_id = d.id AND sd2.category_id = c.id)
WHERE d.id = sqlc.arg(driver_id) AND d.is_active;

-- name: CreateDriver :one
INSERT INTO drivers (first_name, last_name, license_number, country_code)
VALUES (sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(license_number), sqlc.arg(country_code))
RETURNING id;

-- name: UpdateDriver :exec
UPDATE drivers
SET first_name = sqlc.arg(first_name),
    last_name = sqlc.arg(last_name),
    license_number = sqlc.arg(license_number),
    country_code = sqlc.arg(country_code)
WHERE id = sqlc.arg(id);

-- name: TeamCategory :one
SELECT category_id FROM teams WHERE id = $1;

-- name: LatestSeason :one
SELECT id, year FROM seasons ORDER BY year DESC LIMIT 1;

-- name: CreateAssignment :exec
INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
VALUES (sqlc.arg(driver_id), sqlc.arg(team_id), now(), sqlc.arg(is_starter));

-- name: ReleaseAssignment :exec
INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
VALUES (sqlc.arg(driver_id), NULL, now(), false);

-- name: UpsertSeasonDriver :exec
INSERT INTO season_drivers (driver_id, season_id, category_id, number)
VALUES (sqlc.arg(driver_id), sqlc.arg(season_id), sqlc.arg(category_id), sqlc.arg(number))
ON CONFLICT (driver_id, season_id, category_id) DO UPDATE SET number = EXCLUDED.number;
