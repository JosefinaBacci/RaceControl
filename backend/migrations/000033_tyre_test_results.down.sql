-- 000033 down: tyre test results go away with their view and guards.
DROP TRIGGER tyre_tests_keep_tested_final ON tyre_tests;
DROP TRIGGER events_keep_tyre_tested_final ON events;
DROP VIEW tyre_test_result_entries;
DROP TABLE tyre_test_results;
