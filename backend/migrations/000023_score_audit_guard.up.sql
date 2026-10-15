-- 000023_score_audit_guard: a score change has to declare who made it. The audit
-- trail (one score_change_log version plus its per-driver items) is written by the
-- service in the same transaction as the entries; what the schema guarantees is
-- the weaker, cheap invariant: nothing writes score_entries without stating an
-- actor. The actor travels as a transaction-local setting (SET LOCAL
-- racecontrol.actor_id) and a guard fails, with RC005, when it is missing.
--
-- The guard does not assemble the log. Grouping the rows of one edit into a single
-- version would need state smuggled between rows (a local change_log_id or
-- transition tables), more machinery than the audit is worth; 02-modelo-de-datos.md
-- §8 records the decision.

CREATE FUNCTION assert_score_actor_declared() RETURNS trigger AS $$
BEGIN
    IF current_setting('racecontrol.actor_id', true) IS NULL THEN
        RAISE EXCEPTION 'a score change must set racecontrol.actor_id'
            USING ERRCODE = 'RC005';
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER score_entries_require_actor
    BEFORE INSERT OR UPDATE OR DELETE ON score_entries
    FOR EACH ROW EXECUTE FUNCTION assert_score_actor_declared();
