-- Sample calendar, drivers and scores used by the demo and by the app outside
-- production. Idempotent like the reference data, and kept apart from it so a
-- production database starts with an empty calendar instead of made-up races.

INSERT INTO events (id, event_type_id, season_id, category_id, name, circuit, location, country_code, starts_at, ends_at, round)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 1, 1, 'Gran Premio de Singapur',         'Marina Bay',        'Singapur',        'sg', TIMESTAMPTZ '2026-10-04 15:00:00+00', TIMESTAMPTZ '2026-10-04 17:00:00+00', 18),
    (2, 3, 1, 1, 'Control técnico post-carrera',    'Marina Bay',        'Singapur',        'sg', TIMESTAMPTZ '2026-10-05 09:00:00+00', TIMESTAMPTZ '2026-10-05 12:00:00+00', NULL),
    (3, 1, 1, 1, 'Gran Premio de Estados Unidos',   'Circuit of the Americas', 'Austin',    'us', TIMESTAMPTZ '2026-10-18 16:00:00+00', TIMESTAMPTZ '2026-10-18 18:00:00+00', 19),
    (4, 2, 1, 1, 'Prueba de neumáticos Pirelli 2027', 'Circuit of the Americas', 'Austin',   'us', TIMESTAMPTZ '2026-10-20 09:00:00+00', TIMESTAMPTZ '2026-10-20 17:00:00+00', NULL),
    (5, 1, 1, 2, 'Ronda 13 — Qatar',                'Lusail',            'Lusail',          'qa', TIMESTAMPTZ '2026-11-27 15:00:00+00', TIMESTAMPTZ '2026-11-27 17:00:00+00', 13),
    (6, 1, 1, 3, 'Ronda 10 — Monza',                'Monza',             'Monza',           'it', TIMESTAMPTZ '2026-09-06 13:00:00+00', TIMESTAMPTZ '2026-09-06 15:00:00+00', 10),
    (7, 1, 1, 4, 'Ronda 7 — Las Vegas',             'Las Vegas Strip',   'Las Vegas',       'us', TIMESTAMPTZ '2026-11-21 18:00:00+00', TIMESTAMPTZ '2026-11-21 20:00:00+00', 7),
    (8, 3, 1, 2, 'Verificación de pesos',           'Lusail',            'Lusail',          'qa', TIMESTAMPTZ '2026-11-26 09:00:00+00', TIMESTAMPTZ '2026-11-26 12:00:00+00', NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO drivers (id, team_id, first_name, last_name, license_number, number, country_code, is_starter)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 'Charles',  'Leclerc',   'FIA-0001', 16, 'mc', true),
    (2, 1, 'Lewis',    'Hamilton',  'FIA-0002', 44, 'gb', true),
    (3, 1, 'Antonio',  'Giovinazzi', 'FIA-0003', 99, 'it', false),
    (4, 2, 'Max',      'Verstappen', 'FIA-0004',  1, 'nl', true),
    (5, 2, 'Yuki',     'Tsunoda',   'FIA-0005', 22, 'jp', true),
    (6, 3, 'George',   'Russell',   'FIA-0006', 63, 'gb', true),
    (7, 3, 'Andrea Kimi', 'Antonelli', 'FIA-0007', 12, 'it', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO score_entries (id, event_id, driver_id, team_id, position, points)
OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 1, 1, 1, 25),
    (2, 1, 4, 2, 2, 18),
    (3, 1, 6, 3, 3, 15),
    (4, 1, 2, 1, 4, 12)
ON CONFLICT (event_id, driver_id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('events',        'id'), COALESCE((SELECT max(id) FROM events),        0));
SELECT setval(pg_get_serial_sequence('drivers',       'id'), COALESCE((SELECT max(id) FROM drivers),       0));
SELECT setval(pg_get_serial_sequence('score_entries', 'id'), COALESCE((SELECT max(id) FROM score_entries), 0));
