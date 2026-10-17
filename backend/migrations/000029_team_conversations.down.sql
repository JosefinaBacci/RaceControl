DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM conversations WHERE other_team_id IS NOT NULL) THEN
        RAISE EXCEPTION 'cannot downgrade while team-to-team conversations exist';
    END IF;
END $$;
DROP TRIGGER conversations_append_only ON conversations;
DROP TRIGGER conversation_reads_require_participant ON conversation_reads;
DROP FUNCTION assert_conversation_reader_participates();
CREATE OR REPLACE FUNCTION assert_message_author_participates() RETURNS trigger AS $$
BEGIN
    IF NOT is_conversation_participant(NEW.author_id, (SELECT team_id FROM conversations WHERE id = NEW.conversation_id)) THEN
        RAISE EXCEPTION 'author is not a participant' USING ERRCODE = 'RC007';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION assert_conversation_creator_participates() RETURNS trigger AS $$
BEGIN
    IF NOT is_conversation_participant(NEW.created_by, NEW.team_id) THEN
        RAISE EXCEPTION 'creator is not a participant' USING ERRCODE = 'RC007';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP FUNCTION conversation_has_participant(bigint, bigint);
ALTER TABLE conversations DROP COLUMN other_team_id;
