-- 000021_calendar_read_model: the calendar as the API reads it.
--
-- An event is its own row plus a child of exactly one kind (000016), so listing
-- the calendar means reaching the same three children and taking the championship
-- from whichever one exists. The list and the detail of the calendar need the same
-- shape, so the join lives here once, as a read-only view.
--
-- The columns every kind shares (id, kind, timing, season, category) are not null
-- for every branch. The ones only a kind has (round, Grand Prix name, control
-- type, circuit) are null for the kinds that lack them, and the race branch is the
-- first of the UNION, so it is the one that declares their nullability: it reaches
-- its round and its circuit with a LEFT JOIN even though a race always has both.
-- That keeps the view's shape honest for every reader instead of borrowing the
-- race's NOT NULL.

CREATE VIEW calendar_events AS
SELECT
    e.id,
    e.kind,
    e.starts_at,
    e.ends_at,
    e.is_active,
    e.deactivated_at,
    ro.id           AS round_id,
    ro.number       AS round_number,
    ro.gp_name      AS gp_name,
    NULL::text      AS control_type,
    s.id            AS season_id,
    s.year          AS season_year,
    c.id            AS category_id,
    c.code          AS category_code,
    c.name          AS category_name,
    ci.id           AS circuit_id,
    ci.name         AS circuit_name,
    ci.country_code AS country_code
FROM events e
JOIN races r ON r.event_id = e.id
LEFT JOIN rounds ro ON ro.id = r.round_id
JOIN seasons s ON s.id = ro.season_id
JOIN categories c ON c.id = ro.category_id
LEFT JOIN circuits ci ON ci.id = ro.circuit_id

UNION ALL

SELECT
    e.id, e.kind, e.starts_at, e.ends_at, e.is_active, e.deactivated_at,
    NULL::bigint, NULL::integer, NULL::text, NULL::text,
    s.id, s.year, c.id, c.code, c.name,
    ci.id, ci.name, ci.country_code
FROM events e
JOIN tyre_tests tt ON tt.event_id = e.id
JOIN seasons s ON s.id = tt.season_id
JOIN categories c ON c.id = tt.category_id
JOIN circuits ci ON ci.id = tt.circuit_id

UNION ALL

SELECT
    e.id, e.kind, e.starts_at, e.ends_at, e.is_active, e.deactivated_at,
    NULL::bigint, NULL::integer, NULL::text, tc.control_type,
    s.id, s.year, c.id, c.code, c.name,
    NULL::bigint, NULL::text, NULL::text
FROM events e
JOIN technical_controls tc ON tc.event_id = e.id
JOIN seasons s ON s.id = tc.season_id
JOIN categories c ON c.id = tc.category_id;
