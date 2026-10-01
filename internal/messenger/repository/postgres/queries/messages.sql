-- name: InsertMessage :one
WITH dup AS (
    SELECT m.seqnum, m.sent_at
      FROM messages m
     WHERE m.chat_id = @chat_id
       AND m.sender_id = @sender_id
       AND m.client_message_id = @client_message_id
), updchat AS (
    UPDATE chats c
       SET last_seqnum     = c.last_seqnum + 1,
           last_message_at = clock_timestamp()
     WHERE c.id = @chat_id
       AND NOT EXISTS (SELECT 1 FROM dup)
    RETURNING last_seqnum, last_message_at
), ins AS (
    INSERT INTO messages (chat_id, seqnum, sender_id, content, client_message_id, sent_at)
    SELECT @chat_id, updchat.last_seqnum, @sender_id, @content, @client_message_id, updchat.last_message_at
      FROM updchat
    RETURNING seqnum, sent_at
)
SELECT seqnum, sent_at FROM ins
UNION ALL
SELECT seqnum, sent_at FROM dup;

-- name: GetMessagesBefore :many
SELECT *
  FROM messages
 WHERE chat_id = @chat_id
   AND seqnum < @before
 ORDER BY seqnum DESC
 LIMIT @row_limit;

-- name: GetMessagesAfter :many
SELECT *
  FROM messages
 WHERE chat_id = @chat_id
   AND seqnum > @after
 ORDER BY seqnum ASC
 LIMIT @row_limit;

-- name: SearchMessagesBefore :many
SELECT *
  FROM messages
 WHERE chat_id = @chat_id
   AND seqnum < @before
   AND strpos(lower(content), lower(@query::text)) > 0
 ORDER BY seqnum DESC
 LIMIT @row_limit;
