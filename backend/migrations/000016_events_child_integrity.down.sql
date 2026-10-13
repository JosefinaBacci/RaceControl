-- 000016 down: deshace la relación 1:1 y el tipo del evento, devuelve el nombre a
-- las hijas copiándolo del evento y vuelve opcional la ronda.

-- 6. Se van los triggers de "exactamente una hija" y sus funciones.
DROP TRIGGER events_require_child ON events;
DROP TRIGGER races_keep_event_child ON races;
DROP TRIGGER tyre_tests_keep_event_child ON tyre_tests;
DROP TRIGGER technical_controls_keep_event_child ON technical_controls;
DROP FUNCTION assert_event_has_child();
DROP FUNCTION assert_event_keeps_child();

-- 5. Las hijas vuelven a la FK simple a events y pierden el tipo declarado.
ALTER TABLE races
    DROP CONSTRAINT races_event_kind_fkey,
    DROP CONSTRAINT races_event_kind_check,
    DROP COLUMN event_kind,
    ADD CONSTRAINT races_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE RESTRICT;

ALTER TABLE tyre_tests
    DROP CONSTRAINT tyre_tests_event_kind_fkey,
    DROP CONSTRAINT tyre_tests_event_kind_check,
    DROP COLUMN event_kind,
    ADD CONSTRAINT tyre_tests_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE RESTRICT;

ALTER TABLE technical_controls
    DROP CONSTRAINT technical_controls_event_kind_fkey,
    DROP CONSTRAINT technical_controls_event_kind_check,
    DROP COLUMN event_kind,
    ADD CONSTRAINT technical_controls_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE RESTRICT;

-- 4. events pierde el tipo.
ALTER TABLE events
    DROP CONSTRAINT events_kind_valid,
    DROP CONSTRAINT events_id_kind_unique,
    DROP COLUMN kind;

-- 3. Unicidad por evento -> índice plano
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

-- 1. El nombre vuelve a las hijas, copiado del evento
ALTER TABLE races ADD COLUMN name text;
UPDATE races r SET name = e.name FROM events e WHERE e.id = r.event_id;
ALTER TABLE races ALTER COLUMN name SET NOT NULL;

ALTER TABLE tyre_tests ADD COLUMN name text;
UPDATE tyre_tests t SET name = e.name FROM events e WHERE e.id = t.event_id;
ALTER TABLE tyre_tests ALTER COLUMN name SET NOT NULL;

ALTER TABLE technical_controls ADD COLUMN name text;
UPDATE technical_controls c SET name = e.name FROM events e WHERE e.id = c.event_id;
ALTER TABLE technical_controls ALTER COLUMN name SET NOT NULL;
