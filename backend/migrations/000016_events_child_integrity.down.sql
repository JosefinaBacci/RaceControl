-- 000016 down: returns the championship and the stored name to the event, the
-- free-text circuit and the round to the race, gives each child its own identity
-- again, drops the event kind, the round and circuit catalogues and the guards,
-- and narrows the team derivation back to the driver and the event.

-- 9. The guard on raced events goes away.
DROP TRIGGER events_keep_raced_event_final ON events;
DROP FUNCTION assert_raced_event_is_final();

-- 8. The derived team goes back to firing only when the driver or the event changes.
DROP TRIGGER score_entries_set_team ON score_entries;
CREATE TRIGGER score_entries_set_team
    BEFORE INSERT OR UPDATE OF event_id, driver_id ON score_entries
    FOR EACH ROW EXECUTE FUNCTION set_score_entry_team();

-- 7. The "exactly one child" triggers and their functions go away.
DROP TRIGGER events_require_child ON events;
DROP TRIGGER races_keep_event_child ON races;
DROP TRIGGER tyre_tests_keep_event_child ON tyre_tests;
DROP TRIGGER technical_controls_keep_event_child ON technical_controls;
DROP FUNCTION assert_event_has_child();
DROP FUNCTION assert_event_keeps_child();

-- 6. The children go back to the plain FK to events and lose the declared kind.
ALTER TABLE races
    DROP CONSTRAINT races_event_fkey,
    DROP CONSTRAINT races_event_kind_check,
    DROP COLUMN event_kind,
    ADD CONSTRAINT races_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE RESTRICT;

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

-- 5. The event gets its championship and name back and the race its round and
--    free-text circuit, all rebuilt from the children, the rounds and the circuits.
DROP INDEX events_starts_at_idx;
ALTER TABLE events
    ADD COLUMN season_id bigint REFERENCES seasons (id) ON DELETE RESTRICT,
    ADD COLUMN category_id bigint REFERENCES categories (id) ON DELETE RESTRICT,
    ADD COLUMN name text;

UPDATE events e SET season_id = ro.season_id, category_id = ro.category_id, name = ro.gp_name
FROM races r JOIN rounds ro ON ro.id = r.round_id
WHERE r.event_id = e.id;
UPDATE events e SET season_id = t.season_id, category_id = t.category_id, name = 'Prueba de neumáticos'
FROM tyre_tests t WHERE t.event_id = e.id;
UPDATE events e SET season_id = c.season_id, category_id = c.category_id, name = c.control_type
FROM technical_controls c WHERE c.event_id = e.id;

ALTER TABLE events
    ALTER COLUMN season_id SET NOT NULL,
    ALTER COLUMN category_id SET NOT NULL,
    ALTER COLUMN name SET NOT NULL,
    ADD CONSTRAINT events_name_not_blank CHECK (btrim(name) <> ''),
    ADD CONSTRAINT events_id_season_category_unique UNIQUE (id, season_id, category_id);
CREATE INDEX events_season_starts_at_idx ON events (season_id, starts_at);

ALTER TABLE races
    ADD COLUMN season_id bigint,
    ADD COLUMN category_id bigint,
    ADD COLUMN round integer,
    ADD COLUMN circuit text,
    ADD COLUMN location text,
    ADD COLUMN country_code text;

UPDATE races r
SET season_id = ro.season_id, category_id = ro.category_id, round = ro.number,
    circuit = c.name, location = c.locality, country_code = c.country_code
FROM rounds ro JOIN circuits c ON c.id = ro.circuit_id
WHERE ro.id = r.round_id;

ALTER TABLE races
    ALTER COLUMN season_id SET NOT NULL,
    ALTER COLUMN category_id SET NOT NULL,
    ALTER COLUMN circuit SET NOT NULL,
    ALTER COLUMN location SET NOT NULL,
    ALTER COLUMN country_code SET NOT NULL,
    ADD CONSTRAINT races_circuit_not_blank CHECK (btrim(circuit) <> ''),
    ADD CONSTRAINT races_country_code_format CHECK (country_code ~ '^[a-z]{2}$'),
    ADD CONSTRAINT races_round_positive CHECK (round IS NULL OR round > 0),
    ADD CONSTRAINT races_event_season_category_fkey
        FOREIGN KEY (event_id, season_id, category_id)
        REFERENCES events (id, season_id, category_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    ADD CONSTRAINT races_round_unique_per_season_category UNIQUE (season_id, category_id, round),
    DROP COLUMN round_id;

DROP TABLE rounds;
DROP TABLE circuits;

-- 4. The tyre tests and technical controls lose their own championship.
DROP INDEX tyre_tests_season_category_idx;
ALTER TABLE tyre_tests DROP COLUMN season_id, DROP COLUMN category_id;
DROP INDEX technical_controls_season_category_idx;
ALTER TABLE technical_controls DROP COLUMN season_id, DROP COLUMN category_id;

-- 3. events loses the kind.
ALTER TABLE events
    DROP CONSTRAINT events_kind_valid,
    DROP CONSTRAINT events_id_kind_unique,
    DROP COLUMN kind;

-- 2. Each child gets its own identity back; the event id becomes a plain indexed FK.
ALTER TABLE races DROP CONSTRAINT races_pkey;
ALTER TABLE races ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;
CREATE INDEX races_event_id_idx ON races (event_id);

ALTER TABLE tyre_tests DROP CONSTRAINT tyre_tests_pkey;
ALTER TABLE tyre_tests ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;
CREATE INDEX tyre_tests_event_id_idx ON tyre_tests (event_id);

ALTER TABLE technical_controls DROP CONSTRAINT technical_controls_pkey;
ALTER TABLE technical_controls ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;
CREATE INDEX technical_controls_event_id_idx ON technical_controls (event_id);

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
