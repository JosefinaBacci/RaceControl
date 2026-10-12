-- 000014_events_simplified: mover circuit/location/country_code/round de events a races

-- 1. Añadir columnas a races SIN NOT NULL (la tabla ya tiene 5 filas)
ALTER TABLE races ADD COLUMN circuit text;
ALTER TABLE races ADD COLUMN location text;
ALTER TABLE races ADD COLUMN country_code text;
ALTER TABLE races ADD COLUMN round integer;

-- 2. Copiar datos de events a races (emparejamos por event_id)
--    Las 5 carreras tendrán los datos; las 3 pruebas sueltas tendrán NULL (aunque races
--    solo tiene 5 filas, las emparejamos por event_id; las que no tienen match se quedan NULL).
UPDATE races r
SET     circuit = e.circuit,
        location = e.location,
        country_code = e.country_code,
        round = e.round
FROM events e
WHERE r.event_id = e.id;

-- 3. Ahora las columnas de events pasan a ser NULLables para poder borrarlas después.
ALTER TABLE events ALTER COLUMN circuit DROP NOT NULL;
ALTER TABLE events ALTER COLUMN location DROP NOT NULL;
ALTER TABLE events ALTER COLUMN country_code DROP NOT NULL;
ALTER TABLE events ALTER COLUMN round DROP NOT NULL;

-- 4. Borrar las columnas de events
ALTER TABLE events DROP COLUMN circuit;
ALTER TABLE events DROP COLUMN location;
ALTER TABLE events DROP COLUMN country_code;
ALTER TABLE events DROP COLUMN round;
