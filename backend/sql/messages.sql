-- name: CreateConversation :one
INSERT INTO conversations (team_id, other_team_id, subject, created_by)
VALUES (sqlc.arg(team_id), sqlc.narg(other_team_id), sqlc.arg(subject), sqlc.arg(created_by))
RETURNING id;

-- Locking the conversation before the id is drawn makes a thread's ids follow
-- the commit order, so a cursor never skips a message that commits late.
-- name: AddMessage :one
WITH conversation AS (
    SELECT conversations.id FROM conversations WHERE conversations.id = sqlc.arg(conversation_id) FOR UPDATE
)
INSERT INTO messages (conversation_id, author_id, body)
SELECT conversation.id, sqlc.arg(author_id)::bigint, sqlc.arg(body)::text
FROM conversation
RETURNING id, created_at;

-- name: GetConversation :one
SELECT c.id, c.subject, c.created_at,
       t.id AS team_id, t.code AS team_code, t.name AS team_name,
       c.other_team_id, COALESCE(other.name, 'FIA')::text AS other_team_name
FROM conversations c
JOIN teams t ON t.id = c.team_id
LEFT JOIN teams other ON other.id = c.other_team_id
WHERE c.id = sqlc.arg(conversation_id)
  AND conversation_has_participant(sqlc.arg(reader_id), c.id);

-- name: ListConversationMessages :many
SELECT m.id, m.body, m.created_at, m.author_id,
       author.username AS author_username, author.role AS author_role,
       (m.author_id <> sqlc.arg(reader_id) AND m.id > COALESCE(r.last_read_message_id, 0))::boolean AS is_unread,
       EXISTS (
         SELECT 1 FROM conversation_reads receipt JOIN users recipient ON recipient.id = receipt.user_id
         WHERE receipt.conversation_id = m.conversation_id AND receipt.last_read_message_id >= m.id
           AND conversation_has_participant(recipient.id, m.conversation_id)
           AND (recipient.role <> author.role OR recipient.team_id IS DISTINCT FROM author.team_id)
       )::boolean AS is_read_by_recipient
FROM messages m
JOIN users author ON author.id = m.author_id
LEFT JOIN conversation_reads r
    ON r.conversation_id = m.conversation_id AND r.user_id = sqlc.arg(reader_id)
WHERE m.conversation_id = sqlc.arg(conversation_id)
ORDER BY m.id;

-- Every conversation the reader takes part in, the latest activity first, with
-- its last message, how many are unread and whether the reader wrote in it.
-- name: ListConversations :many
SELECT c.id, c.subject, c.created_at,
       t.id AS team_id, t.code AS team_code, t.name AS team_name,
       c.other_team_id, COALESCE(other.name, 'FIA')::text AS other_team_name,
       last.id AS last_message_id, last.body AS last_message_body,
       last.created_at AS last_message_created_at, last.author_id AS last_message_author_id,
       last_author.username AS last_message_author_username, last_author.role AS last_message_author_role,
       (
           SELECT count(*)
           FROM messages unread
           WHERE unread.conversation_id = c.id
             AND unread.author_id <> sqlc.arg(reader_id)
             AND unread.id > COALESCE(r.last_read_message_id, 0)
       ) AS unread_count,
       EXISTS (
           SELECT 1 FROM messages own
           WHERE own.conversation_id = c.id AND own.author_id = sqlc.arg(reader_id)
       ) AS has_own_messages
FROM conversations c
JOIN teams t ON t.id = c.team_id
LEFT JOIN teams other ON other.id = c.other_team_id
JOIN LATERAL (
    SELECT id, body, created_at, author_id
    FROM messages
    WHERE messages.conversation_id = c.id
    ORDER BY id DESC
    LIMIT 1
) last ON true
JOIN users last_author ON last_author.id = last.author_id
LEFT JOIN conversation_reads r ON r.conversation_id = c.id AND r.user_id = sqlc.arg(reader_id)
WHERE conversation_has_participant(sqlc.arg(reader_id), c.id)
ORDER BY last.id DESC;

-- The cursor only moves forward, so reading an older page again is a no-op.
-- name: MarkConversationRead :exec
INSERT INTO conversation_reads (conversation_id, user_id, last_read_message_id)
VALUES (sqlc.arg(conversation_id), sqlc.arg(user_id), sqlc.arg(last_read_message_id))
ON CONFLICT (conversation_id, user_id) DO UPDATE
SET last_read_message_id = EXCLUDED.last_read_message_id
WHERE conversation_reads.last_read_message_id < EXCLUDED.last_read_message_id;
