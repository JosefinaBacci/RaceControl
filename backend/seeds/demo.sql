-- Sample calendar, drivers, team assignments and scores used by the demo and by
-- the app outside production. Idempotent like the reference data, and kept apart
-- from it so a production database starts with an empty calendar instead of
-- made-up races.

-- 1. events: el calendario oficial de GP/ronda (ya no hay circuit/location/country_code/round;
--    estos datos viven en la tabla races hija)
INSERT INTO events (id, season_id, category_id, name, starts_at, ends_at)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 1, 'Gran Premio de Singapur', TIMESTAMPTZ '2026-10-04 15:00:00+00', TIMESTAMPTZ '2026-10-04 17:00:00+00'),
    (2, 1, 1, 'Control técnico post-carrera', TIMESTAMPTZ '2026-10-05 09:00:00+00', TIMESTAMPTZ '2026-10-05 12:00:00+00'),
    (3, 1, 1, 'Gran Premio de Estados Unidos', TIMESTAMPTZ '2026-10-18 16:00:00+00', TIMESTAMPTZ '2026-10-18 18:00:00+00'),
    (4, 1, 1, 'Prueba de neumáticos Pirelli 2027', TIMESTAMPTZ '2026-10-20 09:00:00+00', TIMESTAMPTZ '2026-10-20 17:00:00+00'),
    (5, 1, 2, 'Ronda 13 — Qatar', TIMESTAMPTZ '2026-11-27 15:00:00+00', TIMESTAMPTZ '2026-11-27 17:00:00+00'),
    (6, 1, 3, 'Ronda 10 — Monza', TIMESTAMPTZ '2026-09-06 13:00:00+00', TIMESTAMPTZ '2026-09-06 15:00:00+00'),
    (7, 1, 4, 'Ronda 7 — Las Vegas', TIMESTAMPTZ '2026-11-21 18:00:00+00', TIMESTAMPTZ '2026-11-21 20:00:00+00'),
    (8, 1, 2, 'Verificación de pesos', TIMESTAMPTZ '2026-11-26 09:00:00+00', TIMESTAMPTZ '2026-11-26 12:00:00+00')
ON CONFLICT (id) DO NOTHING;

-- 2. races: la carrera concreta, con los datos de circuito antes en events
INSERT INTO races (id, event_id, name, circuit, location, country_code, round, starts_at, ends_at)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 'Gran Premio de Singapur', 'Marina Bay', 'Singapore', 'sg', 18, TIMESTAMPTZ '2026-10-04 15:00:00+00', TIMESTAMPTZ '2026-10-04 17:00:00+00'),
    (2, 3, 'Gran Premio de Estados Unidos', 'Circuit of the Americas', 'Austin', 'us', 19, TIMESTAMPTZ '2026-10-18 16:00:00+00', TIMESTAMPTZ '2026-10-18 18:00:00+00'),
    (3, 5, 'Ronda 13 — Qatar', 'Lusail', 'Lusail', 'qa', 13, TIMESTAMPTZ '2026-11-27 15:00:00+00', TIMESTAMPTZ '2026-11-27 17:00:00+00'),
    (4, 6, 'Ronda 10 — Monza', 'Monza', 'Monza', 'it', 10, TIMESTAMPTZ '2026-09-06 13:00:00+00', TIMESTAMPTZ '2026-09-06 15:00:00+00'),
    (5, 7, 'Ronda 7 — Las Vegas', 'Las Vegas Strip', 'Las Vegas', 'us', 7, TIMESTAMPTZ '2026-11-21 18:00:00+00', TIMESTAMPTZ '2026-11-21 20:00:00+00')
ON CONFLICT DO NOTHING;

-- 3. tyre_tests: prueba de neumáticos (suelta o bajo GP)
INSERT INTO tyre_tests (id, event_id, name, starts_at, ends_at)
OVERRIDING SYSTEM VALUE VALUES
    (1, 4, 'Prueba de neumáticos Pirelli 2027', TIMESTAMPTZ '2026-10-20 09:00:00+00', TIMESTAMPTZ '2026-10-20 17:00:00+00')
ON CONFLICT DO NOTHING;

