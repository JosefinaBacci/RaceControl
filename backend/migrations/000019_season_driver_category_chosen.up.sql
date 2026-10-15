-- 000019_season_driver_category_chosen: a driver's season number belongs to a
-- championship the admin picks, not to the team the driver had on January 1st.
-- Deriving it punished the two cases that stray from that date: a rookie who
-- joins mid-season had no assignment then and could not get a number, and a
-- driver who changes category kept the old one. Uniqueness moves to
-- (driver, season, category), so a promoted driver holds one number per
-- championship, and the trigger now rejects a category the driver never raced in
-- during the season.

-- 1. One number per (driver, season, category) instead of per (driver, season).
ALTER TABLE season_drivers
    DROP CONSTRAINT season_drivers_unique_per_driver_season,
    ADD CONSTRAINT season_drivers_unique_per_driver_season_category
        UNIQUE (driver_id, season_id, category_id);

-- 2. The category is an input now, so the trigger validates instead of deriving.
-- The function is renamed rather than dropped and recreated: sqlc does not track a
-- function renamed in an earlier migration and fails on the DROP.
DROP TRIGGER season_drivers_set_category ON season_drivers;
ALTER FUNCTION set_season_driver_category() RENAME TO assert_season_driver_category;

-- The driver must have raced in that category at some point during the season.
-- The season window is the calendar year, the same convention 000011 used, and
-- an assignment is active from its effective_from until the next one starts
-- (implicit periods, 000007), so the check is that the two windows overlap.
CREATE OR REPLACE FUNCTION assert_season_driver_category() RETURNS trigger AS $$
DECLARE
    season_start   timestamptz;
    season_end     timestamptz;
    has_assignment boolean;
BEGIN
    SELECT make_timestamptz(year, 1, 1, 0, 0, 0, 'UTC'),
           make_timestamptz(year + 1, 1, 1, 0, 0, 0, 'UTC')
    INTO season_start, season_end
    FROM seasons WHERE id = NEW.season_id;

    SELECT EXISTS (
        SELECT 1
        FROM driver_team_assignments assignment
        JOIN teams ON teams.id = assignment.team_id
        WHERE assignment.driver_id = NEW.driver_id
          AND teams.category_id = NEW.category_id
          AND assignment.effective_from < season_end
          AND COALESCE(
                  (SELECT min(later.effective_from)
                   FROM driver_team_assignments later
                   WHERE later.driver_id = assignment.driver_id
                     AND later.effective_from > assignment.effective_from),
                  'infinity'::timestamptz) > season_start
    ) INTO has_assignment;

    IF NOT has_assignment THEN
        RAISE EXCEPTION 'driver % has no assignment in category % during season %',
            NEW.driver_id, NEW.category_id, NEW.season_id
            USING ERRCODE = 'RC003';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER season_drivers_assert_category
    BEFORE INSERT OR UPDATE OF driver_id, season_id, category_id ON season_drivers
    FOR EACH ROW EXECUTE FUNCTION assert_season_driver_category();
