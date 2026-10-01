-- name: GetUser :one
SELECT id, username, visible_name
  FROM users
 WHERE id = $1;
