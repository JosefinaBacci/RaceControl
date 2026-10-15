-- 000020_drop_child_guard: an event is never left without a child by the
-- application. A child is the event's identity, so repointing it is not an
-- operation, and events leave the calendar through deactivated_at, never by
-- DELETE. The mirror guard only covered the half-delete (the "parent goes too"
-- case was ignored), so it enforced a rule the code never breaks. The "at least
-- one child" invariant stays on the events side (assert_event_has_child, 000016).
--
-- The guard cost three deferred constraint triggers plus a function that reaches
-- for a different table through dynamic SQL, the same unfair price the calendar
-- refused to pay for a cancelled round. Hard deletes remain a raw-SQL-only
-- concern; if one is ever wanted, reject_mutation on the child is the simpler net.

DROP TRIGGER races_keep_event_child ON races;
DROP TRIGGER tyre_tests_keep_event_child ON tyre_tests;
DROP TRIGGER technical_controls_keep_event_child ON technical_controls;
DROP FUNCTION assert_event_keeps_child();
