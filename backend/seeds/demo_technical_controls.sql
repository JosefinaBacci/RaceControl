-- Results of the demo technical control (event 2, Formula 1, 5 October), so the
-- public history and the FIA sheet show every state: passed, failed, corrected
-- and pending (McLaren). It names fia.admin, which the seeder creates after
-- demo.sql, so it runs after the demo accounts and inserts nothing without them.
-- Idempotent: fixed ids, and a row that clashes with one already there (its id
-- or a result the FIA recorded for the same team) is skipped, never rewritten.

INSERT INTO technical_control_results
    (id, event_id, category_id, team_id, conducted_on, passed, notes, supersedes_id, correction_note, recorded_by, recorded_at)
OVERRIDING SYSTEM VALUE
SELECT seeded.id, 2, 1, seeded.team_id, DATE '2026-10-05', seeded.passed, seeded.notes,
       seeded.supersedes_id, seeded.correction_note, users.id, seeded.recorded_at
FROM (VALUES
    (1, 1, true,  '', NULL::bigint, NULL::text, TIMESTAMPTZ '2026-10-05 13:00:00+00'),
    (2, 2, false, 'Alerón delantero con flexión por encima del límite reglamentario', NULL, NULL,
        TIMESTAMPTZ '2026-10-05 13:05:00+00'),
    (3, 3, false, 'Peso mínimo no alcanzado', NULL, NULL, TIMESTAMPTZ '2026-10-05 13:10:00+00'),
    (4, 3, true,  'Pesaje repetido con la balanza calibrada: dentro del mínimo', 3,
        'La primera balanza estaba descalibrada', TIMESTAMPTZ '2026-10-05 15:00:00+00')
) AS seeded (id, team_id, passed, notes, supersedes_id, correction_note, recorded_at)
JOIN users ON users.username = 'fia.admin'
ON CONFLICT DO NOTHING;

-- Skipped when the accounts were absent and nothing was inserted.
SELECT setval(pg_get_serial_sequence('technical_control_results', 'id'), max(id))
FROM technical_control_results HAVING max(id) IS NOT NULL;
