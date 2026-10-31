-- 000032_tyre_compounds: the catalogue of compounds a tyre test evaluates.
-- The "type" of a tyre test (US19) is the compound it puts on track, so it is data
-- the filters and the form read, not a list repeated in code. The rows are
-- reference data and live in seeds/reference.sql.

CREATE TABLE tyre_compounds (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       text        NOT NULL UNIQUE,
    name       text        NOT NULL,
    kind       text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT tyre_compounds_code_format CHECK (code ~ '^[a-z0-9_]{2,32}$'),
    CONSTRAINT tyre_compounds_kind_check CHECK (kind IN ('slick', 'intermediate', 'wet'))
);

CREATE TRIGGER tyre_compounds_set_updated_at BEFORE UPDATE ON tyre_compounds
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
