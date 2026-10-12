DROP INDEX notifications_event_id_idx;

ALTER TABLE notifications
    DROP COLUMN event_id,
    ADD COLUMN kind  text NOT NULL,
    ADD COLUMN title text NOT NULL,
    ADD COLUMN body  text NOT NULL;

ALTER TABLE notifications
    ADD CONSTRAINT notifications_kind_valid CHECK (
        kind IN ('score_published', 'score_updated', 'sanction', 'announcement', 'message')
    ),
    ADD CONSTRAINT notifications_title_not_blank CHECK (btrim(title) <> '');