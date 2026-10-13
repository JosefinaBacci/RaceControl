-- 000016_events_child_integrity: un evento declara su tipo (kind) y tiene
-- exactamente una hija de ese tipo. Una carrera siempre tiene ronda y el nombre
-- del evento no se duplica en las hijas.

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

-- 4. El tipo del evento. Sin él nada impide que de un evento cuelguen hijas de dos
--    tipos distintos. Se deriva de la hija que ya existía, que era la única fuente
--    del tipo antes de esta migración.
ALTER TABLE events ADD COLUMN kind text;
UPDATE events e SET kind = CASE
        WHEN EXISTS (SELECT 1 FROM races r WHERE r.event_id = e.id)              THEN 'race'
        WHEN EXISTS (SELECT 1 FROM tyre_tests t WHERE t.event_id = e.id)         THEN 'tyre_test'
        WHEN EXISTS (SELECT 1 FROM technical_controls c WHERE c.event_id = e.id) THEN 'technical_control'
    END;
ALTER TABLE events
    ALTER COLUMN kind SET NOT NULL,
    ADD CONSTRAINT events_kind_valid
        CHECK (kind IN ('race', 'tyre_test', 'technical_control')),
    ADD CONSTRAINT events_id_kind_unique UNIQUE (id, kind);

-- 5. Cada hija declara el tipo del evento y lo fija en el suyo, así su FK de
--    identidad no puede apuntar a un evento de otro tipo: de un evento de carreras
--    cuelga una carrera y de uno de controles, un control técnico. La columna es
--    constante, por eso el DEFAULT alcanza y no hay que escribirla en cada INSERT.
ALTER TABLE races
    ADD COLUMN event_kind text NOT NULL DEFAULT 'race',
    ADD CONSTRAINT races_event_kind_check CHECK (event_kind = 'race'),
    DROP CONSTRAINT races_event_id_fkey,
    ADD CONSTRAINT races_event_kind_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE tyre_tests
    ADD COLUMN event_kind text NOT NULL DEFAULT 'tyre_test',
    ADD CONSTRAINT tyre_tests_event_kind_check CHECK (event_kind = 'tyre_test'),
    DROP CONSTRAINT tyre_tests_event_id_fkey,
    ADD CONSTRAINT tyre_tests_event_kind_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE technical_controls
    ADD COLUMN event_kind text NOT NULL DEFAULT 'technical_control',
    ADD CONSTRAINT technical_controls_event_kind_check CHECK (event_kind = 'technical_control'),
    DROP CONSTRAINT technical_controls_event_id_fkey,
    ADD CONSTRAINT technical_controls_event_kind_fkey
        FOREIGN KEY (event_id, event_kind) REFERENCES events (id, kind)
        ON DELETE RESTRICT ON UPDATE CASCADE;

-- 6. Exactamente una hija. Un evento sin hija no es de ningún tipo: su nombre y su
--    tipo no se pueden calcular. Las dos mitades del invariante se chequean recién
--    al commit (DEFERRABLE) para que insertar el evento y su hija en una misma
--    transacción siga siendo posible.
CREATE FUNCTION assert_event_has_child() RETURNS trigger AS $$
DECLARE
    has_child boolean;
BEGIN
    has_child := CASE NEW.kind
        WHEN 'race'              THEN EXISTS (SELECT 1 FROM races              WHERE event_id = NEW.id)
        WHEN 'tyre_test'         THEN EXISTS (SELECT 1 FROM tyre_tests         WHERE event_id = NEW.id)
        WHEN 'technical_control' THEN EXISTS (SELECT 1 FROM technical_controls WHERE event_id = NEW.id)
    END;

    IF NOT has_child THEN
        RAISE EXCEPTION 'event % of kind % has no child row', NEW.id, NEW.kind;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER events_require_child
    AFTER INSERT OR UPDATE OF kind ON events
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION assert_event_has_child();

-- Borrar solo la hija dejaría al evento huérfano y el trigger de events no se
-- enteraría (nadie tocó el evento). Se ignora el caso "el padre también se borra".
CREATE FUNCTION assert_event_keeps_child() RETURNS trigger AS $$
DECLARE
    has_child boolean;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM events WHERE id = OLD.event_id) THEN
        RETURN NULL;
    END IF;

    EXECUTE format('SELECT EXISTS (SELECT 1 FROM %I WHERE event_id = $1)', TG_TABLE_NAME)
        INTO has_child USING OLD.event_id;

    IF NOT has_child THEN
        RAISE EXCEPTION 'event % would be left without a child in %', OLD.event_id, TG_TABLE_NAME;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER races_keep_event_child
    AFTER DELETE OR UPDATE OF event_id ON races
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION assert_event_keeps_child();

CREATE CONSTRAINT TRIGGER tyre_tests_keep_event_child
    AFTER DELETE OR UPDATE OF event_id ON tyre_tests
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION assert_event_keeps_child();

CREATE CONSTRAINT TRIGGER technical_controls_keep_event_child
    AFTER DELETE OR UPDATE OF event_id ON technical_controls
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION assert_event_keeps_child();
