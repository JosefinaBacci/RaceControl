-- Calendar: the event catalogue (race, tyre test, technical control) and the
-- events themselves. One table covers every kind of event because they differ
-- in attributes, not in structure.

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

CREATE TABLE events (
    id            bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_type_id bigint      NOT NULL REFERENCES event_types (id) ON DELETE RESTRICT,
    season_id     bigint      NOT NULL REFERENCES seasons (id) ON DELETE RESTRICT,
    category_id   bigint      NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    name          text        NOT NULL,
    circuit       text        NOT NULL,
    location      text        NOT NULL,
    country_code  text        NOT NULL,
    starts_at     timestamptz NOT NULL,
    ends_at       timestamptz NOT NULL,
    round         integer,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT events_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT events_circuit_not_blank CHECK (btrim(circuit) <> ''),
    CONSTRAINT events_country_code_format CHECK (country_code ~ '^[a-z]{2}$'),
    CONSTRAINT events_ends_after_start CHECK (ends_at >= starts_at),
    CONSTRAINT events_round_positive CHECK (round IS NULL OR round > 0),
    -- A round number identifies a race inside a season and category. Null rounds
    -- (tyre tests and controls) stay out of the constraint because NULLs are
    -- distinct in a unique index.
    CONSTRAINT events_round_unique_per_season_category UNIQUE (season_id, category_id, round)
);

CREATE INDEX events_season_starts_at_idx ON events (season_id, starts_at);
-- The (season_id, category_id, round) unique index already serves lookups by
-- season and category, so no separate category index is needed.
CREATE INDEX events_event_type_id_idx ON events (event_type_id);

CREATE TRIGGER events_set_updated_at BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
