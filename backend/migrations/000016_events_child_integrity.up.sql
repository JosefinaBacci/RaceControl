-- 000016_events_child_integrity: an event declares its kind and has exactly one
-- child of that kind, which shares its primary key. The event keeps only what
-- every kind shares (kind and timing); each child owns its championship. A race
-- runs a round of the championship, and a round is unique among the active rounds
-- of its season and category. No event name is stored: the label is composed by
-- the backend. Only a race takes results; a result always derives its credited
-- team, and a race that already has results keeps its date, its round and the
-- round's championship and circuit, and stays on the calendar.
--
-- A tyre test now requires a circuit that the old schema never stored, so a
-- database that already holds tyre tests cannot migrate: run make db-reset.

-- 1. The name belongs to no row: the backend composes the label from the kind
--    and the child's attributes, so a stored name could only diverge from them.
ALTER TABLE races DROP COLUMN name;
ALTER TABLE tyre_tests DROP COLUMN name;
ALTER TABLE technical_controls DROP COLUMN name;

-- 2. A child is its event: the event id is its primary key, so a child has a
--    single id and at most one child hangs from each event. The own identity
--    column was referenced by nothing.
DROP INDEX races_event_id_idx;
ALTER TABLE races DROP COLUMN id, ADD PRIMARY KEY (event_id);

DROP INDEX tyre_tests_event_id_idx;
ALTER TABLE tyre_tests DROP COLUMN id, ADD PRIMARY KEY (event_id);

DROP INDEX technical_controls_event_id_idx;
ALTER TABLE technical_controls DROP COLUMN id, ADD PRIMARY KEY (event_id);

-- 3. The kind of the event. Without it nothing stops an event from hanging children
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

-- 4. The championship belongs to the child: a tyre test and a technical control
--    own their season and category, and a race takes them from its round (§5).
--    The race's copies and the FKs that kept them honest go away with them.
--    A tyre test also runs on a circuit of the catalogue, added in §5.
ALTER TABLE tyre_tests
    ADD COLUMN season_id bigint REFERENCES seasons (id) ON DELETE RESTRICT,
    ADD COLUMN category_id bigint REFERENCES categories (id) ON DELETE RESTRICT;
UPDATE tyre_tests t SET season_id = e.season_id, category_id = e.category_id
FROM events e WHERE e.id = t.event_id;
ALTER TABLE tyre_tests
    ALTER COLUMN season_id SET NOT NULL,
    ALTER COLUMN category_id SET NOT NULL;

ALTER TABLE technical_controls
    ADD COLUMN season_id bigint REFERENCES seasons (id) ON DELETE RESTRICT,
    ADD COLUMN category_id bigint REFERENCES categories (id) ON DELETE RESTRICT;
UPDATE technical_controls c SET season_id = e.season_id, category_id = e.category_id
FROM events e WHERE e.id = c.event_id;
ALTER TABLE technical_controls
    ALTER COLUMN season_id SET NOT NULL,
    ALTER COLUMN category_id SET NOT NULL;

CREATE INDEX tyre_tests_season_category_idx ON tyre_tests (season_id, category_id);
CREATE INDEX technical_controls_season_category_idx ON technical_controls (season_id, category_id);

-- 5. Circuits and rounds. A circuit is a catalogue entry a round picks instead of
--    retyping its name, place and country. A round is a date of a championship:
--    its number is unique among the active rounds of the season and category, so
--    a cancelled round frees its number without copying any status elsewhere. The
--    Grand Prix name lives on the round because a circuit may host differently
--    named Grands Prix across categories and seasons.
CREATE TABLE circuits (
    id             bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name           text        NOT NULL UNIQUE,
    locality       text        NOT NULL,
    country_code   text        NOT NULL,
    is_active      boolean     NOT NULL GENERATED ALWAYS AS (deactivated_at IS NULL) STORED,
    deactivated_at timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT circuits_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT circuits_locality_not_blank CHECK (btrim(locality) <> ''),
    CONSTRAINT circuits_country_code_format CHECK (country_code ~ '^[a-z]{2}$')
);

