-- name: GetMessagesBefore :many
SELECT chat_id, seqnum, sender_id, content, sent_at
  FROM messages
 WHERE chat_id = @chat_id
   AND seqnum < @before
 ORDER BY seqnum DESC
 LIMIT @row_limit;
