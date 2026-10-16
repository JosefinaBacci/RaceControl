-- 000023 down: the acknowledgement takes its team from the caller again.
DROP TRIGGER score_acknowledgements_set_team ON score_acknowledgements;
DROP FUNCTION set_score_acknowledgement_team();
