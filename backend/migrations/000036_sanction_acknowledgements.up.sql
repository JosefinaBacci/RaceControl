-- 000036_sanction_acknowledgements: every team of the championship confirms it
-- was notified of each version of a sanction (US26).
--
-- The statement asks that sanctions reach every team and that the teams record
-- they were notified, so the recipients of a version are all the teams of the
-- accountable team's category that are active when the version is written. They
-- are frozen in sanction_recipients at that moment: a team created later owes
-- nothing for an old version, and a team deactivated later still shows whether
-- it confirmed what it was sent.

CREATE TABLE sanction_recipients (
    sanction_version_id bigint NOT NULL REFERENCES sanction_versions (id) ON DELETE RESTRICT,
    team_id             bigint NOT NULL REFERENCES teams (id) ON DELETE RESTRICT,

    PRIMARY KEY (sanction_version_id, team_id)
);

CREATE INDEX sanction_recipients_team_id_idx ON sanction_recipients (team_id);

CREATE TRIGGER sanction_recipients_append_only
    BEFORE UPDATE OR DELETE ON sanction_recipients
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

CREATE FUNCTION notify_sanction_version() RETURNS trigger AS $$
BEGIN
    INSERT INTO sanction_recipients (sanction_version_id, team_id)
    SELECT NEW.id, recipient.id
    FROM sanctions
    JOIN teams accountable ON accountable.id = sanctions.accountable_team_id
    JOIN teams recipient ON recipient.category_id = accountable.category_id
    WHERE sanctions.id = NEW.sanction_id
      AND recipient.is_active;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sanction_versions_notify
    AFTER INSERT ON sanction_versions
    FOR EACH ROW EXECUTE FUNCTION notify_sanction_version();

-- A team confirms a version once; the first confirmation is the one kept.
CREATE TABLE sanction_acknowledgements (
    id                  bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sanction_version_id bigint      NOT NULL,
    team_id             bigint      NOT NULL,
    acknowledged_by     bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    acknowledged_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT sanction_acknowledgements_unique_per_version UNIQUE (sanction_version_id, team_id),
    CONSTRAINT sanction_acknowledgements_recipient_fkey FOREIGN KEY (sanction_version_id, team_id)
        REFERENCES sanction_recipients (sanction_version_id, team_id) ON DELETE RESTRICT
);

CREATE INDEX sanction_acknowledgements_team_id_idx ON sanction_acknowledgements (team_id);

CREATE TRIGGER sanction_acknowledgements_append_only
    BEFORE UPDATE OR DELETE ON sanction_acknowledgements
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

-- The team is the acknowledging user's, never the caller's word, so nobody can
-- confirm for another team and a FIA administrator (no team) never confirms. A
-- version the team was not sent fails with RC013, and one replaced by a later
-- version with RC014: confirming it would record a decision that no longer
-- stands. The service answers RC013 with a 404 and RC014 with a 409.
CREATE FUNCTION set_sanction_acknowledgement_team() RETURNS trigger AS $$
BEGIN
    NEW.team_id := (SELECT team_id FROM users WHERE id = NEW.acknowledged_by AND is_active);

    IF NOT EXISTS (
        SELECT 1
        FROM sanction_recipients
        WHERE sanction_version_id = NEW.sanction_version_id AND team_id = NEW.team_id
    ) THEN
        RAISE EXCEPTION 'user % was not notified of sanction version %', NEW.acknowledged_by, NEW.sanction_version_id
            USING ERRCODE = 'RC013';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM current_sanctions WHERE version_id = NEW.sanction_version_id) THEN
        RAISE EXCEPTION 'sanction version % was replaced', NEW.sanction_version_id
            USING ERRCODE = 'RC014';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sanction_acknowledgements_set_team
    BEFORE INSERT ON sanction_acknowledgements
    FOR EACH ROW EXECUTE FUNCTION set_sanction_acknowledgement_team();
