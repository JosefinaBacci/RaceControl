-- 000016_events_child_integrity: an event declares its kind and has exactly one
-- child of that kind, which shares its primary key. A race always has a round,
-- unique among the active races of its championship, and runs on a circuit of
-- the catalogue. No event name is stored: the label is composed by the backend.
-- A result always derives its credited team and a race that already has results
-- keeps its date and stays on the calendar.

-- 1. The name belongs to no row: the backend composes the label from the kind
--    and the child's attributes, so a stored name could only diverge from them.
ALTER TABLE races DROP COLUMN name;
ALTER TABLE tyre_tests DROP COLUMN name;
ALTER TABLE technical_controls DROP COLUMN name;

-- 2. A race always has a round. With NOT NULL, the per-championship UNIQUE can no
--    longer be dodged by loading the round as NULL.
ALTER TABLE races DROP CONSTRAINT races_round_positive;
ALTER TABLE races
    ALTER COLUMN round SET NOT NULL,
    ADD CONSTRAINT races_round_positive CHECK (round > 0);

-- 3. A child is its event: the event id is its primary key, so a child has a
--    single id and at most one child hangs from each event. The own identity
--    column was referenced by nothing.
DROP INDEX races_event_id_idx;
ALTER TABLE races DROP COLUMN id, ADD PRIMARY KEY (event_id);

DROP INDEX tyre_tests_event_id_idx;
ALTER TABLE tyre_tests DROP COLUMN id, ADD PRIMARY KEY (event_id);

DROP INDEX technical_controls_event_id_idx;
ALTER TABLE technical_controls DROP COLUMN id, ADD PRIMARY KEY (event_id);

-- 4. The kind of the event. Without it nothing stops an event from hanging children
--    of two different kinds. It is derived from the child that already existed,
--    which was the only source of the kind before this migration.
ALTER TABLE events ADD COLUMN kind text;
UPDATE events e SET kind = CASE
        WHEN EXISTS (SELECT 1 FROM races r WHERE r.event_id = e.id)              THEN 'race'
        WHEN EXISTS (SELECT 1 FROM tyre_tests t WHERE t.event_id = e.id)         THEN 'tyre_test'
        WHEN EXISTS (SELECT 1 FROM technical_controls c WHERE c.event_id = e.id) THEN 'technical_control'
    END;
ALTER TABLE events
    ALTER COLUMN kind SET NOT NULL,
    ADD CONSTRAINT events_kind_valid
        CHECK (kind IN ('race', 'tyre_test', 'technical_control')),
    ADD CONSTRAINT events_id_kind_unique UNIQUE (id, kind);

-- 5. Each child declares the event kind and pins it as its own, so its identity FK
--    cannot point at an event of another kind: a race event hangs a race and a
--    control event hangs a technical control. The column is constant, so the
--    DEFAULT is enough and it need not be written on every INSERT. The kind of an
--    event never changes while it has a child, hence ON UPDATE RESTRICT.
ALTER TABLE tyre_tests
    ADD COLUMN event_kind text NOT NULL DEFAULT 'tyre_test',
    ADD CONSTRAINT tyre_tests_event_kind_check CHECK (event_kind = 'tyre_test'),
    DROP CONSTRAINT tyre_tests_event_id_fkey,
    ADD CONSTRAINT tyre_tests_event_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE RESTRICT;

ALTER TABLE technical_controls
    ADD COLUMN event_kind text NOT NULL DEFAULT 'technical_control',
    ADD CONSTRAINT technical_controls_event_kind_check CHECK (event_kind = 'technical_control'),
    DROP CONSTRAINT technical_controls_event_id_fkey,
    ADD CONSTRAINT technical_controls_event_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE RESTRICT,
    ADD CONSTRAINT technical_controls_control_type_not_blank CHECK (btrim(control_type) <> '');

-- 6. A race also copies the championship and whether its event is active: the
--    round is unique among the active races of a championship, and a soft-deleted
--    race must release its round so it can be scheduled again. One FK holds every
--    copy honest; CASCADE carries a change of season, category or status, while a
--    change of kind still fails on the CHECK.
ALTER TABLE races
    DROP CONSTRAINT races_round_unique_per_season_category,
    DROP CONSTRAINT races_event_season_category_fkey,
    DROP CONSTRAINT races_event_id_fkey;

ALTER TABLE events
    DROP CONSTRAINT events_id_season_category_unique,
    ADD CONSTRAINT events_race_identity_unique UNIQUE (id, kind, season_id, category_id, is_active);

ALTER TABLE races
    ADD COLUMN event_kind text NOT NULL DEFAULT 'race',
    ADD COLUMN event_is_active boolean NOT NULL DEFAULT true,
    ADD CONSTRAINT races_event_kind_check CHECK (event_kind = 'race'),
    ADD CONSTRAINT races_event_fkey
        FOREIGN KEY (event_id, event_kind, season_id, category_id, event_is_active)
        REFERENCES events (id, kind, season_id, category_id, is_active)
        ON DELETE RESTRICT ON UPDATE CASCADE;

CREATE UNIQUE INDEX races_round_unique_per_championship
    ON races (season_id, category_id, round) WHERE event_is_active;

