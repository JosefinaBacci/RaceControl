-- 000020 down: brings back the mirror guard dropped in the matching up, so the
-- pair is reversible even though the application does not rely on it. The body is
-- the one 000016 created: if the event is still there, it must keep a child.

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
