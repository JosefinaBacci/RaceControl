-- 000037_sanction_read_models: the sanctions as the API reads them.
--
-- sanction_version_details is every version with what the list, the detail and
-- the notices show next to it: the penalty, the sanctioned driver or team, the
-- championship and the event. The join lives here once instead of in every
-- query. is_current tells the version that stands from the ones a later
-- amendment or revocation replaced.
--
-- sanction_notices is one row per version and recipient team, with the
-- acknowledgement when there is one, like score_notices (000025).

CREATE VIEW sanction_version_details AS
SELECT
    s.id                           AS sanction_id,
    s.driver_id,
    s.team_id,
    s.accountable_team_id,
    s.issued_at,
    v.id                           AS version_id,
    v.action,
    v.penalty_value,
    v.reason,
    v.note,
    v.changed_at,
    v.actor_user_id,
    (current.version_id IS NOT NULL)::boolean AS is_current,
    pt.code                        AS penalty_code,
    pt.name                        AS penalty_name,
    pt.unit                        AS penalty_unit,
    d.first_name                   AS driver_first_name,
    d.last_name                    AS driver_last_name,
    t.code                         AS accountable_team_code,
    t.name                         AS accountable_team_name,
    c.code                         AS category_code,
    c.name                         AS category_name,
    v.event_id,
    ce.kind                        AS event_kind,
    ce.gp_name                     AS event_gp_name,
    ce.round_number                AS event_round_number,
    ce.control_type                AS event_control_type,
    ce.starts_at                   AS event_starts_at
FROM sanction_versions v
JOIN sanctions s ON s.id = v.sanction_id
JOIN penalty_types pt ON pt.id = v.penalty_type_id
JOIN teams t ON t.id = s.accountable_team_id
JOIN categories c ON c.id = t.category_id
LEFT JOIN drivers d ON d.id = s.driver_id
LEFT JOIN calendar_events ce ON ce.id = v.event_id
LEFT JOIN current_sanctions current ON current.version_id = v.id;

CREATE VIEW sanction_notices AS
SELECT
    recipient.sanction_version_id  AS version_id,
    recipient.team_id,
    (current.version_id IS NOT NULL)::boolean AS is_current,
    ack.acknowledged_at,
    ack.acknowledged_by
FROM sanction_recipients recipient
LEFT JOIN current_sanctions current ON current.version_id = recipient.sanction_version_id
LEFT JOIN sanction_acknowledgements ack
    ON ack.sanction_version_id = recipient.sanction_version_id
   AND ack.team_id = recipient.team_id;
