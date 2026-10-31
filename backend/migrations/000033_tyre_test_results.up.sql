-- 000033_tyre_test_results: the official result of a tyre test (US19).
--
-- The test itself is the tyre_test event of the calendar; this table records, per
-- compound it evaluated, the day it ran, the verdict and the report, with the FIA
-- administrator who loaded it. A result is official evidence, so it is append-only:
-- a mistake is fixed by a new row that supersedes the previous one and says why.
--
-- The supersede constraints are the ones every results table uses (technical controls too):
-- the composite foreign key keeps a correction on the same test and compound, a
-- row is superseded at most once and only by a later one, and only one original
-- exists per test and compound. Together they make each compound's results a
-- single linear chain whose head is the current result.

CREATE TABLE tyre_test_results (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id        bigint      NOT NULL REFERENCES tyre_tests (event_id) ON DELETE RESTRICT,
    compound_id     bigint      NOT NULL REFERENCES tyre_compounds (id) ON DELETE RESTRICT,
    conducted_on    date        NOT NULL,
    passed          boolean     NOT NULL,
    notes           text        NOT NULL DEFAULT '',
    supersedes_id   bigint,
    correction_note text,
    recorded_by     bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    recorded_at     timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT tyre_test_results_notes_length CHECK (char_length(notes) <= 2000),
    CONSTRAINT tyre_test_results_failure_explained CHECK (passed OR btrim(notes) <> ''),
    CONSTRAINT tyre_test_results_correction_note_length CHECK (char_length(correction_note) <= 2000),
    CONSTRAINT tyre_test_results_subject_unique UNIQUE (id, event_id, compound_id),
    CONSTRAINT tyre_test_results_supersedes_fkey FOREIGN KEY (supersedes_id, event_id, compound_id)
        REFERENCES tyre_test_results (id, event_id, compound_id) ON DELETE RESTRICT,
    CONSTRAINT tyre_test_results_supersedes_once UNIQUE (supersedes_id),
    CONSTRAINT tyre_test_results_supersedes_earlier CHECK (supersedes_id < id),
    CONSTRAINT tyre_test_results_correction_note CHECK ((supersedes_id IS NULL) = (correction_note IS NULL))
);

CREATE UNIQUE INDEX tyre_test_results_one_original
    ON tyre_test_results (event_id, compound_id) WHERE supersedes_id IS NULL;
CREATE INDEX tyre_test_results_conducted_on_idx ON tyre_test_results (conducted_on);
CREATE INDEX tyre_test_results_compound_id_idx ON tyre_test_results (compound_id);

CREATE TRIGGER tyre_test_results_append_only
    BEFORE UPDATE OR DELETE ON tyre_test_results
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

-- The result as every reader shows it: with its compound, the test's circuit and
-- championship and the author. The current result of each compound is the head of
-- its chain, the row nothing supersedes: public reads keep only those, and the FIA
-- reads the whole chain. One view keeps that join and that rule in a single place.
CREATE VIEW tyre_test_result_entries AS
SELECT
    r.id,
    r.event_id,
    r.conducted_on,
    r.passed,
    r.notes,
    r.supersedes_id,
    r.correction_note,
    NOT EXISTS (SELECT 1 FROM tyre_test_results later WHERE later.supersedes_id = r.id) AS is_current,
    r.recorded_at,
    u.username      AS recorded_by_username,
    tc.code         AS compound_code,
    tc.name         AS compound_name,
    ci.name         AS circuit_name,
    s.year          AS season_year,
    c.code          AS category_code
FROM tyre_test_results r
JOIN tyre_compounds tc ON tc.id = r.compound_id
JOIN tyre_tests t ON t.event_id = r.event_id
JOIN circuits ci ON ci.id = t.circuit_id
JOIN seasons s ON s.id = t.season_id
JOIN categories c ON c.id = t.category_id
JOIN users u ON u.id = r.recorded_by;

-- A test with results is final, like a race with scores: moving or cancelling it
-- would leave its official results pointing at a test that no longer happened then.
CREATE TRIGGER events_keep_tyre_tested_final
    BEFORE UPDATE OF starts_at, ends_at, deactivated_at ON events
    FOR EACH ROW WHEN (OLD.kind = 'tyre_test')
    EXECUTE FUNCTION assert_recorded_event_is_final('tyre_test_results', 'id', 'starts_at', 'ends_at', 'deactivated_at');

CREATE TRIGGER tyre_tests_keep_tested_final
    BEFORE UPDATE OF season_id, category_id, circuit_id ON tyre_tests
    FOR EACH ROW EXECUTE FUNCTION assert_recorded_event_is_final('tyre_test_results', 'event_id', 'season_id', 'category_id', 'circuit_id');
