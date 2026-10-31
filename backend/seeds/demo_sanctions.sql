-- Sanctions of the demo Formula 1 races, so the public list, the team inbox and
-- the FIA follow-up show every state: a standing sanction, an amended one, one on
-- a team and a revoked one, some confirmed and some pending. They name fia.admin
-- and the demo team accounts, so the seeder applies this file after the demo
-- accounts and it inserts nothing when those accounts are absent.
--
-- Idempotent like the other seeds, with one twist: the versions and the
-- acknowledgements have BEFORE INSERT guards (000035, 000036) that run before
-- ON CONFLICT is checked, so a re-run skips rows with NOT EXISTS instead.
-- The recipients of each version are written by a trigger: the teams of
-- Formula 1 active when the seed runs.

-- 1. sanctions: the subject and the day of issue
INSERT INTO sanctions (id, driver_id, team_id, issued_at)
OVERRIDING SYSTEM VALUE
SELECT sanction.id, sanction.driver_id, sanction.team_id, sanction.issued_at
FROM (VALUES
    (1, 4,         NULL::bigint, TIMESTAMPTZ '2026-10-04 17:30:00+00'),
    (2, 2,         NULL,         TIMESTAMPTZ '2026-10-04 12:00:00+00'),
    (3, NULL,      3,            TIMESTAMPTZ '2026-10-04 18:00:00+00'),
    (4, 7,         NULL,         TIMESTAMPTZ '2026-10-18 19:00:00+00')
) AS sanction (id, driver_id, team_id, issued_at)
WHERE EXISTS (SELECT 1 FROM users WHERE username = 'fia.admin')
ON CONFLICT (id) DO NOTHING;

-- The sanctions this seed owns: its fixed ids with the subject and the day it
-- gave them. A sanction that already took one of those ids belongs to someone
-- else, and the record is append-only, so its versions are left alone.
CREATE OR REPLACE TEMPORARY VIEW demo_sanctions AS
SELECT sanctions.id
FROM sanctions
JOIN (VALUES
    (1, 4,    NULL::bigint, TIMESTAMPTZ '2026-10-04 17:30:00+00'),
    (2, 2,    NULL,         TIMESTAMPTZ '2026-10-04 12:00:00+00'),
    (3, NULL, 3,            TIMESTAMPTZ '2026-10-04 18:00:00+00'),
    (4, 7,    NULL,         TIMESTAMPTZ '2026-10-18 19:00:00+00')
) AS demo (id, driver_id, team_id, issued_at)
  ON demo.id = sanctions.id
 AND demo.driver_id IS NOT DISTINCT FROM sanctions.driver_id
 AND demo.team_id IS NOT DISTINCT FROM sanctions.team_id
 AND demo.issued_at = sanctions.issued_at;

-- 2. sanction_versions, oldest first so each one follows the one it replaces.
-- Penalty types by id from reference.sql: 2 reprimand, 3 time penalty, 6 grid
-- penalty, 7 licence points, 8 fine.
INSERT INTO sanction_versions (id, sanction_id, action, penalty_type_id, penalty_value, reason, event_id, note, actor_user_id, changed_at)
OVERRIDING SYSTEM VALUE
SELECT version.id, version.sanction_id, version.action, version.penalty_type_id, version.penalty_value,
       version.reason, version.event_id, version.note, users.id, version.changed_at
FROM (VALUES
    (1, 2, 'issued',  2, NULL::integer, 'Obstaculizar a otro piloto en la clasificación (impeding) en la curva 14.',
        1::bigint, NULL::text, TIMESTAMPTZ '2026-10-04 12:00:00+00'),
    (2, 1, 'issued',  3, 5, 'Exceder los límites de pista de forma reiterada en la vuelta 34.',
        1, NULL, TIMESTAMPTZ '2026-10-04 17:30:00+00'),
    (3, 3, 'issued',  8, 10000, 'Liberación insegura de un auto desde su box durante la carrera.',
        1, NULL, TIMESTAMPTZ '2026-10-04 18:00:00+00'),
    (4, 2, 'amended', 6, 3, 'Obstaculizar a otro piloto en la clasificación (impeding) en la curva 14.',
        1, 'Las imágenes a bordo mostraron que el piloto afectado tuvo que abortar su vuelta.', TIMESTAMPTZ '2026-10-05 10:00:00+00'),
    (5, 4, 'issued',  7, 2, 'Provocar una colisión en la primera vuelta.',
        3, NULL, TIMESTAMPTZ '2026-10-18 19:00:00+00'),
    (6, 4, 'revoked', 7, 2, 'Provocar una colisión en la primera vuelta.',
        3, 'Apelación aceptada: el incidente se consideró un lance de carrera.', TIMESTAMPTZ '2026-10-20 15:00:00+00')
) AS version (id, sanction_id, action, penalty_type_id, penalty_value, reason, event_id, note, changed_at)
JOIN demo_sanctions ON demo_sanctions.id = version.sanction_id
JOIN users ON users.username = 'fia.admin'
WHERE NOT EXISTS (SELECT 1 FROM sanction_versions existing WHERE existing.id = version.id)
ORDER BY version.changed_at;

-- 3. sanction_acknowledgements: the team comes from the user (000036). Red Bull
-- confirmed Verstappen's penalty and Mercedes its own fine; Ferrari still owes
-- the amended penalty of Hamilton.
INSERT INTO sanction_acknowledgements (id, sanction_version_id, acknowledged_by, acknowledged_at)
OVERRIDING SYSTEM VALUE
SELECT ack.id, ack.version_id, users.id, ack.acknowledged_at
FROM (VALUES
    (1, 2::bigint, 'redbull.admin',  TIMESTAMPTZ '2026-10-04 20:00:00+00'),
    (2, 3,         'mercedes.admin', TIMESTAMPTZ '2026-10-05 09:00:00+00'),
    (3, 3,         'redbull.admin',  TIMESTAMPTZ '2026-10-05 09:30:00+00')
) AS ack (id, version_id, username, acknowledged_at)
JOIN users ON users.username = ack.username
JOIN sanction_versions version ON version.id = ack.version_id
JOIN demo_sanctions ON demo_sanctions.id = version.sanction_id
WHERE NOT EXISTS (SELECT 1 FROM sanction_acknowledgements existing WHERE existing.id = ack.id)
  AND NOT EXISTS (
      SELECT 1 FROM sanction_acknowledgements existing
      WHERE existing.sanction_version_id = ack.version_id AND existing.acknowledged_by = users.id
  );

-- 4. setval for every sequence, skipped when the accounts were absent and the
-- table stayed empty: a sequence cannot be set below 1
SELECT setval(pg_get_serial_sequence('sanctions',                 'id'), max(id)) FROM sanctions                 HAVING max(id) IS NOT NULL;
SELECT setval(pg_get_serial_sequence('sanction_versions',         'id'), max(id)) FROM sanction_versions         HAVING max(id) IS NOT NULL;
SELECT setval(pg_get_serial_sequence('sanction_acknowledgements', 'id'), max(id)) FROM sanction_acknowledgements HAVING max(id) IS NOT NULL;

DROP VIEW demo_sanctions;
