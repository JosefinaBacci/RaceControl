-- Sample calendar, drivers, team assignments and scores used by the demo and by
-- the app outside production. Idempotent like the reference data, and kept apart
-- from it so a production database starts with an empty calendar instead of
-- made-up races.

-- 1. rounds: the dates of each championship, with the circuit and the Grand Prix name
INSERT INTO rounds (id, season_id, category_id, number, circuit_id, gp_name)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 1, 18, 18, 'Gran Premio de Singapur'),
    (2, 1, 1, 19, 19, 'Gran Premio de Estados Unidos'),
    (3, 1, 2, 13, 23, 'Gran Premio de Qatar'),
    (4, 1, 3, 10, 15, 'Gran Premio de Italia'),
    (5, 1, 4, 7,  22, 'Gran Premio de Las Vegas')
ON CONFLICT (id) DO NOTHING;

-- 2. events: kind and timing. The children repeat the kind by DEFAULT and their FK
-- forces it to match. No name is stored: the backend composes the label.
INSERT INTO events (id, kind, starts_at, ends_at)
OVERRIDING SYSTEM VALUE VALUES
    (1, 'race',              TIMESTAMPTZ '2026-10-04 15:00:00+00', TIMESTAMPTZ '2026-10-04 17:00:00+00'),
    (2, 'technical_control', TIMESTAMPTZ '2026-10-05 09:00:00+00', TIMESTAMPTZ '2026-10-05 12:00:00+00'),
    (3, 'race',              TIMESTAMPTZ '2026-10-18 16:00:00+00', TIMESTAMPTZ '2026-10-18 18:00:00+00'),
    (4, 'tyre_test',         TIMESTAMPTZ '2026-10-20 09:00:00+00', TIMESTAMPTZ '2026-10-20 17:00:00+00'),
    (5, 'race',              TIMESTAMPTZ '2026-11-27 15:00:00+00', TIMESTAMPTZ '2026-11-27 17:00:00+00'),
    (6, 'race',              TIMESTAMPTZ '2026-09-06 13:00:00+00', TIMESTAMPTZ '2026-09-06 15:00:00+00'),
    (7, 'race',              TIMESTAMPTZ '2026-11-21 18:00:00+00', TIMESTAMPTZ '2026-11-21 20:00:00+00'),
    (8, 'technical_control', TIMESTAMPTZ '2026-11-26 09:00:00+00', TIMESTAMPTZ '2026-11-26 12:00:00+00')
ON CONFLICT (id) DO NOTHING;

-- 3. races: the race of each round; the championship comes from the round
INSERT INTO races (event_id, round_id) VALUES
    (1, 1),
    (3, 2),
    (5, 3),
    (6, 4),
    (7, 5)
ON CONFLICT DO NOTHING;

-- 4. tyre_tests and technical_controls own their championship
INSERT INTO tyre_tests (event_id, season_id, category_id) VALUES
    (4, 1, 1)
ON CONFLICT DO NOTHING;

INSERT INTO technical_controls (event_id, season_id, category_id, control_type) VALUES
    (2, 1, 1, 'Control técnico post-carrera'),
    (8, 1, 2, 'Verificación de pesos')
ON CONFLICT DO NOTHING;

-- 5. drivers (unchanged)
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

-- 6. driver_team_assignments (unchanged)
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

-- 7. season_drivers (car number per driver-season)
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

-- 8. score_entries (unchanged)
INSERT INTO score_entries (id, event_id, driver_id, points)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 1, 25),
    (2, 1, 4, 18),
    (3, 1, 6, 15),
    (4, 1, 2, 12)
ON CONFLICT (event_id, driver_id) DO NOTHING;

-- 9. setval for every sequence
SELECT setval(pg_get_serial_sequence('rounds',                 'id'), COALESCE((SELECT max(id) FROM rounds),                 0));
SELECT setval(pg_get_serial_sequence('events',                 'id'), COALESCE((SELECT max(id) FROM events),                 0));
SELECT setval(pg_get_serial_sequence('drivers',                'id'), COALESCE((SELECT max(id) FROM drivers),                0));
SELECT setval(pg_get_serial_sequence('season_drivers',   'id'), COALESCE((SELECT max(id) FROM season_drivers),   0));
SELECT setval(pg_get_serial_sequence('driver_team_assignments', 'id'), COALESCE((SELECT max(id) FROM driver_team_assignments), 0));
SELECT setval(pg_get_serial_sequence('score_entries',          'id'), COALESCE((SELECT max(id) FROM score_entries),          0));
