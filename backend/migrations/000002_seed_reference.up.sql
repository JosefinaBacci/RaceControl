-- Seed data with deterministic IDs so every developer and every environment shares
-- the same identifiers, which keeps demos reproducible. The identity sequences are
-- reset afterwards so later inserts never collide with the seeded rows.

INSERT INTO categories (id, code, name) OVERRIDING SYSTEM VALUE VALUES
    (1, 'f1',        'Formula 1'),
    (2, 'f2',        'Formula 2'),
    (3, 'f3',        'Formula 3'),
    (4, 'academy',   'F1 Academy');

INSERT INTO seasons (id, year) OVERRIDING SYSTEM VALUE VALUES
    (1, 2026);

INSERT INTO teams (id, category_id, code, name) OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 'ferrari', 'Scuderia Ferrari'),
    (2, 1, 'red_bull', 'Red Bull Racing'),
    (3, 1, 'mercedes', 'Mercedes-AMG Petronas'),
    (4, 1, 'mclaren', 'McLaren'),
    (5, 2, 'art',     'ART Grand Prix'),
    (6, 2, 'prema',   'PREMA Racing'),
    (7, 3, 'campos',  'Campos Racing'),
    (8, 4, 'academy_ferrari', 'Ferrari Driver Academy');

SELECT setval(pg_get_serial_sequence('categories', 'id'), (SELECT max(id) FROM categories));
SELECT setval(pg_get_serial_sequence('seasons', 'id'), (SELECT max(id) FROM seasons));
SELECT setval(pg_get_serial_sequence('teams', 'id'), (SELECT max(id) FROM teams));