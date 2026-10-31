-- 000035 down: the sanctions record goes away with its guards.
DROP VIEW current_sanctions;
DROP TABLE sanction_versions;
DROP FUNCTION assert_sanction_version_order();
DROP FUNCTION assert_sanction_penalty_value();
DROP TABLE sanctions;
DROP FUNCTION set_sanction_accountable_team();