CREATE TRIGGER circuits_set_updated_at BEFORE UPDATE ON circuits
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE rounds (
    id             bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    season_id      bigint      NOT NULL REFERENCES seasons (id) ON DELETE RESTRICT,
    category_id    bigint      NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    number         integer     NOT NULL,
    circuit_id     bigint      NOT NULL REFERENCES circuits (id) ON DELETE RESTRICT,
    gp_name        text        NOT NULL,
    is_active      boolean     NOT NULL GENERATED ALWAYS AS (deactivated_at IS NULL) STORED,
    deactivated_at timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT rounds_number_positive CHECK (number > 0),
    CONSTRAINT rounds_gp_name_not_blank CHECK (btrim(gp_name) <> '')
);

CREATE UNIQUE INDEX rounds_number_unique_per_championship
    ON rounds (season_id, category_id, number) WHERE is_active;
CREATE INDEX rounds_circuit_id_idx ON rounds (circuit_id);

CREATE TRIGGER rounds_set_updated_at BEFORE UPDATE ON rounds
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO circuits (name, locality, country_code)
SELECT DISTINCT ON (circuit) circuit, location, country_code
FROM races
ORDER BY circuit, event_id;

INSERT INTO rounds (season_id, category_id, number, circuit_id, gp_name)
SELECT r.season_id, r.category_id, r.round, c.id, e.name
FROM races r
JOIN circuits c ON c.name = r.circuit
JOIN events e ON e.id = r.event_id;

-- A race is the race of its round. The brief has no sprints, so the round holds
-- a single race; dropping races_round_id_unique is all a sprint would need.
ALTER TABLE races ADD COLUMN round_id bigint REFERENCES rounds (id) ON DELETE RESTRICT;
UPDATE races r SET round_id = ro.id
FROM rounds ro
WHERE ro.season_id = r.season_id AND ro.category_id = r.category_id AND ro.number = r.round;
ALTER TABLE races
    ALTER COLUMN round_id SET NOT NULL,
    ADD CONSTRAINT races_round_id_unique UNIQUE (round_id),
    DROP CONSTRAINT races_event_season_category_fkey,
    DROP CONSTRAINT races_round_unique_per_season_category,
    DROP COLUMN season_id,
    DROP COLUMN category_id,
    DROP COLUMN round,
    DROP COLUMN circuit,
    DROP COLUMN location,
    DROP COLUMN country_code;

ALTER TABLE events
    DROP CONSTRAINT events_id_season_category_unique,
    DROP COLUMN season_id,
    DROP COLUMN category_id,
    DROP COLUMN name;

CREATE INDEX events_starts_at_idx ON events (starts_at);

ALTER TABLE tyre_tests ADD COLUMN circuit_id bigint NOT NULL REFERENCES circuits (id) ON DELETE RESTRICT;
CREATE INDEX tyre_tests_circuit_id_idx ON tyre_tests (circuit_id);

-- 6. Each child declares the event kind and pins it as its own, so its FK cannot
--    point at an event of another kind: a race event hangs a race and a control
--    event hangs a technical control. The column is constant, so the DEFAULT is
--    enough and it need not be written on every INSERT. The kind of an event never
--    changes while it has a child, hence ON UPDATE RESTRICT.
ALTER TABLE races
    ADD COLUMN event_kind text NOT NULL DEFAULT 'race',
    ADD CONSTRAINT races_event_kind_check CHECK (event_kind = 'race'),
    DROP CONSTRAINT races_event_id_fkey,
    ADD CONSTRAINT races_event_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE RESTRICT;

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

-- 7. Exactly one child. An event without a child is of no kind: its label cannot
--    be composed. Both halves of the invariant are checked only at commit
--    (DEFERRABLE) so that inserting the event and its child in a single
--    transaction stays possible. "At most one" needs no trigger: the child's
--    primary key and its kind FK already guarantee it.
CREATE FUNCTION assert_event_has_child() RETURNS trigger AS $$
DECLARE
    has_child boolean;
