-- Notifications are pure facts: who, which event, and whether they were read.
-- The event type lives in events.event_type_id and the publication history in
-- score_change_log, so the message text, the kind and anything else shown to the
-- user is derived at read time. Messaging (FIA <-> team conversations) is a
-- separate concern and gets its own table, so 'message' is not a kind here.

ALTER TABLE notifications
    DROP COLUMN title,
    DROP COLUMN body,
    DROP COLUMN kind,
    ADD COLUMN event_id bigint NOT NULL REFERENCES events (id) ON DELETE RESTRICT;

-- Every notification points at an event (any of its types: race, tyre test or
-- technical control), so a plain index serves the "notifications of this event"
-- lookups.
CREATE INDEX notifications_event_id_idx ON notifications (event_id);