-- Tyre test results (US19, US20).

-- name: ListTyreCompounds :many
SELECT id, code, name, kind
FROM tyre_compounds
ORDER BY id;

-- name: GetTyreCompoundByCode :one
SELECT id, code, name, kind
FROM tyre_compounds
WHERE code = $1;

-- name: GetTyreTestWindow :one
SELECT e.starts_at, e.ends_at, e.is_active
FROM events e
JOIN tyre_tests t ON t.event_id = e.id
WHERE e.id = $1;

-- name: GetTyreTestResultSubject :one
SELECT event_id, compound_id
FROM tyre_test_results
WHERE id = $1;

-- name: InsertTyreTestResult :one
INSERT INTO tyre_test_results (
    event_id, compound_id, conducted_on, passed, notes, supersedes_id, correction_note, recorded_by
) VALUES (
    $1, $2, $3, $4, $5, sqlc.narg(supersedes_id), sqlc.narg(correction_note), $6
)
RETURNING id;

-- name: GetTyreTestResultEntry :one
SELECT *
FROM tyre_test_result_entries
WHERE id = $1;

-- name: ListCurrentTyreTestResults :many
SELECT *
FROM tyre_test_result_entries
WHERE is_current
  AND (sqlc.narg(from_day)::date IS NULL OR conducted_on >= sqlc.narg(from_day)::date)
  AND (sqlc.narg(until_day)::date IS NULL OR conducted_on < sqlc.narg(until_day)::date)
  AND (sqlc.narg(compound_code)::text IS NULL OR compound_code = sqlc.narg(compound_code)::text)
ORDER BY conducted_on DESC, id DESC;

-- name: ListTyreTestResultChain :many
SELECT *
FROM tyre_test_result_entries
WHERE event_id = $1
ORDER BY compound_code, id;
