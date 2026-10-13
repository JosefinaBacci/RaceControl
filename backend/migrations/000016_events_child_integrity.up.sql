-- 000016_events_child_integrity: cada evento admite a lo sumo una hija de cada
-- tipo, el nombre vive solo en events y una carrera siempre tiene ronda.

-- 1. El nombre es del evento. La hija lo duplicaba y los dos podían divergir.
ALTER TABLE races DROP COLUMN name;
ALTER TABLE tyre_tests DROP COLUMN name;
ALTER TABLE technical_controls DROP COLUMN name;

-- 2. Una carrera siempre tiene ronda. Con NOT NULL, el UNIQUE por campeonato deja
--    de poder esquivarse cargando la ronda en NULL.
ALTER TABLE races DROP CONSTRAINT races_round_positive;
ALTER TABLE races
    ALTER COLUMN round SET NOT NULL,
    ADD CONSTRAINT races_round_positive CHECK (round > 0);

-- 3. A lo sumo una hija por evento. El índice único reemplaza al plano, cuyas
--    búsquedas por event_id ya cubre el prefijo del único.
DROP INDEX races_event_id_idx;
ALTER TABLE races ADD CONSTRAINT races_event_id_unique UNIQUE (event_id);

DROP INDEX tyre_tests_event_id_idx;
ALTER TABLE tyre_tests ADD CONSTRAINT tyre_tests_event_id_unique UNIQUE (event_id);

DROP INDEX technical_controls_event_id_idx;
ALTER TABLE technical_controls ADD CONSTRAINT technical_controls_event_id_unique UNIQUE (event_id);
