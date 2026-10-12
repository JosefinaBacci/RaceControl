-- 000012 down: revierte la jerarquía, vuelve al modelo plano de events + event_types

DROP INDEX IF EXISTS technical_controls_event_id_idx;
DROP TABLE IF EXISTS technical_controls CASCADE;

DROP INDEX IF EXISTS tyre_tests_event_id_idx;
DROP TABLE IF EXISTS tyre_tests CASCADE;

DROP INDEX IF EXISTS races_event_id_idx;
DROP TABLE IF EXISTS races CASCADE;

-- 6. events: devolvemos event_type_id
ALTER TABLE events ADD COLUMN event_type_id bigint NOT NULL REFERENCES event_types (id) ON DELETE RESTRICT;

-- 7. event_types: volvemos a crear (sin datos, setval se hará con make db-reset)
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

-- 8. setval vacío para event_types (reemplazado en make db-reset)
SELECT setval(pg_get_serial_sequence('event_types', 'id'), 0);