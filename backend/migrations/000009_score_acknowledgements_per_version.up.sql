-- A team confirms a published version of the scores, not the event: correcting
-- the scores afterwards leaves the previous confirmation stale and asks the
-- affected teams to confirm again. The version is the score_change_log row that
-- was published, which already carries the event and the moment of the change.

ALTER TABLE score_acknowledgements DROP CONSTRAINT score_acknowledgements_unique_per_team;
ALTER TABLE score_acknowledgements DROP CONSTRAINT score_acknowledgements_event_id_fkey;
ALTER TABLE score_acknowledgements DROP COLUMN event_id;

ALTER TABLE score_acknowledgements ADD COLUMN change_log_id bigint NOT NULL;
ALTER TABLE score_acknowledgements ADD CONSTRAINT score_acknowledgements_change_log_id_fkey
    FOREIGN KEY (change_log_id) REFERENCES score_change_log (id) ON DELETE RESTRICT;
ALTER TABLE score_acknowledgements ADD CONSTRAINT score_acknowledgements_unique_per_version
    UNIQUE (change_log_id, team_id);
