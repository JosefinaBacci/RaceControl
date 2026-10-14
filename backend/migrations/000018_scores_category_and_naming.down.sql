-- 000018 down: the season_drivers objects get their old names back, a
-- zero-length event is allowed again, a change may repeat a driver and the
-- credited team is no longer checked against the race's category.

-- 4. Old names.
ALTER FUNCTION set_season_driver_category() RENAME TO set_driver_season_number_category;
ALTER TRIGGER season_drivers_set_updated_at ON season_drivers RENAME TO driver_season_numbers_set_updated_at;
ALTER TRIGGER season_drivers_set_category ON season_drivers RENAME TO driver_season_numbers_set_category;
ALTER SEQUENCE season_drivers_id_seq RENAME TO driver_season_numbers_id_seq;
ALTER INDEX season_drivers_season_id_idx RENAME TO driver_season_numbers_season_id_idx;

DO $$
DECLARE
    new_name text;
BEGIN
    FOR new_name IN
        SELECT conname FROM pg_constraint
        WHERE conrelid = 'season_drivers'::regclass AND conname LIKE 'season_drivers%'
    LOOP
        EXECUTE format('ALTER TABLE season_drivers RENAME CONSTRAINT %I TO %I',
            new_name, replace(new_name, 'season_drivers', 'driver_season_numbers'));
    END LOOP;
END;
$$;

-- 3. A zero-length event is valid again.
ALTER TABLE events
    DROP CONSTRAINT events_ends_after_start,
    ADD CONSTRAINT events_ends_after_start CHECK (ends_at >= starts_at);

-- 2. A change may list a driver more than once.
ALTER TABLE score_change_log_items DROP CONSTRAINT score_change_log_items_driver_unique_per_change;
CREATE INDEX score_change_log_items_change_log_id_idx ON score_change_log_items (change_log_id);

-- 1. The credited team is derived without checking its category.
CREATE OR REPLACE FUNCTION set_score_entry_team() RETURNS trigger AS $$
BEGIN
    NEW.team_id := assigned_team(NEW.driver_id, (SELECT starts_at FROM events WHERE id = NEW.event_id));

    IF NEW.team_id IS NULL THEN
        RAISE EXCEPTION 'driver % has no team assignment for event %', NEW.driver_id, NEW.event_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
