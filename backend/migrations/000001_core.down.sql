DROP TRIGGER IF EXISTS users_set_updated_at ON users;
DROP TRIGGER IF EXISTS teams_set_updated_at ON teams;
DROP TRIGGER IF EXISTS seasons_set_updated_at ON seasons;
DROP TRIGGER IF EXISTS categories_set_updated_at ON categories;

DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS seasons;
DROP TABLE IF EXISTS categories;

DROP FUNCTION IF EXISTS set_updated_at;