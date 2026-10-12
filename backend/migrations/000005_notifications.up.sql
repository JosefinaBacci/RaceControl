-- In-app notifications addressed to a single user. read_at marks the moment the
-- recipient opened it; a NULL means unread.

CREATE TABLE notifications (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    bigint      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind       text        NOT NULL,
    title      text        NOT NULL,
    body       text        NOT NULL,
    read_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT notifications_kind_valid CHECK (
        kind IN ('score_published', 'score_updated', 'sanction', 'announcement', 'message')
    ),
    CONSTRAINT notifications_title_not_blank CHECK (btrim(title) <> '')
);

CREATE INDEX notifications_user_created_at_idx ON notifications (user_id, created_at DESC);
-- Unread lookups (the badge count) only walk rows without read_at.
CREATE INDEX notifications_user_unread_idx ON notifications (user_id) WHERE read_at IS NULL;

CREATE TRIGGER notifications_set_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
