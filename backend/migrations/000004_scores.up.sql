-- Scores per event. team_id is deliberately denormalised so a standings query
-- does not have to join drivers, and the composite foreign key keeps it
-- consistent with the driver's own team.

CREATE TABLE score_entries (
    id         bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id   bigint        NOT NULL REFERENCES events (id) ON DELETE RESTRICT,
    driver_id  bigint        NOT NULL,
    team_id    bigint        NOT NULL,
    position   integer       NOT NULL,
    points     numeric(6, 2) NOT NULL DEFAULT 0,
    created_at timestamptz   NOT NULL DEFAULT now(),
    updated_at timestamptz   NOT NULL DEFAULT now(),

    CONSTRAINT score_entries_position_positive CHECK (position > 0),
    CONSTRAINT score_entries_points_not_negative CHECK (points >= 0),
    CONSTRAINT score_entries_driver_unique_per_event UNIQUE (event_id, driver_id),
    CONSTRAINT score_entries_position_unique_per_event UNIQUE (event_id, position),
    CONSTRAINT score_entries_driver_team_fk FOREIGN KEY (driver_id, team_id)
        REFERENCES drivers (id, team_id) ON DELETE RESTRICT
);

CREATE INDEX score_entries_driver_id_idx ON score_entries (driver_id);
CREATE INDEX score_entries_team_id_idx ON score_entries (team_id);

CREATE TRIGGER score_entries_set_updated_at BEFORE UPDATE ON score_entries
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- One row per publication or change of an event's scores, so "who changed what
-- and when" survives even after the entries were overwritten.
CREATE TABLE score_change_log (
    id            bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id      bigint      NOT NULL REFERENCES events (id) ON DELETE RESTRICT,
    actor_user_id bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    changed_at    timestamptz NOT NULL DEFAULT now(),
    summary       text        NOT NULL
);

CREATE INDEX score_change_log_event_id_idx ON score_change_log (event_id, changed_at DESC);

-- A team acknowledges once per event that its scores were published.
CREATE TABLE score_acknowledgements (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id        bigint      NOT NULL REFERENCES events (id) ON DELETE RESTRICT,
    team_id         bigint      NOT NULL REFERENCES teams (id) ON DELETE RESTRICT,
    acknowledged_by bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    acknowledged_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT score_acknowledgements_unique_per_team UNIQUE (event_id, team_id)
);

CREATE INDEX score_acknowledgements_team_id_idx ON score_acknowledgements (team_id);
