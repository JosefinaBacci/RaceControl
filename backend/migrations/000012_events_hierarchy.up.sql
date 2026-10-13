-- 000012_events_hierarchy: Jerarquía de eventos → races / tyre_tests / technical_controls
-- todos los hijos apuntan al mismo events.id (el GP/ronda del calendario oficial).
-- No hay event_types/event_type_id ni enum: cada tipo tiene su tabla propia.

-- 1. events: quitamos event_type_id y event_types (ya no los usamos)
ALTER TABLE events DROP COLUMN event_type_id;
DROP TABLE IF EXISTS event_types CASCADE;

-- 2. races: la carrera concreta del fin de semana (1 por GP hoy)
CREATE TABLE races (
    id          bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id    bigint      NOT NULL REFERENCES events (id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    starts_at   timestamptz NOT NULL,
    ends_at     timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT races_ends_after_start CHECK (ends_at > starts_at)
);
CREATE INDEX races_event_id_idx ON races (event_id);
CREATE TRIGGER races_set_updated_at BEFORE UPDATE ON races
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 3. tyre_tests: prueba de neumáticos, siempre colgando del evento que la agrupa
CREATE TABLE tyre_tests (
    id          bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id    bigint      NOT NULL REFERENCES events (id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    starts_at   timestamptz NOT NULL,
    ends_at     timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT tyre_tests_ends_after_start CHECK (ends_at > starts_at)
);
CREATE INDEX tyre_tests_event_id_idx ON tyre_tests (event_id);
CREATE TRIGGER tyre_tests_set_updated_at BEFORE UPDATE ON tyre_tests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 4. technical_controls: control técnico. The timing lives in events, like every
-- other event, so this table only adds the control-specific attributes.
CREATE TABLE technical_controls (
    id           bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id     bigint      NOT NULL REFERENCES events (id) ON DELETE RESTRICT,
    name         text        NOT NULL,
    control_type text        NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX technical_controls_event_id_idx ON technical_controls (event_id);
CREATE TRIGGER technical_controls_set_updated_at BEFORE UPDATE ON technical_controls
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
