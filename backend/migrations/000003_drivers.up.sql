-- Drivers belong to a team. The category is not stored: it is derived from the
-- team, which prevents a Formula 2 driver from sitting on a Formula 1 team.

CREATE TABLE drivers (
    id             bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_id        bigint      NOT NULL REFERENCES teams (id) ON DELETE RESTRICT,
    first_name     text        NOT NULL,
    last_name      text        NOT NULL,
    license_number text        NOT NULL UNIQUE,
    number         integer,
    country_code   text        NOT NULL,
    is_starter     boolean     NOT NULL DEFAULT true,
    is_active      boolean     NOT NULL GENERATED ALWAYS AS (deactivated_at IS NULL) STORED,
    deactivated_at timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT drivers_first_name_not_blank CHECK (btrim(first_name) <> ''),
    CONSTRAINT drivers_last_name_not_blank CHECK (btrim(last_name) <> ''),
    CONSTRAINT drivers_license_number_length CHECK (char_length(license_number) BETWEEN 3 AND 32),
    CONSTRAINT drivers_country_code_format CHECK (country_code ~ '^[a-z]{2}$'),
    CONSTRAINT drivers_number_range CHECK (number IS NULL OR number BETWEEN 0 AND 999),
    -- Redundant with the primary key, but it lets score_entries carry a
    -- composite foreign key that pins the driver and its team together.
    CONSTRAINT drivers_id_team_unique UNIQUE (id, team_id)
);

CREATE INDEX drivers_team_id_idx ON drivers (team_id);

CREATE TRIGGER drivers_set_updated_at BEFORE UPDATE ON drivers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