BEGIN
    -- A kind added to events_kind_valid but not here must fail loudly, not let
    -- every event of that kind pass without a child.
    CASE NEW.kind
        WHEN 'race'              THEN has_child := EXISTS (SELECT 1 FROM races              WHERE event_id = NEW.id);
        WHEN 'tyre_test'         THEN has_child := EXISTS (SELECT 1 FROM tyre_tests         WHERE event_id = NEW.id);
        WHEN 'technical_control' THEN has_child := EXISTS (SELECT 1 FROM technical_controls WHERE event_id = NEW.id);
        ELSE RAISE EXCEPTION 'event % has kind %, which has no child table', NEW.id, NEW.kind;
    END CASE;

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

-- 8. The credited team is derived on every write, not only when the driver or the
--    event changes, so a direct write can no longer overwrite it.
DROP TRIGGER score_entries_set_team ON score_entries;
CREATE TRIGGER score_entries_set_team
    BEFORE INSERT OR UPDATE ON score_entries
    FOR EACH ROW EXECUTE FUNCTION set_score_entry_team();

-- 9. Only a race takes results. Scores and their change log point at the race
--    instead of the event, so a tyre test or a technical control cannot get any.
--    Notifications stay on events: any kind of event can be announced.
ALTER TABLE score_entries
    DROP CONSTRAINT score_entries_event_id_fkey,
    ADD CONSTRAINT score_entries_race_fkey
        FOREIGN KEY (event_id) REFERENCES races (event_id) ON DELETE RESTRICT;

ALTER TABLE score_change_log
    DROP CONSTRAINT score_change_log_event_id_fkey,
    ADD CONSTRAINT score_change_log_race_fkey
        FOREIGN KEY (event_id) REFERENCES races (event_id) ON DELETE RESTRICT;

-- 10. A race that already has results has been run, so what places it is final:
--     its date, its round, and the round's championship and circuit. Changing any
--     of them would move the results to another date, standings table or track,
--     and removing the race or its round would leave them hanging from something
--     nobody sees. The round number and the Grand Prix name stay editable: rounds
--     are renumbered after a cancellation and a typo must remain fixable. Every
--     guard fires only on a real change, so re-sending the same values passes.
CREATE FUNCTION race_has_results(race_event_id bigint) RETURNS boolean AS $$
    SELECT EXISTS (SELECT 1 FROM score_entries WHERE event_id = race_event_id);
$$ LANGUAGE sql STABLE;

CREATE FUNCTION assert_raced_event_is_final() RETURNS trigger AS $$
BEGIN
    IF (NEW.starts_at, NEW.deactivated_at) IS DISTINCT FROM (OLD.starts_at, OLD.deactivated_at)
        AND race_has_results(NEW.id) THEN
        RAISE EXCEPTION 'event % has results: its date and status are final', NEW.id
            USING ERRCODE = 'RC001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER events_keep_raced_event_final
    BEFORE UPDATE OF starts_at, deactivated_at ON events
    FOR EACH ROW EXECUTE FUNCTION assert_raced_event_is_final();

CREATE FUNCTION assert_raced_race_keeps_round() RETURNS trigger AS $$
BEGIN
    IF NEW.round_id IS DISTINCT FROM OLD.round_id AND race_has_results(NEW.event_id) THEN
        RAISE EXCEPTION 'race % has results: its round is final', NEW.event_id
            USING ERRCODE = 'RC001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER races_keep_raced_round
    BEFORE UPDATE OF round_id ON races
    FOR EACH ROW EXECUTE FUNCTION assert_raced_race_keeps_round();

CREATE FUNCTION assert_raced_round_is_final() RETURNS trigger AS $$
BEGIN
    IF (NEW.season_id, NEW.category_id, NEW.circuit_id, NEW.deactivated_at)
            IS DISTINCT FROM (OLD.season_id, OLD.category_id, OLD.circuit_id, OLD.deactivated_at)
        AND EXISTS (SELECT 1 FROM races WHERE round_id = NEW.id AND race_has_results(event_id)) THEN
        RAISE EXCEPTION 'round % has a race with results: its championship, circuit and status are final', NEW.id
            USING ERRCODE = 'RC001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER rounds_keep_raced_round_final
    BEFORE UPDATE OF season_id, category_id, circuit_id, deactivated_at ON rounds
    FOR EACH ROW EXECUTE FUNCTION assert_raced_round_is_final();
