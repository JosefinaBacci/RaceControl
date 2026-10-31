-- 000034_penalty_types: the catalogue of penalties a sanction can carry (US15).
--
-- A penalty is a type of the FIA sporting code plus, for some types, a magnitude:
-- seconds of a time penalty, grid places, licence points, euros of a fine, races
-- of a suspension. The unit and its ceiling are data, so a new type is an INSERT
-- in seeds/reference.sql and no code branches on the type. A type without a unit
-- carries no value (a reprimand, a disqualification).

CREATE TABLE penalty_types (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       text        NOT NULL UNIQUE,
    name       text        NOT NULL,
    unit       text,
    max_value  integer,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT penalty_types_code_format CHECK (code ~ '^[a-z_]{2,32}$'),
    CONSTRAINT penalty_types_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT penalty_types_unit_valid CHECK (
        unit IN ('seconds', 'grid_places', 'license_points', 'euros', 'races')
    ),
    CONSTRAINT penalty_types_value_needs_unit CHECK ((unit IS NULL) = (max_value IS NULL)),
    CONSTRAINT penalty_types_max_value_positive CHECK (max_value > 0)
);

CREATE TRIGGER penalty_types_set_updated_at BEFORE UPDATE ON penalty_types
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
