-- Core reference data, identity and session tables.
-- Roles and states are CHECK constraints rather than native enums so that adding a
-- value is an additive migration instead of a table rewrite.

CREATE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE categories (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       text        NOT NULL UNIQUE,
    name       text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT categories_code_format CHECK (code ~ '^[a-z0-9_]{2,32}$')
);

CREATE TRIGGER categories_set_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE seasons (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    year       integer     NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT seasons_year_range CHECK (year BETWEEN 1950 AND 2100)
);

CREATE TRIGGER seasons_set_updated_at BEFORE UPDATE ON seasons
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- The unique (category_id, code) index also serves lookups and FK checks by category,
-- so category_id needs no index of its own.
CREATE TABLE teams (
    id          bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id bigint      NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    code        text        NOT NULL,
    name        text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT teams_code_unique_per_category UNIQUE (category_id, code),
    CONSTRAINT teams_code_format CHECK (code ~ '^[a-z0-9_]{2,16}$')
);

CREATE TRIGGER teams_set_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- is_active is derived from deactivated_at so the two can never disagree.
CREATE TABLE users (
    id             bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username       text        NOT NULL UNIQUE,
    email          text        UNIQUE,
    password_hash  text        NOT NULL,
    role           text        NOT NULL,
    team_id        bigint      REFERENCES teams (id) ON DELETE RESTRICT,
    is_active      boolean     NOT NULL GENERATED ALWAYS AS (deactivated_at IS NULL) STORED,
    deactivated_at timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT users_role_valid CHECK (role IN ('fia_admin', 'team_admin')),
    CONSTRAINT users_username_format CHECK (username ~ '^[a-z0-9._-]{3,32}$'),
    CONSTRAINT users_email_lowercase CHECK (email IS NULL OR email = lower(email)),
    CONSTRAINT users_team_admin_has_team CHECK ((role = 'team_admin') = (team_id IS NOT NULL))
);

CREATE INDEX users_team_id_idx ON users (team_id);

CREATE TRIGGER users_set_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE sessions (
    id                  bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id             bigint      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash          text        NOT NULL UNIQUE,
    created_at          timestamptz NOT NULL DEFAULT now(),
    last_used_at        timestamptz NOT NULL DEFAULT now(),
    absolute_expires_at timestamptz NOT NULL,
    revoked_at          timestamptz,
    ip                  inet,
    user_agent          text,

    CONSTRAINT sessions_expiry_after_creation CHECK (absolute_expires_at > created_at),
    CONSTRAINT sessions_revocation_after_creation CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);

CREATE TABLE login_attempts (
    id           bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username     text        NOT NULL,
    succeeded    boolean     NOT NULL,
    ip           inet,
    attempted_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX login_attempts_username_time_idx ON login_attempts (username, attempted_at DESC);
