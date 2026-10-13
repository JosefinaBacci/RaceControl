-- 000015_events_stripped: drop starts_at/ends_at from races and tyre_tests
-- (the timing lives only in events, as in technical_controls)

-- 1. Drop the CHECK and NOT NULL from races first
ALTER TABLE races DROP CONSTRAINT races_ends_after_start;
ALTER TABLE races ALTER COLUMN ends_at DROP NOT NULL;
ALTER TABLE races ALTER COLUMN starts_at DROP NOT NULL;

-- 2. Drop the races columns
ALTER TABLE races DROP COLUMN ends_at;
ALTER TABLE races DROP COLUMN starts_at;

-- 3. Same for tyre_tests: drop the CHECK and NOT NULL
ALTER TABLE tyre_tests DROP CONSTRAINT tyre_tests_ends_after_start;
ALTER TABLE tyre_tests ALTER COLUMN ends_at DROP NOT NULL;
ALTER TABLE tyre_tests ALTER COLUMN starts_at DROP NOT NULL;

-- 4. Drop the tyre_tests columns
ALTER TABLE tyre_tests DROP COLUMN ends_at;
ALTER TABLE tyre_tests DROP COLUMN starts_at;
