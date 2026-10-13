-- 000013 down: reverts the season_drivers → driver_season_numbers rename
ALTER TABLE season_drivers RENAME TO driver_season_numbers;
