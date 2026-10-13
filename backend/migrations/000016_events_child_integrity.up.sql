-- 000016_events_child_integrity: an event declares its kind and has exactly one
-- child of that kind. A race always has a round, the event name is not duplicated
-- on the children, a result always derives its credited team and a race that
-- already has results keeps its date.

-- 1. The name belongs to the event. The child duplicated it and the two could diverge.
ALTER TABLE races DROP COLUMN name;
ALTER TABLE tyre_tests DROP COLUMN name;
ALTER TABLE technical_controls DROP COLUMN name;

-- 2. A race always has a round. With NOT NULL, the per-championship UNIQUE can no
--    longer be dodged by loading the round as NULL.
ALTER TABLE races DROP CONSTRAINT races_round_positive;
ALTER TABLE races
    ALTER COLUMN round SET NOT NULL,
    ADD CONSTRAINT races_round_positive CHECK (round > 0);

-- 3. At most one child per event. The unique constraint replaces the plain index,
--    whose event_id lookups the unique prefix already covers.
DROP INDEX races_event_id_idx;
ALTER TABLE races ADD CONSTRAINT races_event_id_unique UNIQUE (event_id);

DROP INDEX tyre_tests_event_id_idx;
ALTER TABLE tyre_tests ADD CONSTRAINT tyre_tests_event_id_unique UNIQUE (event_id);

DROP INDEX technical_controls_event_id_idx;
ALTER TABLE technical_controls ADD CONSTRAINT technical_controls_event_id_unique UNIQUE (event_id);

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
--    DEFAULT is enough and it need not be written on every INSERT.
ALTER TABLE races
    ADD COLUMN event_kind text NOT NULL DEFAULT 'race',
    ADD CONSTRAINT races_event_kind_check CHECK (event_kind = 'race'),
    DROP CONSTRAINT races_event_id_fkey,
    ADD CONSTRAINT races_event_kind_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE tyre_tests
    ADD COLUMN event_kind text NOT NULL DEFAULT 'tyre_test',
    ADD CONSTRAINT tyre_tests_event_kind_check CHECK (event_kind = 'tyre_test'),
    DROP CONSTRAINT tyre_tests_event_id_fkey,
    ADD CONSTRAINT tyre_tests_event_kind_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE technical_controls
    ADD COLUMN event_kind text NOT NULL DEFAULT 'technical_control',
    ADD CONSTRAINT technical_controls_event_kind_check CHECK (event_kind = 'technical_control'),
    DROP CONSTRAINT technical_controls_event_id_fkey,
    ADD CONSTRAINT technical_controls_event_kind_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE CASCADE;

-- 6. Exactly one child. An event without a child is of no kind: neither its name nor
--    its kind can be derived. Both halves of the invariant are checked only at
--    commit (DEFERRABLE) so that inserting the event and its child in a single
--    transaction stays possible.
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

-- 7. The credited team is derived on every write, not only when the driver or the
--    event changes, so a direct write can no longer overwrite it.
DROP TRIGGER score_entries_set_team ON score_entries;
CREATE TRIGGER score_entries_set_team
    BEFORE INSERT OR UPDATE ON score_entries
    FOR EACH ROW EXECUTE FUNCTION set_score_entry_team();

-- 8. A race that already has results has been run: its date is final. The guard
--    fires only when the date really changes, so editing anything else on the
--    event still works.
CREATE FUNCTION assert_event_date_is_final() RETURNS trigger AS $$
BEGIN
    IF NEW.starts_at IS DISTINCT FROM OLD.starts_at
        AND EXISTS (SELECT 1 FROM score_entries WHERE event_id = NEW.id) THEN
        RAISE EXCEPTION 'event % has results: its date is final', NEW.id
            USING ERRCODE = 'RC001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER events_keep_final_date
    BEFORE UPDATE OF starts_at ON events
    FOR EACH ROW EXECUTE FUNCTION assert_event_date_is_final();
