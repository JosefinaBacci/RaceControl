-- 000015 down: devuelve starts_at/ends_at a races y tyre_tests, copiando el
-- timing real del evento (el DEFAULT now() anterior dejaba las dos iguales y el
-- CHECK ends_at > starts_at no validaba).

ALTER TABLE races ADD COLUMN starts_at timestamptz;
ALTER TABLE races ADD COLUMN ends_at timestamptz;
ALTER TABLE tyre_tests ADD COLUMN starts_at timestamptz;
ALTER TABLE tyre_tests ADD COLUMN ends_at timestamptz;

UPDATE races r SET starts_at = e.starts_at, ends_at = e.ends_at
FROM events e WHERE e.id = r.event_id;
UPDATE tyre_tests t SET starts_at = e.starts_at, ends_at = e.ends_at
FROM events e WHERE e.id = t.event_id;

ALTER TABLE races
    ALTER COLUMN starts_at SET NOT NULL,
    ALTER COLUMN ends_at SET NOT NULL,
    ADD CONSTRAINT races_ends_after_start CHECK (ends_at >= starts_at);
ALTER TABLE tyre_tests
    ALTER COLUMN starts_at SET NOT NULL,
    ALTER COLUMN ends_at SET NOT NULL,
    ADD CONSTRAINT tyre_tests_ends_after_start CHECK (ends_at >= starts_at);
