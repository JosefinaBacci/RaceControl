-- 000015_events_stripped: quitar starts_at/ends_at de races y tyre_tests
-- (el timing vive solo en events, igual que en technical_controls)

-- 1. Quitar CHECK y NOT NULL de races primero
ALTER TABLE races DROP CONSTRAINT races_ends_after_start;
ALTER TABLE races ALTER COLUMN ends_at DROP NOT NULL;
ALTER TABLE races ALTER COLUMN starts_at DROP NOT NULL;

-- 2. Borrar columns de races
ALTER TABLE races DROP COLUMN ends_at;
ALTER TABLE races DROP COLUMN starts_at;

-- 3. Igual para tyre_tests: soltar CHECK y NOT NULL
ALTER TABLE tyre_tests DROP CONSTRAINT tyre_tests_ends_after_start;
ALTER TABLE tyre_tests ALTER COLUMN ends_at DROP NOT NULL;
ALTER TABLE tyre_tests ALTER COLUMN starts_at DROP NOT NULL;

-- 4. Borrar columns de tyre_tests
ALTER TABLE tyre_tests DROP COLUMN ends_at;
ALTER TABLE tyre_tests DROP COLUMN starts_at;
