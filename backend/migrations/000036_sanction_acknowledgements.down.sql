-- 000036 down: the acknowledgements and the recipients they point at go away.
DROP TABLE sanction_acknowledgements;
DROP FUNCTION set_sanction_acknowledgement_team();
DROP TRIGGER sanction_versions_notify ON sanction_versions;
DROP FUNCTION notify_sanction_version();
DROP TABLE sanction_recipients;
