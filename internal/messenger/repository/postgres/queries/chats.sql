-- name: IsChatMember :one
SELECT EXISTS (
    SELECT 1
      FROM chat_members m
      JOIN chats c ON c.id = m.chat_id
     WHERE m.chat_id = @chat_id
       AND m.user_id = @user_id
       AND NOT c.deleted
);
