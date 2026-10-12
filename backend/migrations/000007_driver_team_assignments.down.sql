DROP TRIGGER IF EXISTS driver_team_assignments_append_only ON driver_team_assignments;
DROP FUNCTION IF EXISTS reject_mutation();
DROP FUNCTION IF EXISTS assigned_team(bigint, timestamptz);
DROP TABLE IF EXISTS driver_team_assignments;
