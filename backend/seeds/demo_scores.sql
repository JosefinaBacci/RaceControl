-- Score history of the demo race (event 1), so the score notices show both of
-- their states. It names the demo accounts, which the seeder creates after
-- demo.sql, so it runs last and inserts nothing when those accounts are absent.
-- Idempotent like the other seeds: the ids are fixed and every insert skips a
-- row that already exists.
--
-- fia.admin publishes the race with Hamilton on 10 points and corrects him to the
-- 12 the score entries hold. Red Bull confirmed the publication; Ferrari, touched
-- by both changes, and Mercedes still have to confirm.

-- 1. score_change_log: the publication and its correction
INSERT INTO score_change_log (id, event_id, actor_user_id, changed_at)
OVERRIDING SYSTEM VALUE
SELECT change.id, 1, users.id, change.changed_at
FROM (VALUES
    (1, TIMESTAMPTZ '2026-10-04 18:00:00+00'),
    (2, TIMESTAMPTZ '2026-10-06 10:00:00+00')
) AS change (id, changed_at)
JOIN users ON users.username = 'fia.admin'
ON CONFLICT (id) DO NOTHING;

-- 2. score_change_log_items: the team is derived by a trigger (000008)
INSERT INTO score_change_log_items (change_log_id, driver_id, old_points, new_points)
SELECT item.change_log_id, item.driver_id, item.old_points, item.new_points
FROM (VALUES
    (1, 1, NULL::numeric, 25::numeric),
    (1, 4, NULL, 18),
    (1, 6, NULL, 15),
    (1, 2, NULL, 10),
    (2, 2, 10, 12)
) AS item (change_log_id, driver_id, old_points, new_points)
WHERE EXISTS (SELECT 1 FROM score_change_log WHERE id = item.change_log_id)
ON CONFLICT (change_log_id, driver_id) DO NOTHING;

-- 3. score_acknowledgements: the team comes from the user (000023)
INSERT INTO score_acknowledgements (id, change_log_id, acknowledged_by, acknowledged_at)
OVERRIDING SYSTEM VALUE
SELECT 1, 1, users.id, TIMESTAMPTZ '2026-10-05 09:30:00+00'
FROM users
WHERE users.username = 'redbull.admin'
  AND EXISTS (SELECT 1 FROM score_change_log WHERE id = 1)
ON CONFLICT DO NOTHING;

-- 4. setval for every sequence, skipped when the accounts were absent and the
-- table stayed empty: a sequence cannot be set below 1
SELECT setval(pg_get_serial_sequence('score_change_log',       'id'), max(id)) FROM score_change_log       HAVING max(id) IS NOT NULL;
SELECT setval(pg_get_serial_sequence('score_change_log_items', 'id'), max(id)) FROM score_change_log_items HAVING max(id) IS NOT NULL;
SELECT setval(pg_get_serial_sequence('score_acknowledgements', 'id'), max(id)) FROM score_acknowledgements HAVING max(id) IS NOT NULL;
