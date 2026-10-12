-- 000015 down: revierte la quita de starts_at/ends_at

-- 1. Añade columns de nuevo a tyre_tests
ALTER TABLE tyre_tests ADD COLUMN starts_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE tyre_tests ADD COLUMN ends_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE tyre_tests ADD CONSTRAINT tyre_tests_ends_after_start CHECK (ends_at > starts_at);

-- 2. Añade columns de nuevo a races
ALTER TABLE races ADD COLUMN starts_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE races ADD COLUMN ends_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE races ADD CONSTRAINT races_ends_after_start CHECK (ends_at > starts_at);
