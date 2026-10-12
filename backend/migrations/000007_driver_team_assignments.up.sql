-- A driver can change teams, and the change itself is history. The periods are
-- implicit: closing one means inserting the next row, so a previous period is
-- never edited and the team derived for a past date can never change.

CREATE TABLE driver_team_assignments (
    id             bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    driver_id      bigint      NOT NULL REFERENCES drivers (id) ON DELETE RESTRICT,
    team_id        bigint      NOT NULL REFERENCES teams (id) ON DELETE RESTRICT,
    effective_from timestamptz NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT driver_team_assignments_from_unique UNIQUE (driver_id, effective_from)
);

CREATE INDEX driver_team_assignments_team_id_idx ON driver_team_assignments (team_id);

-- The team a driver belongs to on a given date: the latest assignment that
-- started on or before it.
CREATE FUNCTION assigned_team(p_driver_id bigint, p_at timestamptz) RETURNS bigint AS $$
    SELECT team_id
    FROM driver_team_assignments
    WHERE driver_id = p_driver_id AND effective_from <= p_at
    ORDER BY effective_from DESC
    LIMIT 1;
$$ LANGUAGE sql STABLE;

-- The history is append-only: a recorded period is never edited nor removed.
CREATE FUNCTION reject_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION '% is append-only', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER driver_team_assignments_append_only
    BEFORE UPDATE OR DELETE ON driver_team_assignments
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();
