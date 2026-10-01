-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetSessionForRefresh :one
SELECT s.id, s.user_id, s.expires_at, s.revoked_at, u.suspended
  FROM sessions s
  JOIN users u ON u.id = s.user_id
 WHERE s.refresh_token_hash = $1
   FOR UPDATE OF s;

-- name: UpdateRefreshToken :exec
UPDATE sessions
   SET refresh_token_hash = @refresh_token_hash,
       expires_at         = @expires_at
 WHERE id = @id;

-- name: RevokeSession :execrows
UPDATE sessions
   SET revoked_at = now()
 WHERE id = @id
   AND user_id = @user_id
   AND revoked_at IS NULL;

-- name: GetRevokedSessions :many
SELECT id, revoked_at
  FROM sessions
 WHERE revoked_at > @since::timestamptz;
