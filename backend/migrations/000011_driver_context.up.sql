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
    id          bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    driver_id   bigint      NOT NULL REFERENCES drivers (id) ON DELETE RESTRICT,
    season_id   bigint      NOT NULL REFERENCES seasons (id) ON DELETE RESTRICT,
    category_id bigint      NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    number      integer     NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT driver_season_numbers_number_range CHECK (number BETWEEN 0 AND 999),
    CONSTRAINT driver_season_numbers_unique_per_driver_season UNIQUE (driver_id, season_id),
    -- A number is unique inside its championship, not the whole season: F1 and
    -- F2 both number their cars from 1, so the same number may exist twice.
    CONSTRAINT driver_season_numbers_unique_per_season_category UNIQUE (season_id, category_id, number)
);

-- Listing a season's numbers (the public directory) scans by season.
CREATE INDEX driver_season_numbers_season_id_idx ON driver_season_numbers (season_id);

-- The category is derived from the driver's team, the same derivation the score
-- entries use, so it is materialised here and the UNIQUE above can hold. The
-- reference date is the start of the season.
CREATE FUNCTION set_driver_season_number_category() RETURNS trigger AS $$
DECLARE
    season_start timestamptz;
BEGIN
    SELECT make_timestamptz(year, 1, 1, 0, 0, 0, 'UTC') INTO season_start
    FROM seasons WHERE id = NEW.season_id;

    NEW.category_id := (SELECT category_id FROM teams WHERE id = assigned_team(NEW.driver_id, season_start));

    IF NEW.category_id IS NULL THEN
        RAISE EXCEPTION 'driver % has no team assignment for season %', NEW.driver_id, NEW.season_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER driver_season_numbers_set_category
    BEFORE INSERT OR UPDATE OF driver_id, season_id ON driver_season_numbers
    FOR EACH ROW EXECUTE FUNCTION set_driver_season_number_category();

CREATE TRIGGER driver_season_numbers_set_updated_at BEFORE UPDATE ON driver_season_numbers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();