ALTER TABLE conversations ADD COLUMN other_team_id bigint REFERENCES teams(id) ON DELETE RESTRICT;
ALTER TABLE conversations ADD CONSTRAINT conversations_distinct_teams CHECK (other_team_id IS NULL OR other_team_id <> team_id);
CREATE INDEX conversations_other_team_id_idx ON conversations(other_team_id);

CREATE FUNCTION conversation_has_participant(p_user_id bigint, p_conversation_id bigint) RETURNS boolean AS $$
    SELECT EXISTS (
        SELECT 1 FROM conversations c JOIN users u ON u.id = p_user_id
        WHERE c.id = p_conversation_id AND u.is_active
          AND ((c.other_team_id IS NULL AND u.role = 'fia_admin')
            OR (u.role = 'team_admin' AND u.team_id IN (c.team_id, c.other_team_id)))
    );
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION assert_conversation_creator_participates() RETURNS trigger AS $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users u WHERE u.id = NEW.created_by AND u.is_active
        AND ((NEW.other_team_id IS NULL AND u.role = 'fia_admin')
          OR (u.role = 'team_admin' AND u.team_id = NEW.team_id))) THEN
        RAISE EXCEPTION 'creator is not a participant' USING ERRCODE = 'RC007';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION assert_message_author_participates() RETURNS trigger AS $$
BEGIN
    IF NOT conversation_has_participant(NEW.author_id, NEW.conversation_id) THEN
        RAISE EXCEPTION 'author is not a participant' USING ERRCODE = 'RC007';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION assert_conversation_reader_participates() RETURNS trigger AS $$
BEGIN
    IF NOT conversation_has_participant(NEW.user_id, NEW.conversation_id) THEN
        RAISE EXCEPTION 'reader is not a participant' USING ERRCODE = 'RC007';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER conversation_reads_require_participant BEFORE INSERT OR UPDATE ON conversation_reads
    FOR EACH ROW EXECUTE FUNCTION assert_conversation_reader_participates();
CREATE TRIGGER conversations_append_only BEFORE UPDATE OR DELETE ON conversations
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();
