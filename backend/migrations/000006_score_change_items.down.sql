DROP TABLE IF EXISTS score_change_log_items;

-- Recreate the position from the points, which is where the removed code used to
-- read it from, so the table is back in its 000004 shape.
ALTER TABLE score_entries ADD COLUMN position integer;

UPDATE score_entries AS entry
SET position = ranked.position
FROM (
    SELECT id, row_number() OVER (PARTITION BY event_id ORDER BY points DESC, id) AS position
    FROM score_entries
) AS ranked
WHERE ranked.id = entry.id;

ALTER TABLE score_entries ALTER COLUMN position SET NOT NULL;
ALTER TABLE score_entries ADD CONSTRAINT score_entries_position_positive CHECK (position > 0);
ALTER TABLE score_entries ADD CONSTRAINT score_entries_position_unique_per_event UNIQUE (event_id, position);
