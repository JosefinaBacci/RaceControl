-- 000022_assignment_lifecycle: a driver can leave a team, and the history only
-- grows forward. A release is a row without a team from a date: assigned_team()
-- then answers NULL, so the guards that need a team refuse to score or number the
-- driver and the roster stops listing them. The periods stay append-only and every
-- new row extends the history instead of rewriting it.

-- 1. A release carries no team. The NULL is the fact "no team from this date",
--    not a missing value, and the latest row still wins.
ALTER TABLE driver_team_assignments ALTER COLUMN team_id DROP NOT NULL;

-- 2. A new assignment may only extend the history, because the team derived for a
--    past date must never change: a score entry already stored the team its driver
--    raced for, and a retroactive row would contradict it. The row has to start
--    after the driver's last assignment and after the last race the driver already
--    has results in. RC004 lets the service answer with a 4xx.
--
--    AFTER INSERT, not BEFORE: the guard must run only for rows that were really
--    inserted, so the idempotent seed can re-run its assignments under ON CONFLICT
--    DO NOTHING. It excludes the new row when reading the last assignment, since
--    AFTER INSERT already sees it.
CREATE FUNCTION assert_assignment_extends_history() RETURNS trigger AS $$
DECLARE
    last_assignment timestamptz;
    last_result     timestamptz;
BEGIN
    SELECT max(effective_from) INTO last_assignment
    FROM driver_team_assignments
    WHERE driver_id = NEW.driver_id AND id <> NEW.id;

    IF last_assignment IS NOT NULL AND NEW.effective_from <= last_assignment THEN
        RAISE EXCEPTION 'assignment of driver % starts at or before the previous one', NEW.driver_id
            USING ERRCODE = 'RC004';
    END IF;

    SELECT max(events.starts_at) INTO last_result
    FROM score_entries
    JOIN events ON events.id = score_entries.event_id
    WHERE score_entries.driver_id = NEW.driver_id;

    IF last_result IS NOT NULL AND NEW.effective_from <= last_result THEN
        RAISE EXCEPTION 'assignment of driver % would rewrite a race they already scored in', NEW.driver_id
            USING ERRCODE = 'RC004';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER driver_team_assignments_extends_history
    AFTER INSERT ON driver_team_assignments
    FOR EACH ROW EXECUTE FUNCTION assert_assignment_extends_history();
