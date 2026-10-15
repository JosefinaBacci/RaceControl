-- 000022 down: releases go away and a team is required again, so the pair is
-- reversible. Releases cannot survive a NOT NULL team and the append-only guard
-- would block removing them, so it steps aside for the cleanup and is restored
-- right after.

DROP TRIGGER driver_team_assignments_extends_history ON driver_team_assignments;
DROP FUNCTION assert_assignment_extends_history();

DROP TRIGGER driver_team_assignments_append_only ON driver_team_assignments;
DELETE FROM driver_team_assignments WHERE team_id IS NULL;
CREATE TRIGGER driver_team_assignments_append_only
    BEFORE UPDATE OR DELETE ON driver_team_assignments
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

ALTER TABLE driver_team_assignments ALTER COLUMN team_id SET NOT NULL;
