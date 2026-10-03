-- name: InsertLoginAttempt :exec
INSERT INTO login_attempts (username, succeeded, ip)
VALUES ($1, $2, $3);

-- name: CountFailuresSinceLastSuccess :one
SELECT count(*)
FROM login_attempts AS failed
WHERE failed.username = sqlc.arg(username)
  AND NOT failed.succeeded
  AND failed.attempted_at > sqlc.arg(since)
  AND failed.attempted_at > COALESCE(
      (SELECT max(ok.attempted_at) FROM login_attempts AS ok WHERE ok.username = failed.username AND ok.succeeded),
      '-infinity'::timestamptz);
