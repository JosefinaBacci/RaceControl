-- 000030_recorded_event_freeze: an event with official results is final.
--
-- A technical control or a tyre test is dated by its event, and each result
-- records the day it was obtained inside the event's days. Moving or cancelling
-- the event afterwards would leave those results outside it, and retyping a
-- control would change what its results certify. Races already have this rule
-- (RC001, 000016); the kinds with results tables share one guard instead of a
-- copy per kind.
--
-- Arguments: the results table, the column of the guarded row that holds the
-- event id (id on events, event_id on a child), then the guarded columns. The
-- guard fires only on a real change, so re-sending the same values passes. The
-- results table is a trigger argument, so it is reached through dynamic SQL;
-- %I quotes it as an identifier and the event id travels as a parameter.
CREATE FUNCTION assert_recorded_event_is_final() RETURNS trigger AS $$
DECLARE
    results_table text := TG_ARGV[0];
    event_id      bigint := (to_jsonb(OLD) ->> TG_ARGV[1])::bigint;
    column_name   text;
    has_changed   boolean := false;
    has_results   boolean;
BEGIN
    FOREACH column_name IN ARRAY TG_ARGV[2:] LOOP
        has_changed := has_changed
            OR (to_jsonb(OLD) -> column_name) IS DISTINCT FROM (to_jsonb(NEW) -> column_name);
    END LOOP;
    IF NOT has_changed THEN
        RETURN NEW;
    END IF;

    EXECUTE format('SELECT EXISTS (SELECT 1 FROM %I WHERE event_id = $1)', results_table)
        INTO has_results USING event_id;
    IF has_results THEN
        RAISE EXCEPTION 'event % has recorded results: it is final', event_id
            USING ERRCODE = 'RC020';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
