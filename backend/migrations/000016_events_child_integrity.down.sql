-- 000016 down: returns the stored names and the free-text circuit, gives each
-- child its own identity again, drops the event kind, the active-status copy and
-- the guards, makes the round optional again and narrows the team derivation back
-- to the driver and the event.

-- 10. The date guard goes away.
DROP TRIGGER events_keep_final_date ON events;
DROP FUNCTION assert_event_date_is_final();

-- 9. The derived team goes back to firing only when the driver or the event changes.
DROP TRIGGER score_entries_set_team ON score_entries;
CREATE TRIGGER score_entries_set_team
    BEFORE INSERT OR UPDATE OF event_id, driver_id ON score_entries
    FOR EACH ROW EXECUTE FUNCTION set_score_entry_team();

-- 8. The "exactly one child" triggers and their functions go away.
DROP TRIGGER events_require_child ON events;
DROP TRIGGER races_keep_event_child ON races;
DROP TRIGGER tyre_tests_keep_event_child ON tyre_tests;
DROP TRIGGER technical_controls_keep_event_child ON technical_controls;
DROP FUNCTION assert_event_has_child();
DROP FUNCTION assert_event_keeps_child();

-- 7. The event name and the free-text circuit come back, rebuilt from the label
--    parts, and the circuit catalogue goes away.
ALTER TABLE events ADD COLUMN name text;
UPDATE events e SET name = c.gp_name
FROM races r JOIN circuits c ON c.id = r.circuit_id
WHERE r.event_id = e.id;
UPDATE events e SET name = tc.control_type
FROM technical_controls tc WHERE tc.event_id = e.id;
UPDATE events SET name = 'Prueba de neumáticos' WHERE kind = 'tyre_test';
ALTER TABLE events
    ALTER COLUMN name SET NOT NULL,
    ADD CONSTRAINT events_name_not_blank CHECK (btrim(name) <> '');

ALTER TABLE races
    ADD COLUMN circuit text,
    ADD COLUMN location text,
    ADD COLUMN country_code text;
UPDATE races r
SET circuit = c.name, location = c.locality, country_code = c.country_code
FROM circuits c WHERE c.id = r.circuit_id;
ALTER TABLE races
    ALTER COLUMN circuit SET NOT NULL,
    ALTER COLUMN location SET NOT NULL,
    ALTER COLUMN country_code SET NOT NULL,
    ADD CONSTRAINT races_circuit_not_blank CHECK (btrim(circuit) <> ''),
    ADD CONSTRAINT races_country_code_format CHECK (country_code ~ '^[a-z]{2}$'),
    DROP COLUMN circuit_id;

DROP TABLE circuits;

-- 6. The race loses its active-status copy and goes back to the plain event FK
--    plus the championship FK, with the round unique over every race.
DROP INDEX races_round_unique_per_championship;
ALTER TABLE races
    DROP CONSTRAINT races_event_fkey,
    DROP CONSTRAINT races_event_kind_check,
    DROP COLUMN event_kind,
    DROP COLUMN event_is_active;

ALTER TABLE events
    DROP CONSTRAINT events_race_identity_unique,
    ADD CONSTRAINT events_id_season_category_unique UNIQUE (id, season_id, category_id);

ALTER TABLE races
    ADD CONSTRAINT races_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE RESTRICT,
    ADD CONSTRAINT races_event_season_category_fkey
        FOREIGN KEY (event_id, season_id, category_id)
        REFERENCES events (id, season_id, category_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    ADD CONSTRAINT races_round_unique_per_season_category UNIQUE (season_id, category_id, round);

-- 5. The other children go back to the plain FK to events and lose the declared kind.
ALTER TABLE tyre_tests
    DROP CONSTRAINT tyre_tests_event_fkey,
    DROP CONSTRAINT tyre_tests_event_kind_check,
    DROP COLUMN event_kind,
    ADD CONSTRAINT tyre_tests_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE RESTRICT;

ALTER TABLE technical_controls
    DROP CONSTRAINT technical_controls_event_fkey,
    DROP CONSTRAINT technical_controls_event_kind_check,
    DROP CONSTRAINT technical_controls_control_type_not_blank,
    DROP COLUMN event_kind,
    ADD CONSTRAINT technical_controls_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE RESTRICT;

-- 4. events loses the kind.
ALTER TABLE events
    DROP CONSTRAINT events_kind_valid,
    DROP CONSTRAINT events_id_kind_unique,
    DROP COLUMN kind;

-- 3. Each child gets its own identity back; the event id becomes a plain indexed FK.
ALTER TABLE races DROP CONSTRAINT races_pkey;
ALTER TABLE races ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;
CREATE INDEX races_event_id_idx ON races (event_id);

ALTER TABLE tyre_tests DROP CONSTRAINT tyre_tests_pkey;
ALTER TABLE tyre_tests ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;
CREATE INDEX tyre_tests_event_id_idx ON tyre_tests (event_id);

ALTER TABLE technical_controls DROP CONSTRAINT technical_controls_pkey;
ALTER TABLE technical_controls ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;
CREATE INDEX technical_controls_event_id_idx ON technical_controls (event_id);

-- 2. The round becomes optional again.
ALTER TABLE races DROP CONSTRAINT races_round_positive;
ALTER TABLE races
    ALTER COLUMN round DROP NOT NULL,
    ADD CONSTRAINT races_round_positive CHECK (round IS NULL OR round > 0);

-- 1. The name returns to the children, copied from the event.
ALTER TABLE races ADD COLUMN name text;
UPDATE races r SET name = e.name FROM events e WHERE e.id = r.event_id;
ALTER TABLE races ALTER COLUMN name SET NOT NULL;

ALTER TABLE tyre_tests ADD COLUMN name text;
UPDATE tyre_tests t SET name = e.name FROM events e WHERE e.id = t.event_id;
ALTER TABLE tyre_tests ALTER COLUMN name SET NOT NULL;

ALTER TABLE technical_controls ADD COLUMN name text;
UPDATE technical_controls c SET name = e.name FROM events e WHERE e.id = c.event_id;
ALTER TABLE technical_controls ALTER COLUMN name SET NOT NULL;
