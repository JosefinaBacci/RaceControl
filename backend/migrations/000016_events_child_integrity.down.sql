-- 000016 down: devuelve el nombre a las hijas copiándolo del evento, vuelve
-- opcional la ronda y cambia la unicidad por evento por el índice plano.

-- 1. Unicidad por evento -> índice plano
ALTER TABLE races DROP CONSTRAINT races_event_id_unique;
CREATE INDEX races_event_id_idx ON races (event_id);

ALTER TABLE tyre_tests DROP CONSTRAINT tyre_tests_event_id_unique;
CREATE INDEX tyre_tests_event_id_idx ON tyre_tests (event_id);

ALTER TABLE technical_controls DROP CONSTRAINT technical_controls_event_id_unique;
CREATE INDEX technical_controls_event_id_idx ON technical_controls (event_id);

-- 2. La ronda vuelve a ser opcional
ALTER TABLE races DROP CONSTRAINT races_round_positive;
ALTER TABLE races
    ALTER COLUMN round DROP NOT NULL,
    ADD CONSTRAINT races_round_positive CHECK (round IS NULL OR round > 0);

-- 3. El nombre vuelve a las hijas, copiado del evento
ALTER TABLE races ADD COLUMN name text;
UPDATE races r SET name = e.name FROM events e WHERE e.id = r.event_id;
ALTER TABLE races ALTER COLUMN name SET NOT NULL;

ALTER TABLE tyre_tests ADD COLUMN name text;
UPDATE tyre_tests t SET name = e.name FROM events e WHERE e.id = t.event_id;
ALTER TABLE tyre_tests ALTER COLUMN name SET NOT NULL;

ALTER TABLE technical_controls ADD COLUMN name text;
UPDATE technical_controls c SET name = e.name FROM events e WHERE e.id = c.event_id;
ALTER TABLE technical_controls ALTER COLUMN name SET NOT NULL;