-- 4. technical_controls: control técnico (suelto o bajo GP)
INSERT INTO technical_controls (id, event_id, name, control_type, started_at, ended_at)
OVERRIDING SYSTEM VALUE VALUES
    (1, 2, 'Control técnico post-carrera', 'post-race', TIMESTAMPTZ '2026-10-05 09:00:00+00', TIMESTAMPTZ '2026-10-05 12:00:00+00'),
    (2, 8, 'Verificación de pesos', 'weight-check', TIMESTAMPTZ '2026-11-26 09:00:00+00', TIMESTAMPTZ '2026-11-26 12:00:00+00')
ON CONFLICT DO NOTHING;

-- 5. drivers (sin cambios)
INSERT INTO drivers (id, first_name, last_name, license_number, country_code)
OVERRIDING SYSTEM VALUE VALUES
    (1, 'Charles',  'Leclerc',    'FIA-0001', 'mc'),
    (2, 'Lewis',    'Hamilton',   'FIA-0002', 'gb'),
    (3, 'Antonio',  'Giovinazzi', 'FIA-0003', 'it'),
    (4, 'Max',      'Verstappen', 'FIA-0004', 'nl'),
    (5, 'Yuki',     'Tsunoda',    'FIA-0005', 'jp'),
    (6, 'George',   'Russell',    'FIA-0006', 'gb'),
    (7, 'Andrea Kimi', 'Antonelli', 'FIA-0007', 'it')
ON CONFLICT (id) DO NOTHING;

-- 6. driver_team_assignments (sin cambios)
INSERT INTO driver_team_assignments (id, driver_id, team_id, effective_from, is_starter)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 1, TIMESTAMPTZ '2026-01-01 00:00:00+00', true),
    (2, 2, 1, TIMESTAMPTZ '2026-01-01 00:00:00+00', true),
    (3, 3, 1, TIMESTAMPTZ '2026-01-01 00:00:00+00', false),
    (4, 4, 2, TIMESTAMPTZ '2026-01-01 00:00:00+00', true),
    (5, 5, 2, TIMESTAMPTZ '2026-01-01 00:00:00+00', true),
    (6, 6, 3, TIMESTAMPTZ '2026-01-01 00:00:00+00', true),
    (7, 7, 3, TIMESTAMPTZ '2026-01-01 00:00:00+00', true)
ON CONFLICT DO NOTHING;

-- 7. season_drivers (número de coche por piloto-temporada)
INSERT INTO season_drivers (id, driver_id, season_id, number)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 1, 16),
    (2, 2, 1, 44),
    (3, 3, 1, 99),
    (4, 4, 1, 1),
    (5, 5, 1, 22),
    (6, 6, 1, 63),
    (7, 7, 1, 12)
ON CONFLICT DO NOTHING;

-- 8. score_entries (sin cambios)
INSERT INTO score_entries (id, event_id, driver_id, points)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 1, 25),
    (2, 1, 4, 18),
    (3, 1, 6, 15),
    (4, 1, 2, 12)
ON CONFLICT (event_id, driver_id) DO NOTHING;

-- 9. setval para todas las secuencias (incluye season_drivers)
SELECT setval(pg_get_serial_sequence('events',                 'id'), COALESCE((SELECT max(id) FROM events),                 0));
SELECT setval(pg_get_serial_sequence('drivers',                'id'), COALESCE((SELECT max(id) FROM drivers),                0));
SELECT setval(pg_get_serial_sequence('season_drivers',   'id'), COALESCE((SELECT max(id) FROM season_drivers),   0));
SELECT setval(pg_get_serial_sequence('driver_team_assignments', 'id'), COALESCE((SELECT max(id) FROM driver_team_assignments), 0));
SELECT setval(pg_get_serial_sequence('score_entries',          'id'), COALESCE((SELECT max(id) FROM score_entries),          0));
SELECT setval(pg_get_serial_sequence('races',                 'id'), COALESCE((SELECT max(id) FROM races),                 0));
SELECT setval(pg_get_serial_sequence('tyre_tests',            'id'), COALESCE((SELECT max(id) FROM tyre_tests),            0));
SELECT setval(pg_get_serial_sequence('technical_controls',    'id'), COALESCE((SELECT max(id) FROM technical_controls),    0));
