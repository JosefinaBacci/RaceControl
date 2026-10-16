-- 000027_score_actor_guard_code: the audit guard of 000026 gets a code of its own.
--
-- 000026 raised RC005, the code of the acknowledgement guard (000023), which the
-- service answers with a 404. Each guard keeps a distinct code so that a caller
-- that translates one can never mistake the other for it.

CREATE OR REPLACE FUNCTION assert_score_actor_declared() RETURNS trigger AS $$
BEGIN
    IF current_setting('racecontrol.actor_id', true) IS NULL THEN
        RAISE EXCEPTION 'a score change must set racecontrol.actor_id'
            USING ERRCODE = 'RC006';
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
