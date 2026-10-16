-- 000028 down: messaging goes away whole; nothing else referenced it.
DROP TABLE conversation_reads;
DROP TABLE messages;
DROP TABLE conversations;
DROP FUNCTION assert_message_author_participates();
DROP FUNCTION assert_conversation_creator_participates();
DROP FUNCTION is_conversation_participant(bigint, bigint);
