-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: RefreshSession :execrows
UPDATE sessions
SET
refresh_token_hash = @new_refresh_token_hash,
expires_at = @expires_at
WHERE
user_id = @user_id
AND sessions.id = @id
AND refresh_token_hash = @refresh_token_hash
AND revoked_at IS NULL
AND expires_at > now()
AND EXISTS (SELECT 1 FROM users u WHERE u.id = @user_id AND NOT u.suspended);

-- name: RevokeSession :execrows
UPDATE sessions
SET revoked_at = now()
WHERE user_id = @user_id
AND id = @id
AND revoked_at IS NULL;
