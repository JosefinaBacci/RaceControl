-- 000018_scores_category_and_naming: a driver scores only in a race of their
-- team's category, a score change lists each driver once, an event lasts a
-- positive span of time, and the objects of season_drivers carry its name.

-- 1. The credited team is already derived from the assignment history on the
--    race date; it must also run in the race's category, or an F2 driver could
--    be scored in an F1 race. RC002 lets the service answer with a 4xx, since a
--    wrong driver is an input mistake, not a bug.
CREATE OR REPLACE FUNCTION set_score_entry_team() RETURNS trigger AS $$
BEGIN
    NEW.team_id := assigned_team(NEW.driver_id, (SELECT starts_at FROM events WHERE id = NEW.event_id));

    IF NEW.team_id IS NULL THEN
        RAISE EXCEPTION 'driver % has no team assignment for event %', NEW.driver_id, NEW.event_id;
    END IF;

    IF (SELECT category_id FROM teams WHERE id = NEW.team_id) IS DISTINCT FROM (
        SELECT ro.category_id
        FROM races r
        JOIN rounds ro ON ro.id = r.round_id
        WHERE r.event_id = NEW.event_id
    ) THEN
        RAISE EXCEPTION 'driver % races for a team of another category than event %', NEW.driver_id, NEW.event_id
            USING ERRCODE = 'RC002';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. A change records at most one line per driver: two lines would leave its old
--    and new points ambiguous. The unique index leads with change_log_id, so it
--    replaces the plain index on that column.
DROP INDEX score_change_log_items_change_log_id_idx;
ALTER TABLE score_change_log_items
    ADD CONSTRAINT score_change_log_items_driver_unique_per_change UNIQUE (change_log_id, driver_id);

-- 3. An event that ends when it starts has no duration and means nothing on a
--    calendar.
ALTER TABLE events
    DROP CONSTRAINT events_ends_after_start,
    ADD CONSTRAINT events_ends_after_start CHECK (ends_at > starts_at);

-- 4. 000013 renamed the table but not what hangs from it.
DO $$
DECLARE
    old_name text;
BEGIN
    FOR old_name IN
        SELECT conname FROM pg_constraint
        WHERE conrelid = 'season_drivers'::regclass AND conname LIKE 'driver_season_numbers%'
    LOOP
        EXECUTE format('ALTER TABLE season_drivers RENAME CONSTRAINT %I TO %I',
            old_name, replace(old_name, 'driver_season_numbers', 'season_drivers'));
    END LOOP;
END;
$$;

ALTER INDEX driver_season_numbers_season_id_idx RENAME TO season_drivers_season_id_idx;
ALTER SEQUENCE driver_season_numbers_id_seq RENAME TO season_drivers_id_seq;
ALTER TRIGGER driver_season_numbers_set_category ON season_drivers RENAME TO season_drivers_set_category;
ALTER TRIGGER driver_season_numbers_set_updated_at ON season_drivers RENAME TO season_drivers_set_updated_at;
ALTER FUNCTION set_driver_season_number_category() RENAME TO set_season_driver_category;
