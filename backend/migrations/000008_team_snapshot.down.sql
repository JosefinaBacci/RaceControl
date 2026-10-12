DROP TRIGGER IF EXISTS score_entries_set_team ON score_entries;
DROP TRIGGER IF EXISTS score_change_log_items_set_team ON score_change_log_items;
DROP FUNCTION IF EXISTS set_score_entry_team();
DROP FUNCTION IF EXISTS set_score_change_item_team();

ALTER TABLE score_change_log_items DROP CONSTRAINT score_change_log_items_team_id_fkey;
ALTER TABLE score_change_log_items DROP COLUMN team_id;

-- Restore the current team on the driver from the latest assignment.
ALTER TABLE drivers ADD COLUMN team_id bigint;

UPDATE drivers
SET team_id = (
    SELECT driver_team_assignments.team_id
    FROM driver_team_assignments
    WHERE driver_team_assignments.driver_id = drivers.id
    ORDER BY driver_team_assignments.effective_from DESC
    LIMIT 1
);

ALTER TABLE drivers ALTER COLUMN team_id SET NOT NULL;
ALTER TABLE drivers ADD CONSTRAINT drivers_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE RESTRICT;
ALTER TABLE drivers ADD CONSTRAINT drivers_id_team_unique UNIQUE (id, team_id);

ALTER TABLE score_entries DROP CONSTRAINT score_entries_team_id_fkey;
ALTER TABLE score_entries DROP CONSTRAINT score_entries_driver_id_fkey;
ALTER TABLE score_entries ADD CONSTRAINT score_entries_driver_team_fk
    FOREIGN KEY (driver_id, team_id) REFERENCES drivers (id, team_id) ON DELETE RESTRICT;
