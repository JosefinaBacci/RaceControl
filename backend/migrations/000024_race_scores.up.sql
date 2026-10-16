-- 000024_race_scores: the results as the API reads them.
--
-- Every result of an active race with what identifies it: the championship of
-- its round, the driver, the team credited at race time and the position. The
-- position is derived from the points (000006), so it is ranked here and never
-- stored: drivers on equal points share it. The race results, the driver
-- standings and the team standings all read from this one definition.

CREATE VIEW race_scores AS
SELECT
    se.event_id,
    e.starts_at,
    ro.season_id,
    s.year          AS season_year,
    ro.category_id,
    c.code          AS category_code,
    se.driver_id,
    d.first_name,
    d.last_name,
    d.country_code,
    sd.number       AS driver_number,
    se.team_id,
    t.code          AS team_code,
    t.name          AS team_name,
    se.points,
    RANK() OVER (PARTITION BY se.event_id ORDER BY se.points DESC)::integer AS position
FROM score_entries se
JOIN events e ON e.id = se.event_id
JOIN races r ON r.event_id = se.event_id
JOIN rounds ro ON ro.id = r.round_id
JOIN seasons s ON s.id = ro.season_id
JOIN categories c ON c.id = ro.category_id
JOIN drivers d ON d.id = se.driver_id
JOIN teams t ON t.id = se.team_id
LEFT JOIN season_drivers sd
    ON sd.driver_id = se.driver_id
   AND sd.season_id = ro.season_id
   AND sd.category_id = ro.category_id
WHERE e.is_active AND ro.is_active;
