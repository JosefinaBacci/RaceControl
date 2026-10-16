-- 000023_acknowledgement_team: only an administrator of an affected team can
-- acknowledge a score change (F6 / B.2).
--
-- The team on an acknowledgement is never entered by hand: it is the team of the
-- user who acknowledges, so a caller cannot confirm on behalf of another team by
-- sending a different team_id. A FIA administrator has no team and therefore can
-- never acknowledge. The team must also appear in the items of the change, since
-- a team is only asked to confirm a change that touched its scores. A missing
-- team, a foreign change and an unknown change all fail the same check, and RC005
-- lets the service answer with a 4xx.

CREATE FUNCTION set_score_acknowledgement_team() RETURNS trigger AS $$
BEGIN
    NEW.team_id := (SELECT team_id FROM users WHERE id = NEW.acknowledged_by);

    IF NOT EXISTS (
        SELECT 1
        FROM score_change_log_items
        WHERE change_log_id = NEW.change_log_id AND team_id = NEW.team_id
    ) THEN
        RAISE EXCEPTION 'user % cannot acknowledge score change %: their team is not affected by it',
            NEW.acknowledged_by, NEW.change_log_id
            USING ERRCODE = 'RC005';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER score_acknowledgements_set_team
    BEFORE INSERT ON score_acknowledgements
    FOR EACH ROW EXECUTE FUNCTION set_score_acknowledgement_team();
