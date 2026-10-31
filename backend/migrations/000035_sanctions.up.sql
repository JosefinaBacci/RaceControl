-- 000035_sanctions: the FIA's official record of sanctions (US15).
--
-- A sanction falls on a driver or on a team, never both: two optional foreign keys
-- and a num_nonnulls check (02-modelo-de-datos.md §2) keep referential integrity
-- on both branches. accountable_team_id is the team the sanction is on: the team
-- itself, or the driver's team on the day of the sanction. It is stored, like
-- score_entries.team_id, because a later assignment may start before an old
-- sanction and deriving it again would rewrite whom it concerned.
--
-- The sanction row is only its identity: on whom, and when it was issued. What it
-- imposes lives in sanction_versions, one row per issue, amendment or revocation,
-- so a correction never overwrites what was notified before and each version is
-- acknowledged on its own (US26). Both tables are append-only, which is why they
-- carry their instant instead of created_at/updated_at, and revoking is the soft
-- delete.

CREATE TABLE sanctions (
    id                  bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    driver_id           bigint      REFERENCES drivers (id) ON DELETE RESTRICT,
    team_id             bigint      REFERENCES teams (id) ON DELETE RESTRICT,
    accountable_team_id bigint      NOT NULL REFERENCES teams (id) ON DELETE RESTRICT,
    issued_at           timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT sanctions_single_subject CHECK (num_nonnulls(driver_id, team_id) = 1),
    CONSTRAINT sanctions_accountable_is_the_team CHECK (team_id IS NULL OR accountable_team_id = team_id)
);

CREATE INDEX sanctions_driver_id_idx ON sanctions (driver_id);
CREATE INDEX sanctions_accountable_team_id_idx ON sanctions (accountable_team_id);
CREATE INDEX sanctions_issued_at_idx ON sanctions (issued_at DESC);

-- The accountable team is never entered by hand. A driver without a team on the
-- day has nobody to notify, so the sanction is refused with RC011.
CREATE FUNCTION set_sanction_accountable_team() RETURNS trigger AS $$
BEGIN
    NEW.accountable_team_id := COALESCE(NEW.team_id, assigned_team(NEW.driver_id, NEW.issued_at));
    IF NEW.accountable_team_id IS NULL THEN
        RAISE EXCEPTION 'driver % has no team on %', NEW.driver_id, NEW.issued_at
            USING ERRCODE = 'RC011';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sanctions_set_accountable_team
    BEFORE INSERT ON sanctions
    FOR EACH ROW EXECUTE FUNCTION set_sanction_accountable_team();

CREATE TRIGGER sanctions_append_only
    BEFORE UPDATE OR DELETE ON sanctions
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

-- An amendment and a revocation explain themselves in the note; the issue does
-- not need one, its reason is the explanation.
CREATE TABLE sanction_versions (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sanction_id     bigint      NOT NULL REFERENCES sanctions (id) ON DELETE RESTRICT,
    action          text        NOT NULL,
    penalty_type_id bigint      NOT NULL REFERENCES penalty_types (id) ON DELETE RESTRICT,
    penalty_value   integer,
    reason          text        NOT NULL,
    event_id        bigint      REFERENCES events (id) ON DELETE RESTRICT,
    note            text,
    actor_user_id   bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    changed_at      timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT sanction_versions_action_valid CHECK (action IN ('issued', 'amended', 'revoked')),
    CONSTRAINT sanction_versions_reason_not_blank CHECK (btrim(reason) <> ''),
    CONSTRAINT sanction_versions_reason_length CHECK (char_length(reason) <= 2000),
    CONSTRAINT sanction_versions_note_only_on_change CHECK ((action = 'issued') = (note IS NULL)),
    CONSTRAINT sanction_versions_note_not_blank CHECK (btrim(note) <> ''),
    CONSTRAINT sanction_versions_note_length CHECK (char_length(note) <= 1000)
);

CREATE INDEX sanction_versions_sanction_idx ON sanction_versions (sanction_id, changed_at DESC, id DESC);

CREATE TRIGGER sanction_versions_append_only
    BEFORE UPDATE OR DELETE ON sanction_versions
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

-- The value has to fit the unit of its type: required and within the ceiling when
-- the type has a unit, absent when it has none. The rule reads another table, so
-- it is a trigger and not a CHECK. RC010 lets the service answer with a 400.
CREATE FUNCTION assert_sanction_penalty_value() RETURNS trigger AS $$
DECLARE
    penalty penalty_types;
BEGIN
    SELECT * INTO penalty FROM penalty_types WHERE id = NEW.penalty_type_id;
    IF penalty.unit IS NULL AND NEW.penalty_value IS NOT NULL
       OR penalty.unit IS NOT NULL
          AND (NEW.penalty_value IS NULL OR NEW.penalty_value NOT BETWEEN 1 AND penalty.max_value) THEN
        RAISE EXCEPTION 'penalty value % does not fit penalty type %', NEW.penalty_value, NEW.penalty_type_id
            USING ERRCODE = 'RC010';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sanction_versions_penalty_value
    BEFORE INSERT ON sanction_versions
    FOR EACH ROW EXECUTE FUNCTION assert_sanction_penalty_value();

-- A sanction's history starts with its issue and ends with its revocation: a
-- second issue, a change before the issue or after the revocation is refused
-- with RC012. The service checks it first under a lock; this is the backstop.
CREATE FUNCTION assert_sanction_version_order() RETURNS trigger AS $$
DECLARE
    latest_action text;
BEGIN
    SELECT action INTO latest_action
    FROM sanction_versions
    WHERE sanction_id = NEW.sanction_id
    ORDER BY changed_at DESC, id DESC
    LIMIT 1;

    IF (NEW.action = 'issued') <> (latest_action IS NULL) OR latest_action = 'revoked' THEN
        RAISE EXCEPTION 'sanction % cannot take an % version after %', NEW.sanction_id, NEW.action, latest_action
            USING ERRCODE = 'RC012';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sanction_versions_order
    BEFORE INSERT ON sanction_versions
    FOR EACH ROW EXECUTE FUNCTION assert_sanction_version_order();

-- The sanction as it stands: its identity plus its latest version. Two versions
-- stamped with the same instant are ordered by id.
CREATE VIEW current_sanctions AS
SELECT DISTINCT ON (s.id)
    s.id,
    s.driver_id,
    s.team_id,
    s.accountable_team_id,
    s.issued_at,
    v.id                 AS version_id,
    v.action,
    v.penalty_type_id,
    v.penalty_value,
    v.reason,
    v.event_id,
    v.note,
    v.changed_at,
    v.action = 'revoked' AS is_revoked
FROM sanctions s
JOIN sanction_versions v ON v.sanction_id = s.id
ORDER BY s.id, v.changed_at DESC, v.id DESC;
