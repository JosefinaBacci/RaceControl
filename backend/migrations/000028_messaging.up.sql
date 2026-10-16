-- 000028_messaging: secure internal messaging between the FIA and the teams (US17).
--
-- A conversation is a thread between one team and the FIA. Its participants are
-- not stored: they are every active administrator of that team plus every active
-- FIA administrator, so the thread is the institutional channel between both
-- parties and a team administrator who leaves does not orphan it.
--
-- Whether a message is unread is derived as well: it is unread for a participant
-- when it is newer than their read cursor and someone else wrote it. Nothing is
-- written into notifications.

-- is_conversation_participant is the one definition of who belongs to a team's
-- conversations, shared by the guards below.
CREATE FUNCTION is_conversation_participant(user_id bigint, team_id bigint) RETURNS boolean AS $$
    SELECT EXISTS (
        SELECT 1
        FROM users
        WHERE users.id = $1
          AND users.is_active
          AND (users.role = 'fia_admin' OR users.team_id = $2)
    );
$$ LANGUAGE sql STABLE;

CREATE TABLE conversations (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_id    bigint      NOT NULL REFERENCES teams (id) ON DELETE RESTRICT,
    subject    text        NOT NULL,
    created_by bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT conversations_subject_not_blank CHECK (btrim(subject) <> ''),
    CONSTRAINT conversations_subject_length CHECK (char_length(subject) <= 120)
);

CREATE INDEX conversations_team_id_idx ON conversations (team_id);

CREATE TRIGGER conversations_set_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- A message is a fact of the past like the audit trail (000017): nobody edits or
-- erases what was said. The unique (conversation_id, id) lists a thread in order
-- and lets a read cursor point only at a message of its own conversation.
CREATE TABLE messages (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id bigint      NOT NULL REFERENCES conversations (id) ON DELETE RESTRICT,
    author_id       bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    body            text        NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT messages_conversation_order UNIQUE (conversation_id, id),
    CONSTRAINT messages_body_not_blank CHECK (btrim(body) <> ''),
    CONSTRAINT messages_body_length CHECK (char_length(body) <= 4000)
);

CREATE INDEX messages_author_id_idx ON messages (author_id, conversation_id);

CREATE TRIGGER messages_append_only
    BEFORE UPDATE OR DELETE ON messages
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

-- The cursor is the last message the user saw in the conversation. Message ids
-- order a thread, so everything after the cursor is new to them.
CREATE TABLE conversation_reads (
    conversation_id      bigint      NOT NULL REFERENCES conversations (id) ON DELETE RESTRICT,
    user_id              bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    last_read_message_id bigint      NOT NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),

    PRIMARY KEY (conversation_id, user_id),
    CONSTRAINT conversation_reads_message_fk FOREIGN KEY (conversation_id, last_read_message_id)
        REFERENCES messages (conversation_id, id) ON DELETE RESTRICT
);

CREATE TRIGGER conversation_reads_set_updated_at BEFORE UPDATE ON conversation_reads
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Only a participant can open a conversation or write in one, whatever the
-- writer sends: a team administrator cannot reach another team's thread and no
-- inactive account writes. The service checks it first and answers 404; RC007
-- is the same rule kept by the schema, and the service answers it the same way.
CREATE FUNCTION assert_conversation_creator_participates() RETURNS trigger AS $$
BEGIN
    IF NOT is_conversation_participant(NEW.created_by, NEW.team_id) THEN
        RAISE EXCEPTION 'user % cannot open a conversation with team %', NEW.created_by, NEW.team_id
            USING ERRCODE = 'RC007';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER conversations_require_participant
    BEFORE INSERT ON conversations
    FOR EACH ROW EXECUTE FUNCTION assert_conversation_creator_participates();

CREATE FUNCTION assert_message_author_participates() RETURNS trigger AS $$
BEGIN
    IF NOT is_conversation_participant(
        NEW.author_id,
        (SELECT team_id FROM conversations WHERE id = NEW.conversation_id)
    ) THEN
        RAISE EXCEPTION 'user % cannot write in conversation %', NEW.author_id, NEW.conversation_id
            USING ERRCODE = 'RC007';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER messages_require_participant
    BEFORE INSERT ON messages
    FOR EACH ROW EXECUTE FUNCTION assert_message_author_participates();
