-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3, $4);
