-- name: LockChat :one
SELECT type, admin
  FROM chats
 WHERE id = $1 AND NOT deleted
   FOR UPDATE;

-- name: IsChatMember :one
SELECT EXISTS (
    SELECT 1
      FROM chat_members m
      JOIN chats c ON c.id = m.chat_id
     WHERE m.chat_id = @chat_id
       AND m.user_id = @user_id
       AND NOT c.deleted
);

-- name: GetChats :many
SELECT c.id, c.type, COALESCE(c.title, peer.visible_name)::text AS title,
       c.last_seqnum, c.last_message_at
  FROM chat_members m
  JOIN chats c ON c.id = m.chat_id
  LEFT JOIN users peer
         ON c.type = 'private'
        AND peer.id = CASE WHEN c.user_a = m.user_id THEN c.user_b ELSE c.user_a END
 WHERE m.user_id = @user_id
   AND NOT c.deleted
 ORDER BY c.last_message_at DESC;

-- name: GetChat :one
SELECT c.id, c.type, COALESCE(c.title, peer.visible_name)::text AS title,
       c.last_seqnum, c.last_message_at
  FROM chat_members m
  JOIN chats c ON c.id = m.chat_id
  LEFT JOIN users peer
         ON c.type = 'private'
        AND peer.id = CASE WHEN c.user_a = m.user_id THEN c.user_b ELSE c.user_a END
 WHERE m.chat_id = @chat_id
   AND m.user_id = @user_id
   AND NOT c.deleted;

-- name: GetChatMembers :many
SELECT m.user_id, COALESCE(c.admin = m.user_id, FALSE)::boolean AS is_admin
  FROM chat_members m
  JOIN chats c ON c.id = m.chat_id
 WHERE m.chat_id = $1
 ORDER BY m.joined_at;

-- name: InsertPrivateChat :one
INSERT INTO chats (id, type, user_a, user_b)
VALUES (@id, 'private', LEAST(@user_id::uuid, @peer_id::uuid), GREATEST(@user_id::uuid, @peer_id::uuid))
ON CONFLICT (user_a, user_b) WHERE type = 'private' AND NOT deleted DO NOTHING
RETURNING id;

-- name: FindPrivateChat :one
SELECT id
  FROM chats
 WHERE type = 'private'
   AND NOT deleted
   AND user_a = LEAST(@user_id::uuid, @peer_id::uuid)
   AND user_b = GREATEST(@user_id::uuid, @peer_id::uuid);

-- name: InsertGroupChat :exec
INSERT INTO chats (id, type, title, admin)
VALUES (@id, 'group', @title::text, @admin::uuid);

-- name: InsertChatMembers :exec
INSERT INTO chat_members (chat_id, user_id)
SELECT @chat_id, unnest(@user_ids::uuid[]);

-- name: AddChatMember :one
INSERT INTO chat_members (chat_id, user_id)
VALUES (@chat_id, @user_id)
ON CONFLICT DO NOTHING
RETURNING user_id;

-- name: DeleteChatMember :execrows
DELETE FROM chat_members
 WHERE chat_id = @chat_id
   AND user_id = @user_id
   AND user_id IS DISTINCT FROM (SELECT admin FROM chats WHERE id = @chat_id);

-- name: PassAdminOnLeave :exec
WITH successor AS (
    SELECT user_id
      FROM chat_members
     WHERE chat_id = @chat_id AND user_id <> @admin_id
     ORDER BY random()
     LIMIT 1
)
UPDATE chats
   SET admin   = COALESCE((SELECT user_id FROM successor), admin),
       deleted = NOT EXISTS (SELECT 1 FROM successor)
 WHERE id = @chat_id;

-- name: DeleteChat :exec
UPDATE chats
   SET deleted = TRUE
 WHERE id = $1;
