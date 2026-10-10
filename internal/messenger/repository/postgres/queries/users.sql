-- name: CreateLogin :exec
INSERT INTO user_logins (login, user_id)
VALUES (@login, @user_id);

-- name: CreateUser :exec
INSERT INTO users (id, login, password_hash)
VALUES (@id, @login, @password_hash);

-- name: GetUser :one
SELECT id, username, visible_name
  FROM users
 WHERE id = $1;

-- name: GetVisibleNames :many
SELECT id, visible_name
  FROM users
 WHERE id = ANY(@ids::uuid[]);

-- name: GetCredentials :one
SELECT id, password_hash, suspended
  FROM users
 WHERE id = (SELECT l.user_id FROM user_logins l WHERE l.login = $1);

-- name: SearchUsers :many
SELECT id, username, visible_name
  FROM users
 WHERE starts_with(lower(username), lower(@query::text))
   AND NOT suspended
 ORDER BY username
 LIMIT @row_limit;

-- name: LockUsername :one
SELECT username
  FROM users
 WHERE id = $1
   FOR UPDATE;

-- name: CreateUsername :exec
INSERT INTO usernames (username, user_id)
VALUES (@username, @user_id);

-- name: DeleteUsername :exec
DELETE FROM usernames
 WHERE username = $1;

-- name: UpdateProfile :exec
UPDATE users
   SET visible_name = COALESCE(sqlc.narg(visible_name), visible_name),
       username     = COALESCE(sqlc.narg(username), username)
 WHERE id = @id;

-- name: IsActiveUser :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE id = $1 AND NOT suspended
);

-- name: CountActiveUsers :one
SELECT count(*)
  FROM users
 WHERE id = ANY(@ids::uuid[])
   AND NOT suspended;
