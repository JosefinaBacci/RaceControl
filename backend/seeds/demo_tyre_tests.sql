-- Results of the demo tyre test (event 4, Circuit of the Americas, 2026-10-20), so
-- the public history and the FIA record show both an original result and a
-- correction. They name fia.admin, which the seeder creates after demo.sql, so this
-- runs after the accounts and inserts nothing when that account is absent.
--
-- Results are append-only, so every insert skips a row whose id or compound is
-- already taken instead of failing, and the correction is only grafted onto the
-- seed's own original.

INSERT INTO tyre_test_results (id, event_id, compound_id, conducted_on, passed, notes, recorded_by, recorded_at)
OVERRIDING SYSTEM VALUE
SELECT seeded.id, 4, tyre_compounds.id, DATE '2026-10-20', seeded.passed, seeded.notes, users.id, TIMESTAMPTZ '2026-10-20 19:00:00+00'
FROM (VALUES
    (1, 'c3',  true, 'Degradación dentro de lo previsto tras 30 vueltas.'),
    (2, 'c4',  true, ''),
    (4, 'wet', true, 'Buen drenaje con pista mojada artificialmente.')
) AS seeded (id, compound_code, passed, notes)
JOIN tyre_compounds ON tyre_compounds.code = seeded.compound_code
JOIN users ON users.username = 'fia.admin'
JOIN tyre_tests ON tyre_tests.event_id = 4
ON CONFLICT DO NOTHING;

INSERT INTO tyre_test_results (id, event_id, compound_id, conducted_on, passed, notes, supersedes_id, correction_note, recorded_by, recorded_at)
OVERRIDING SYSTEM VALUE
SELECT 3, original.event_id, original.compound_id, DATE '2026-10-20', false,
       'Ampollas en el tren trasero a partir de la vuelta 18.',
       original.id, 'Se había cargado como aprobado por error.', original.recorded_by, TIMESTAMPTZ '2026-10-21 10:00:00+00'
FROM tyre_test_results original
JOIN users ON users.id = original.recorded_by AND users.username = 'fia.admin'
JOIN tyre_compounds ON tyre_compounds.id = original.compound_id AND tyre_compounds.code = 'c4'
WHERE original.id = 2 AND original.event_id = 4
ON CONFLICT DO NOTHING;

-- setval skipped when the account was absent and the table stayed empty
SELECT setval(pg_get_serial_sequence('tyre_test_results', 'id'), max(id)) FROM tyre_test_results HAVING max(id) IS NOT NULL;