-- 7. Circuits are a catalogue: a race picks one instead of retyping its name,
--    place and country, and the Grand Prix name the label shows comes with it.
CREATE TABLE circuits (
    id             bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name           text        NOT NULL UNIQUE,
    gp_name        text        NOT NULL,
    locality       text        NOT NULL,
    country_code   text        NOT NULL,
    is_active      boolean     NOT NULL GENERATED ALWAYS AS (deactivated_at IS NULL) STORED,
    deactivated_at timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT circuits_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT circuits_gp_name_not_blank CHECK (btrim(gp_name) <> ''),
    CONSTRAINT circuits_locality_not_blank CHECK (btrim(locality) <> ''),
    CONSTRAINT circuits_country_code_format CHECK (country_code ~ '^[a-z]{2}$')
);

CREATE TRIGGER circuits_set_updated_at BEFORE UPDATE ON circuits
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO circuits (name, gp_name, locality, country_code)
SELECT DISTINCT ON (r.circuit) r.circuit, e.name, r.location, r.country_code
FROM races r
JOIN events e ON e.id = r.event_id
ORDER BY r.circuit, e.id;

ALTER TABLE races ADD COLUMN circuit_id bigint REFERENCES circuits (id) ON DELETE RESTRICT;
UPDATE races r SET circuit_id = c.id FROM circuits c WHERE c.name = r.circuit;
ALTER TABLE races
    ALTER COLUMN circuit_id SET NOT NULL,
    DROP COLUMN circuit,
    DROP COLUMN location,
    DROP COLUMN country_code;

CREATE INDEX races_circuit_id_idx ON races (circuit_id);

ALTER TABLE events DROP COLUMN name;

-- 8. Exactly one child. An event without a child is of no kind: its label cannot
--    be composed. Both halves of the invariant are checked only at commit
--    (DEFERRABLE) so that inserting the event and its child in a single
--    transaction stays possible. "At most one" needs no trigger: the child's
--    primary key and its kind FK already guarantee it.
CREATE FUNCTION assert_event_has_child() RETURNS trigger AS $$
DECLARE
    has_child boolean;
BEGIN
    has_child := CASE NEW.kind
        WHEN 'race'              THEN EXISTS (SELECT 1 FROM races              WHERE event_id = NEW.id)
        WHEN 'tyre_test'         THEN EXISTS (SELECT 1 FROM tyre_tests         WHERE event_id = NEW.id)
        WHEN 'technical_control' THEN EXISTS (SELECT 1 FROM technical_controls WHERE event_id = NEW.id)
    END;

    IF NOT has_child THEN
        RAISE EXCEPTION 'event % of kind % has no child row', NEW.id, NEW.kind;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER events_require_child
    AFTER INSERT OR UPDATE OF kind ON events
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION assert_event_has_child();

-- Deleting only the child would leave the event orphaned and the events trigger
-- would not notice (nobody touched the event). The "parent is deleted too" case
-- is ignored.
CREATE FUNCTION assert_event_keeps_child() RETURNS trigger AS $$
DECLARE
    has_child boolean;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM events WHERE id = OLD.event_id) THEN
        RETURN NULL;
    END IF;

    EXECUTE format('SELECT EXISTS (SELECT 1 FROM %I WHERE event_id = $1)', TG_TABLE_NAME)
        INTO has_child USING OLD.event_id;

    IF NOT has_child THEN
        RAISE EXCEPTION 'event % would be left without a child in %', OLD.event_id, TG_TABLE_NAME;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER races_keep_event_child
    AFTER DELETE OR UPDATE OF event_id ON races
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION assert_event_keeps_child();

CREATE CONSTRAINT TRIGGER tyre_tests_keep_event_child
    AFTER DELETE OR UPDATE OF event_id ON tyre_tests
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION assert_event_keeps_child();

CREATE CONSTRAINT TRIGGER technical_controls_keep_event_child
    AFTER DELETE OR UPDATE OF event_id ON technical_controls
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION assert_event_keeps_child();

-- 9. The credited team is derived on every write, not only when the driver or the
--    event changes, so a direct write can no longer overwrite it.
DROP TRIGGER score_entries_set_team ON score_entries;
CREATE TRIGGER score_entries_set_team
    BEFORE INSERT OR UPDATE ON score_entries
    FOR EACH ROW EXECUTE FUNCTION set_score_entry_team();

-- 10. A race that already has results has been run: its date is final and it can
--     no longer be removed from the calendar, or its results would hang from an
--     event nobody sees. The guard fires only when the date or the status really
--     changes, so editing anything else on the event still works.
CREATE FUNCTION assert_raced_event_is_final() RETURNS trigger AS $$
BEGIN
    IF (NEW.starts_at IS DISTINCT FROM OLD.starts_at
            OR NEW.deactivated_at IS DISTINCT FROM OLD.deactivated_at)
        AND EXISTS (SELECT 1 FROM score_entries WHERE event_id = NEW.id) THEN
        RAISE EXCEPTION 'event % has results: its date and status are final', NEW.id
            USING ERRCODE = 'RC001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER events_keep_raced_event_final
    BEFORE UPDATE OF starts_at, deactivated_at ON events
    FOR EACH ROW EXECUTE FUNCTION assert_raced_event_is_final();
