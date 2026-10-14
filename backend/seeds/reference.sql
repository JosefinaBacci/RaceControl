-- Reference data: the fixed catalogue the platform needs in every environment.
-- Idempotent by construction, so `make seed` can be re-run freely and never
-- duplicates a row nor rewinds an identity sequence below a runtime insert.

INSERT INTO categories (id, code, name) OVERRIDING SYSTEM VALUE VALUES
    (1, 'f1',      'Formula 1'),
    (2, 'f2',      'Formula 2'),
    (3, 'f3',      'Formula 3'),
    (4, 'academy', 'F1 Academy')
ON CONFLICT (code) DO NOTHING;

INSERT INTO seasons (id, year) OVERRIDING SYSTEM VALUE VALUES
    (1, 2026)
ON CONFLICT (year) DO NOTHING;

INSERT INTO teams (id, category_id, code, name) OVERRIDING SYSTEM VALUE VALUES
    (1, 1, 'ferrari',         'Scuderia Ferrari'),
    (2, 1, 'red_bull',        'Red Bull Racing'),
    (3, 1, 'mercedes',        'Mercedes-AMG Petronas'),
    (4, 1, 'mclaren',         'McLaren'),
    (5, 2, 'art',             'ART Grand Prix'),
    (6, 2, 'prema',           'PREMA Racing'),
    (7, 3, 'campos',          'Campos Racing'),
    (8, 4, 'academy_ferrari', 'Ferrari Driver Academy')
ON CONFLICT (category_id, code) DO NOTHING;

-- The circuits of the 2026 Formula 1 calendar. The feeder series race on the same
-- tracks, so one catalogue serves every category. The Grand Prix name is not here:
-- it belongs to the round.
INSERT INTO circuits (id, name, locality, country_code) OVERRIDING SYSTEM VALUE VALUES
    (1,  'Albert Park',                    'Melbourne',        'au'),
    (2,  'Shanghai International Circuit', 'Shanghái',         'cn'),
    (3,  'Suzuka',                         'Suzuka',           'jp'),
    (4,  'Bahrain International Circuit',  'Sakhir',           'bh'),
    (5,  'Jeddah Corniche Circuit',        'Yeda',             'sa'),
    (6,  'Miami International Autodrome',  'Miami',            'us'),
    (7,  'Circuit Gilles Villeneuve',      'Montreal',         'ca'),
    (8,  'Circuit de Monaco',              'Montecarlo',       'mc'),
    (9,  'Circuit de Barcelona-Catalunya', 'Montmeló',         'es'),
    (10, 'Red Bull Ring',                  'Spielberg',        'at'),
    (11, 'Silverstone',                    'Silverstone',      'gb'),
    (12, 'Spa-Francorchamps',              'Stavelot',         'be'),
    (13, 'Hungaroring',                    'Mogyoród',         'hu'),
    (14, 'Zandvoort',                      'Zandvoort',        'nl'),
    (15, 'Monza',                          'Monza',            'it'),
    (16, 'Madring',                        'Madrid',           'es'),
    (17, 'Baku City Circuit',              'Bakú',             'az'),
    (18, 'Marina Bay',                     'Singapur',         'sg'),
    (19, 'Circuit of the Americas',        'Austin',           'us'),
    (20, 'Autódromo Hermanos Rodríguez',   'Ciudad de México', 'mx'),
    (21, 'Interlagos',                     'São Paulo',        'br'),
    (22, 'Las Vegas Strip',                'Las Vegas',        'us'),
    (23, 'Lusail',                         'Lusail',           'qa'),
    (24, 'Yas Marina',                     'Abu Dabi',         'ae')
ON CONFLICT (name) DO NOTHING;

-- COALESCE keeps the empty-table case at 0 so the next insert starts at 1, and
-- taking the live max means a team created through the app is never overwritten.
SELECT setval(pg_get_serial_sequence('categories', 'id'), COALESCE((SELECT max(id) FROM categories), 0));
SELECT setval(pg_get_serial_sequence('seasons',    'id'), COALESCE((SELECT max(id) FROM seasons),    0));
SELECT setval(pg_get_serial_sequence('teams',      'id'), COALESCE((SELECT max(id) FROM teams),      0));
SELECT setval(pg_get_serial_sequence('circuits',   'id'), COALESCE((SELECT max(id) FROM circuits),   0));
