-- Demo conversations (US17), so each side of the messaging shows an unread
-- message. They name the demo accounts, which the seeder creates after demo.sql,
-- so it runs last and inserts nothing when those accounts are absent. Idempotent
-- like the other seeds: the ids are fixed and every insert skips a row that
-- already exists.
--
-- ferrari.admin asks the FIA about the technical control and fia.admin answers,
-- which ferrari.admin has not read yet. fia.admin summons Red Bull, which has not
-- opened it either.

-- 1. conversations
INSERT INTO conversations (id, team_id, subject, created_by, created_at)
OVERRIDING SYSTEM VALUE
SELECT conversation.id, conversation.team_id, conversation.subject, users.id, conversation.created_at
FROM (VALUES
    (1, 1, 'Control técnico de Monza', 'ferrari.admin', TIMESTAMPTZ '2026-10-07 09:00:00+00'),
    (2, 2, 'Citación por el reglamento de neumáticos', 'fia.admin', TIMESTAMPTZ '2026-10-08 15:00:00+00')
) AS conversation (id, team_id, subject, created_by, created_at)
JOIN users ON users.username = conversation.created_by
ON CONFLICT (id) DO NOTHING;

-- The conversations this seed owns: its fixed ids, with the demo team and author.
-- A conversation that already took one of those ids belongs to someone else, and
-- messages are append-only, so everything below skips it instead of grafting the
-- demo thread onto it.
CREATE OR REPLACE TEMPORARY VIEW demo_conversations AS
SELECT conversations.id
FROM conversations
JOIN users ON users.id = conversations.created_by
WHERE (conversations.id, conversations.team_id, users.username) IN ((1, 1, 'ferrari.admin'), (2, 2, 'fia.admin'));

-- 2. messages
INSERT INTO messages (id, conversation_id, author_id, body, created_at)
OVERRIDING SYSTEM VALUE
SELECT message.id, message.conversation_id, users.id, message.body, message.created_at
FROM (VALUES
    (1, 1, 'ferrari.admin', '¿A qué hora tenemos que presentar el auto para el control técnico?', TIMESTAMPTZ '2026-10-07 09:00:00+00'),
    (2, 1, 'fia.admin', 'El jueves a las 10:00 en el box de verificación.', TIMESTAMPTZ '2026-10-07 11:30:00+00'),
    (3, 2, 'fia.admin', 'Presentarse el viernes a las 9:00 con el responsable de neumáticos.', TIMESTAMPTZ '2026-10-08 15:00:00+00')
) AS message (id, conversation_id, author, body, created_at)
JOIN demo_conversations ON demo_conversations.id = message.conversation_id
JOIN users ON users.username = message.author
ON CONFLICT (id) DO NOTHING;

-- 3. conversation_reads: fia.admin read Ferrari's question before answering
INSERT INTO conversation_reads (conversation_id, user_id, last_read_message_id)
SELECT messages.conversation_id, users.id, messages.id
FROM messages
JOIN demo_conversations ON demo_conversations.id = messages.conversation_id
JOIN users ON users.username = 'fia.admin'
WHERE messages.id = 1
ON CONFLICT (conversation_id, user_id) DO NOTHING;

-- 4. setval for every sequence, skipped when the accounts were absent and the
-- table stayed empty: a sequence cannot be set below 1
SELECT setval(pg_get_serial_sequence('conversations', 'id'), max(id)) FROM conversations HAVING max(id) IS NOT NULL;
SELECT setval(pg_get_serial_sequence('messages',      'id'), max(id)) FROM messages      HAVING max(id) IS NOT NULL;

DROP VIEW demo_conversations;
