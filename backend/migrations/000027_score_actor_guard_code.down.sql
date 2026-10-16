-- 000027 down: the audit guard raises RC005 again, as 000026 left it.
CREATE OR REPLACE FUNCTION assert_score_actor_declared() RETURNS trigger AS $$
BEGIN
    IF current_setting('racecontrol.actor_id', true) IS NULL THEN
        RAISE EXCEPTION 'a score change must set racecontrol.actor_id'
            USING ERRCODE = 'RC005';
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
