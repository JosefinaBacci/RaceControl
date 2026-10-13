-- 000017_append_only_audit: an audit trail records what happened, it is not state.
-- A score change, its per-driver detail, an acknowledgement and a login attempt are
-- facts of the past, so rewriting or erasing them would leave a hole in the history.
-- The guard that already protects the assignment history covers them all now.

CREATE TRIGGER score_change_log_append_only
    BEFORE UPDATE OR DELETE ON score_change_log
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

CREATE TRIGGER score_change_log_items_append_only
    BEFORE UPDATE OR DELETE ON score_change_log_items
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

CREATE TRIGGER score_acknowledgements_append_only
    BEFORE UPDATE OR DELETE ON score_acknowledgements
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

-- The lockout counter reads the recent failures and never rewrites them: an attempt
-- is evidence, and letting the app clear it would hand out an unlimited retry.
CREATE TRIGGER login_attempts_append_only
    BEFORE UPDATE OR DELETE ON login_attempts
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();
