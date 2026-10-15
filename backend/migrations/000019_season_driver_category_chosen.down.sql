-- 000019 down: the category is derived from the team at January 1st again and a
-- driver may hold a single season number. Re-adding the old unique key fails if a
-- driver already holds numbers in two categories in one season: that data has no
-- representation in the old model, so the down aborts and `make db-reset` is the
-- way out.

-- 2. The derivation trigger returns; the function goes back to its old name and
-- body, the same shape 000018 left it in.
DROP TRIGGER season_drivers_assert_category ON season_drivers;
ALTER FUNCTION assert_season_driver_category() RENAME TO set_season_driver_category;

CREATE OR REPLACE FUNCTION set_season_driver_category() RETURNS trigger AS $$
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

CREATE TRIGGER season_drivers_set_category
    BEFORE INSERT OR UPDATE OF driver_id, season_id ON season_drivers
    FOR EACH ROW EXECUTE FUNCTION set_season_driver_category();

-- 1. One number per (driver, season) again.
ALTER TABLE season_drivers
    DROP CONSTRAINT season_drivers_unique_per_driver_season_category,
    ADD CONSTRAINT season_drivers_unique_per_driver_season
        UNIQUE (driver_id, season_id);
