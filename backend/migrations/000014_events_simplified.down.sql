-- 000014 down: revuelve la simplificación de events

-- 1. Devuelve las columnas a events
ALTER TABLE events ADD COLUMN circuit text NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN location text NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN country_code text NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN round integer NOT NULL DEFAULT 0;

-- 2. Copiar de races de vuelta a events (solo las 5 que tienen round poblado)
UPDATE events e
SET     circuit = r.circuit,
        location = r.location,
        country_code = r.country_code,
        round = r.round
FROM races r
WHERE e.id = r.event_id AND r.round IS NOT NULL;

-- 3. Borrar columnas de races
ALTER TABLE races DROP COLUMN circuit;
ALTER TABLE races DROP COLUMN location;
ALTER TABLE races DROP COLUMN country_code;
ALTER TABLE races DROP COLUMN round;
