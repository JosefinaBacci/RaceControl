-- 000012 down: reverts the hierarchy, back to the flat events + event_types model.

DROP INDEX IF EXISTS technical_controls_event_id_idx;
DROP TABLE IF EXISTS technical_controls CASCADE;

DROP INDEX IF EXISTS tyre_tests_event_id_idx;
DROP TABLE IF EXISTS tyre_tests CASCADE;

DROP INDEX IF EXISTS races_event_id_idx;
DROP TABLE IF EXISTS races CASCADE;

-- event_types first: events.event_type_id references it. The column comes back
-- nullable because dropping event_types left no type to backfill for the rows
-- that already exist, so NOT NULL cannot be re-imposed honestly.
CREATE TABLE event_types (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       text        NOT NULL UNIQUE,
    name       text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT event_types_code_format CHECK (code ~ '^[a-z0-9_]{2,32}$')
);
CREATE TRIGGER event_types_set_updated_at BEFORE UPDATE ON event_types
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE events ADD COLUMN event_type_id bigint REFERENCES event_types (id) ON DELETE RESTRICT;

-- Reset the sequence so the next insert starts at 1 (make db-reset replaces this).
SELECT setval(pg_get_serial_sequence('event_types', 'id'), 1, false);
