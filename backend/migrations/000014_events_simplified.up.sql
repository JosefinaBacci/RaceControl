-- 000014_events_simplified: mover circuit/location/country_code/round de events a races,
-- y fijar season/category en la carrera para que una ronda sea única por campeonato.

-- 1. Añadir columnas a races SIN NOT NULL (la tabla ya tiene filas)
ALTER TABLE races ADD COLUMN circuit text;
ALTER TABLE races ADD COLUMN location text;
ALTER TABLE races ADD COLUMN country_code text;
ALTER TABLE races ADD COLUMN round integer;
-- season/category are denormalised from the event: the round is unique per
-- championship, and that key needs the season and category right next to it.
ALTER TABLE races ADD COLUMN season_id bigint;
ALTER TABLE races ADD COLUMN category_id bigint;

-- 2. Copiar datos de events a races (emparejamos por event_id)
UPDATE races r
SET     circuit = e.circuit,
        location = e.location,
        country_code = e.country_code,
        round = e.round,
        season_id = e.season_id,
        category_id = e.category_id
FROM events e
WHERE r.event_id = e.id;

-- 3. Ahora las columnas de events pasan a ser NULLables para poder borrarlas después.
ALTER TABLE events ALTER COLUMN circuit DROP NOT NULL;
ALTER TABLE events ALTER COLUMN location DROP NOT NULL;
ALTER TABLE events ALTER COLUMN country_code DROP NOT NULL;
ALTER TABLE events ALTER COLUMN round DROP NOT NULL;

-- 4. races: obligatorias y con los CHECK que tenían en events.
ALTER TABLE races
    ALTER COLUMN circuit SET NOT NULL,
    ALTER COLUMN location SET NOT NULL,
    ALTER COLUMN country_code SET NOT NULL,
    ALTER COLUMN season_id SET NOT NULL,
    ALTER COLUMN category_id SET NOT NULL;
ALTER TABLE races
    ADD CONSTRAINT races_circuit_not_blank CHECK (btrim(circuit) <> ''),
    ADD CONSTRAINT races_country_code_format CHECK (country_code ~ '^[a-z]{2}$'),
    ADD CONSTRAINT races_round_positive CHECK (round IS NULL OR round > 0);

-- 5. La ronda es única por (temporada, categoría). La FK compuesta contra events
--    mantiene honestas las copias de season/category, así el UNIQUE se sostiene.
ALTER TABLE events ADD CONSTRAINT events_id_season_category_unique UNIQUE (id, season_id, category_id);
ALTER TABLE races ADD CONSTRAINT races_event_season_category_fkey
    FOREIGN KEY (event_id, season_id, category_id)
    REFERENCES events (id, season_id, category_id) ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE races ADD CONSTRAINT races_round_unique_per_season_category
    UNIQUE (season_id, category_id, round);

-- 6. Borrar las columnas de events
ALTER TABLE events DROP COLUMN circuit;
ALTER TABLE events DROP COLUMN location;
ALTER TABLE events DROP COLUMN country_code;
ALTER TABLE events DROP COLUMN round;
