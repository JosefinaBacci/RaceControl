-- 000016 down: undoes the 1:1 relationship and the event kind, returns the name to
-- the children copying it from the event, makes the round optional again, drops
-- the date guard and narrows the team derivation back to the driver and the event.

-- 8. The date guard goes away.
DROP TRIGGER events_keep_final_date ON events;
DROP FUNCTION assert_event_date_is_final();

-- 7. The derived team goes back to firing only when the driver or the event changes.
DROP TRIGGER score_entries_set_team ON score_entries;
CREATE TRIGGER score_entries_set_team
    BEFORE INSERT OR UPDATE OF event_id, driver_id ON score_entries
    FOR EACH ROW EXECUTE FUNCTION set_score_entry_team();

-- 6. The "exactly one child" triggers and their functions go away.
DROP TRIGGER events_require_child ON events;
DROP TRIGGER races_keep_event_child ON races;
DROP TRIGGER tyre_tests_keep_event_child ON tyre_tests;
DROP TRIGGER technical_controls_keep_event_child ON technical_controls;
DROP FUNCTION assert_event_has_child();
DROP FUNCTION assert_event_keeps_child();

-- 5. The children go back to the plain FK to events and lose the declared kind.
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

-- 4. events loses the kind.
ALTER TABLE events
    DROP CONSTRAINT events_kind_valid,
    DROP CONSTRAINT events_id_kind_unique,
    DROP COLUMN kind;

-- 3. Uniqueness per event -> plain index
ALTER TABLE races DROP CONSTRAINT races_event_id_unique;
CREATE INDEX races_event_id_idx ON races (event_id);

ALTER TABLE tyre_tests DROP CONSTRAINT tyre_tests_event_id_unique;
CREATE INDEX tyre_tests_event_id_idx ON tyre_tests (event_id);

ALTER TABLE technical_controls DROP CONSTRAINT technical_controls_event_id_unique;
CREATE INDEX technical_controls_event_id_idx ON technical_controls (event_id);

-- 2. The round becomes optional again
ALTER TABLE races DROP CONSTRAINT races_round_positive;
ALTER TABLE races
    ALTER COLUMN round DROP NOT NULL,
    ADD CONSTRAINT races_round_positive CHECK (round IS NULL OR round > 0);

-- 1. The name returns to the children, copied from the event
ALTER TABLE races ADD COLUMN name text;
UPDATE races r SET name = e.name FROM events e WHERE e.id = r.event_id;
ALTER TABLE races ALTER COLUMN name SET NOT NULL;

ALTER TABLE tyre_tests ADD COLUMN name text;
UPDATE tyre_tests t SET name = e.name FROM events e WHERE e.id = t.event_id;
ALTER TABLE tyre_tests ALTER COLUMN name SET NOT NULL;

ALTER TABLE technical_controls ADD COLUMN name text;
UPDATE technical_controls c SET name = e.name FROM events e WHERE e.id = c.event_id;
ALTER TABLE technical_controls ALTER COLUMN name SET NOT NULL;
