-- 000025_score_notices: what each team is asked to confirm.
--
-- One row per score change and per team the change touched. Whether the change
-- published the race or corrected it, and whether the team already confirmed it,
-- are derived from the change log and the acknowledgements, so there is no
-- notification row to write or keep in sync: a change recorded by any writer
-- becomes a notice by itself.
--
-- The first change of a race is its publication and any later one corrects it.
-- Two changes stamped with the same instant are ordered by id, so exactly one of
-- them is the publication.

CREATE VIEW score_notices AS
SELECT
    cl.id             AS change_log_id,
    cl.event_id,
    cl.changed_at,
    cl.actor_user_id,
    affected.team_id,
    NOT EXISTS (
        SELECT 1
        FROM score_change_log earlier
        WHERE earlier.event_id = cl.event_id
          AND (earlier.changed_at, earlier.id) < (cl.changed_at, cl.id)
    )                 AS is_publication,
    ack.acknowledged_at,
    ack.acknowledged_by
FROM score_change_log cl
JOIN (
    SELECT DISTINCT change_log_id, team_id
    FROM score_change_log_items
) affected ON affected.change_log_id = cl.id
LEFT JOIN score_acknowledgements ack
    ON ack.change_log_id = cl.id
   AND ack.team_id = affected.team_id;
