-- The position is derived from the points by the standings query, so storing it
-- only froze an order at insert time and ruled out ties. Points are the data.
ALTER TABLE score_entries DROP CONSTRAINT score_entries_position_unique_per_event;
ALTER TABLE score_entries DROP COLUMN position;

-- Structured audit of score changes: one row per driver whose points changed,
-- so the log answers "what changed" instead of a free-text summary. The kind of
-- change and the position are both derived from the old and new values.
CREATE TABLE score_change_log_items (
    id            bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    change_log_id bigint        NOT NULL REFERENCES score_change_log (id) ON DELETE CASCADE,
    driver_id     bigint        NOT NULL REFERENCES drivers (id) ON DELETE RESTRICT,
    old_points    numeric(6, 2),
    new_points    numeric(6, 2),

    -- A row that changed nothing carries no information, and two NULLs are not
    -- a change either.
    CONSTRAINT score_change_log_items_changed CHECK (old_points IS DISTINCT FROM new_points)
);

CREATE INDEX score_change_log_items_change_log_id_idx ON score_change_log_items (change_log_id);
CREATE INDEX score_change_log_items_driver_id_idx ON score_change_log_items (driver_id);
