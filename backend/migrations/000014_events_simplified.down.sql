-- 000014 down: devuelve circuit/location/country_code/round a events.
-- Non-race events legitimately have no circuit in the current model, so the
-- columns come back nullable and the dropped NOT NULL is not re-imposed: there
-- is no honest value to fill for them.

-- 1. races: soltar la unicidad y la FK compuesta que agregó el up
ALTER TABLE races DROP CONSTRAINT races_round_unique_per_season_category;
ALTER TABLE races DROP CONSTRAINT races_event_season_category_fkey;
ALTER TABLE events DROP CONSTRAINT events_id_season_category_unique;

-- 2. Devuelve las columnas a events
ALTER TABLE events ADD COLUMN circuit text;
ALTER TABLE events ADD COLUMN location text;
ALTER TABLE events ADD COLUMN country_code text;
ALTER TABLE events ADD COLUMN round integer;

-- 3. Copiar de races de vuelta a events
UPDATE events e
SET     circuit = r.circuit,
        location = r.location,
        country_code = r.country_code,
        round = r.round
FROM races r
WHERE e.id = r.event_id;

-- 4. Restaurar los CHECK originales de events (los NULL de los eventos que no
--    son carrera pasan el CHECK, que solo rechaza cadenas vacías).
ALTER TABLE events
    ADD CONSTRAINT events_circuit_not_blank CHECK (btrim(circuit) <> ''),
    ADD CONSTRAINT events_country_code_format CHECK (country_code ~ '^[a-z]{2}$'),
    ADD CONSTRAINT events_round_positive CHECK (round IS NULL OR round > 0),
    ADD CONSTRAINT events_round_unique_per_season_category UNIQUE (season_id, category_id, round);

-- 5. Borrar columnas de races
ALTER TABLE races DROP COLUMN circuit;
ALTER TABLE races DROP COLUMN location;
ALTER TABLE races DROP COLUMN country_code;
ALTER TABLE races DROP COLUMN round;
ALTER TABLE races DROP COLUMN season_id;
ALTER TABLE races DROP COLUMN category_id;
