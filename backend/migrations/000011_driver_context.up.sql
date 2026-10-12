-- Driver context belongs where it varies: the race number is a fact of the
-- (driver, season) pairing, and the starter/reserve role is a fact of the team
-- membership period, already append-only. The driver row keeps only person
-- facts, so number and is_starter leave drivers.

ALTER TABLE drivers DROP COLUMN number;
ALTER TABLE drivers DROP COLUMN is_starter;

-- The role rides on the membership period: a role change mid-season (reserve
-- promoted, regular replaced, team transfer) is a new assignment row, never an
-- edit, so the history stays append-only. The DEFAULT only satisfies the ALTER
-- on existing rows; inserts always pass the explicit value.
ALTER TABLE driver_team_assignments ADD COLUMN is_starter boolean NOT NULL DEFAULT true;

-- The number the championship assigns for the season. It does not change when
-- the driver changes teams mid-season, so no history: one row per (driver,
-- season), updated only if the championship reassigns it (rare).
CREATE TABLE driver_season_numbers (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    driver_id  bigint      NOT NULL REFERENCES drivers (id) ON DELETE RESTRICT,
    season_id  bigint      NOT NULL REFERENCES seasons (id) ON DELETE RESTRICT,
    number     integer     NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT driver_season_numbers_number_range CHECK (number BETWEEN 0 AND 999),
    CONSTRAINT driver_season_numbers_unique_per_driver_season UNIQUE (driver_id, season_id)
);

-- Listing a season's numbers (the public directory) scans by season.
CREATE INDEX driver_season_numbers_season_id_idx ON driver_season_numbers (season_id);

CREATE TRIGGER driver_season_numbers_set_updated_at BEFORE UPDATE ON driver_season_numbers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();