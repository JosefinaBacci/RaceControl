-- 000031_technical_control_results: whether each team's car passed a control.
--
-- The statement asks the FIA to record, for every team, whether its car passed
-- each technical control. The control (its days, type and championship) is the
-- calendar event; a result grades one team in it and is an official record:
-- nothing rewrites or erases it (reject_mutation). A typo is fixed by a new row
-- that supersedes the current one with a stated reason, so the chain of rows is
-- the history of the record.
--
-- Exactly one row per team and control is in force, enforced by the schema:
-- * one original row per team and control (partial unique index);
-- * every row is superseded at most once (unique supersedes_id), and only by a
--   later row (supersedes_id < id), so the chain cannot fork or loop;
-- * a correction points at a row of the same control and team (composite FK).
-- The row in force is the one nothing supersedes.
--
-- The team must race in the control's championship. Carrying category_id lets
-- two composite foreign keys say so: one to the control, one to the team.

ALTER TABLE teams
    ADD CONSTRAINT teams_id_category_unique UNIQUE (id, category_id);

ALTER TABLE technical_controls
    ADD CONSTRAINT technical_controls_event_category_unique UNIQUE (event_id, category_id);

CREATE TABLE technical_control_results (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id        bigint      NOT NULL,
    category_id     bigint      NOT NULL,
    team_id         bigint      NOT NULL,
    conducted_on    date        NOT NULL,
    passed          boolean     NOT NULL,
    notes           text        NOT NULL DEFAULT '',
    supersedes_id   bigint,
    correction_note text,
    recorded_by     bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    recorded_at     timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT technical_control_results_control_fkey FOREIGN KEY (event_id, category_id)
        REFERENCES technical_controls (event_id, category_id) ON DELETE RESTRICT,
    CONSTRAINT technical_control_results_team_fkey FOREIGN KEY (team_id, category_id)
        REFERENCES teams (id, category_id) ON DELETE RESTRICT,
    CONSTRAINT technical_control_results_subject_unique UNIQUE (id, event_id, team_id),
    CONSTRAINT technical_control_results_supersedes_fkey FOREIGN KEY (supersedes_id, event_id, team_id)
        REFERENCES technical_control_results (id, event_id, team_id) ON DELETE RESTRICT,
    CONSTRAINT technical_control_results_supersedes_once UNIQUE (supersedes_id),
    CONSTRAINT technical_control_results_supersedes_earlier CHECK (supersedes_id < id),
    CONSTRAINT technical_control_results_correction_note CHECK (
        (supersedes_id IS NULL) = (correction_note IS NULL)
        AND (correction_note IS NULL OR btrim(correction_note) <> '')
    ),
    CONSTRAINT technical_control_results_failed_needs_notes CHECK (passed OR btrim(notes) <> ''),
    CONSTRAINT technical_control_results_notes_length CHECK (
        char_length(notes) <= 2000 AND char_length(coalesce(correction_note, '')) <= 2000
    )
);

CREATE UNIQUE INDEX technical_control_results_one_original
    ON technical_control_results (event_id, team_id) WHERE supersedes_id IS NULL;
CREATE INDEX technical_control_results_conducted_on_idx ON technical_control_results (conducted_on);

CREATE TRIGGER technical_control_results_append_only
    BEFORE UPDATE OR DELETE ON technical_control_results
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

-- A control with results keeps its days, its status and what it certifies.
CREATE TRIGGER events_keep_controlled_event_final
    BEFORE UPDATE OF starts_at, ends_at, deactivated_at ON events
    FOR EACH ROW WHEN (OLD.kind = 'technical_control')
    EXECUTE FUNCTION assert_recorded_event_is_final(
        'technical_control_results', 'id', 'starts_at', 'ends_at', 'deactivated_at');

CREATE TRIGGER technical_controls_keep_controlled_final
    BEFORE UPDATE OF control_type, season_id, category_id ON technical_controls
    FOR EACH ROW
    EXECUTE FUNCTION assert_recorded_event_is_final(
        'technical_control_results', 'event_id', 'control_type', 'season_id', 'category_id');

-- Every row with what a reader shows next to it. The FIA reads the whole chain;
-- everyone else reads only the rows in force.
CREATE VIEW technical_control_revisions AS
SELECT
    r.id,
    r.event_id,
    r.team_id,
    t.name          AS team_name,
    r.conducted_on,
    r.passed,
    r.notes,
    r.supersedes_id,
    r.correction_note,
    r.recorded_at,
    u.username      AS recorded_by,
    tc.control_type,
    s.year          AS season_year,
    c.code          AS category_code
FROM technical_control_results r
JOIN technical_controls tc ON tc.event_id = r.event_id
JOIN teams t ON t.id = r.team_id
JOIN users u ON u.id = r.recorded_by
JOIN seasons s ON s.id = tc.season_id
JOIN categories c ON c.id = tc.category_id;

CREATE VIEW technical_control_current_results AS
SELECT revision.*
FROM technical_control_revisions revision
WHERE NOT EXISTS (
    SELECT 1 FROM technical_control_results newer WHERE newer.supersedes_id = revision.id
);
