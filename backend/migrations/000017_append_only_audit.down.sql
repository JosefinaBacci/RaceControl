-- 000017 down: the audit tables become mutable again.
DROP TRIGGER score_change_log_append_only ON score_change_log;
DROP TRIGGER score_change_log_items_append_only ON score_change_log_items;
DROP TRIGGER score_acknowledgements_append_only ON score_acknowledgements;
DROP TRIGGER login_attempts_append_only ON login_attempts;
