-- 000023 down: a score change no longer has to declare an actor.
DROP TRIGGER score_entries_require_actor ON score_entries;
DROP FUNCTION assert_score_actor_declared();
