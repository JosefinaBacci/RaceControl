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
-- tracks, so one catalogue serves every category.
INSERT INTO circuits (id, name, gp_name, locality, country_code) OVERRIDING SYSTEM VALUE VALUES
    (1,  'Albert Park',                   'Gran Premio de Australia',          'Melbourne',          'au'),
    (2,  'Shanghai International Circuit','Gran Premio de China',              'Shanghái',           'cn'),
    (3,  'Suzuka',                        'Gran Premio de Japón',              'Suzuka',             'jp'),
    (4,  'Bahrain International Circuit', 'Gran Premio de Baréin',             'Sakhir',             'bh'),
    (5,  'Jeddah Corniche Circuit',       'Gran Premio de Arabia Saudita',     'Yeda',               'sa'),
    (6,  'Miami International Autodrome', 'Gran Premio de Miami',              'Miami',              'us'),
    (7,  'Circuit Gilles Villeneuve',     'Gran Premio de Canadá',             'Montreal',           'ca'),
    (8,  'Circuit de Monaco',             'Gran Premio de Mónaco',             'Montecarlo',         'mc'),
    (9,  'Circuit de Barcelona-Catalunya','Gran Premio de Barcelona-Cataluña', 'Montmeló',           'es'),
    (10, 'Red Bull Ring',                 'Gran Premio de Austria',            'Spielberg',          'at'),
    (11, 'Silverstone',                   'Gran Premio de Gran Bretaña',       'Silverstone',        'gb'),
    (12, 'Spa-Francorchamps',             'Gran Premio de Bélgica',            'Stavelot',           'be'),
    (13, 'Hungaroring',                   'Gran Premio de Hungría',            'Mogyoród',           'hu'),
    (14, 'Zandvoort',                     'Gran Premio de los Países Bajos',   'Zandvoort',          'nl'),
    (15, 'Monza',                         'Gran Premio de Italia',             'Monza',              'it'),
    (16, 'Madring',                       'Gran Premio de España',             'Madrid',             'es'),
    (17, 'Baku City Circuit',             'Gran Premio de Azerbaiyán',         'Bakú',               'az'),
    (18, 'Marina Bay',                    'Gran Premio de Singapur',           'Singapur',           'sg'),
    (19, 'Circuit of the Americas',       'Gran Premio de Estados Unidos',     'Austin',             'us'),
    (20, 'Autódromo Hermanos Rodríguez',  'Gran Premio de la Ciudad de México','Ciudad de México',   'mx'),
    (21, 'Interlagos',                    'Gran Premio de São Paulo',          'São Paulo',          'br'),
    (22, 'Las Vegas Strip',               'Gran Premio de Las Vegas',          'Las Vegas',          'us'),
    (23, 'Lusail',                        'Gran Premio de Qatar',              'Lusail',             'qa'),
    (24, 'Yas Marina',                    'Gran Premio de Abu Dabi',           'Abu Dabi',           'ae')
ON CONFLICT (name) DO NOTHING;

-- COALESCE keeps the empty-table case at 0 so the next insert starts at 1, and
-- taking the live max means a team created through the app is never overwritten.
SELECT setval(pg_get_serial_sequence('categories', 'id'), COALESCE((SELECT max(id) FROM categories), 0));
SELECT setval(pg_get_serial_sequence('seasons',    'id'), COALESCE((SELECT max(id) FROM seasons),    0));
SELECT setval(pg_get_serial_sequence('teams',      'id'), COALESCE((SELECT max(id) FROM teams),      0));
SELECT setval(pg_get_serial_sequence('circuits',   'id'), COALESCE((SELECT max(id) FROM circuits),   0));
