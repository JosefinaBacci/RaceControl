ALTER TABLE score_acknowledgements DROP CONSTRAINT score_acknowledgements_unique_per_version;
ALTER TABLE score_acknowledgements DROP CONSTRAINT score_acknowledgements_change_log_id_fkey;
ALTER TABLE score_acknowledgements DROP COLUMN change_log_id;

ALTER TABLE score_acknowledgements ADD COLUMN event_id bigint NOT NULL;
ALTER TABLE score_acknowledgements ADD CONSTRAINT score_acknowledgements_event_id_fkey
    FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE RESTRICT;
ALTER TABLE score_acknowledgements ADD CONSTRAINT score_acknowledgements_unique_per_team
    UNIQUE (event_id, team_id);
