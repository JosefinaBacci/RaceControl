-- A result keeps the team the driver raced for, but that team is never entered
-- by hand: a trigger fills it from the assignment history, so the value on the
-- result is always the derivation of that history for the event's date.

ALTER TABLE score_entries DROP CONSTRAINT score_entries_driver_team_fk;
ALTER TABLE score_entries ADD CONSTRAINT score_entries_driver_id_fkey
    FOREIGN KEY (driver_id) REFERENCES drivers (id) ON DELETE RESTRICT;
ALTER TABLE score_entries ADD CONSTRAINT score_entries_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE RESTRICT;

-- drivers.team_id is the current assignment, which the history already answers.
ALTER TABLE drivers DROP CONSTRAINT drivers_id_team_unique;
ALTER TABLE drivers DROP COLUMN team_id;

-- The audited team is the same derivation, kept on the item so the change stays
-- readable even after the entries it describes were replaced.
ALTER TABLE score_change_log_items ADD COLUMN team_id bigint NOT NULL;
ALTER TABLE score_change_log_items ADD CONSTRAINT score_change_log_items_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE RESTRICT;

CREATE FUNCTION set_score_entry_team() RETURNS trigger AS $$
BEGIN
    NEW.team_id := assigned_team(NEW.driver_id, (SELECT starts_at FROM events WHERE id = NEW.event_id));

    IF NEW.team_id IS NULL THEN
        RAISE EXCEPTION 'driver % has no team assignment for event %', NEW.driver_id, NEW.event_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER score_entries_set_team
    BEFORE INSERT OR UPDATE OF event_id, driver_id ON score_entries
    FOR EACH ROW EXECUTE FUNCTION set_score_entry_team();

CREATE FUNCTION set_score_change_item_team() RETURNS trigger AS $$
DECLARE
    event_starts_at timestamptz;
BEGIN
    SELECT events.starts_at INTO event_starts_at
    FROM score_change_log
    JOIN events ON events.id = score_change_log.event_id
    WHERE score_change_log.id = NEW.change_log_id;

    NEW.team_id := assigned_team(NEW.driver_id, event_starts_at);

    IF NEW.team_id IS NULL THEN
        RAISE EXCEPTION 'driver % has no team assignment for change %', NEW.driver_id, NEW.change_log_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER score_change_log_items_set_team
    BEFORE INSERT OR UPDATE OF change_log_id, driver_id ON score_change_log_items
    FOR EACH ROW EXECUTE FUNCTION set_score_change_item_team();
