DROP TABLE driver_season_numbers;
DROP FUNCTION set_driver_season_number_category();

ALTER TABLE driver_team_assignments DROP COLUMN is_starter;

ALTER TABLE drivers ADD COLUMN number     integer;
ALTER TABLE drivers ADD COLUMN is_starter boolean NOT NULL DEFAULT true;
ALTER TABLE drivers ADD CONSTRAINT drivers_number_range
    CHECK (number IS NULL OR number BETWEEN 0 AND 999);