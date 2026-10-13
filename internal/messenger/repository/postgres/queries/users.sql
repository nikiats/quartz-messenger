-- name: CreateUser :exec
INSERT INTO users (id, login, password_hash)
VALUES ($1, $2, $3);

-- name: GetUser :one
SELECT id, COALESCE(username, '')::text AS username, visible_name
FROM users
WHERE id = $1;

-- name: GetVisibleNames :many
SELECT id, visible_name
FROM users
WHERE id = ANY(@ids::uuid[]);

-- name: GetCredentials :one
SELECT id AS user_id, password_hash, suspended
FROM users
WHERE login = $1;

-- name: SearchUsers :many
SELECT id, COALESCE(username, '')::text AS username, visible_name
FROM users
-- поиск по началу имени пользователя
WHERE starts_with(lower(username), lower(@query::text))
AND NOT suspended
ORDER BY username
LIMIT @row_limit;

-- name: UpdateProfile :exec
UPDATE users
SET
-- COALESCE берет новое значение если не NULL, иначе прошлое
visible_name = COALESCE(sqlc.narg(visible_name), visible_name),
username = COALESCE(sqlc.narg(username), username)
WHERE id = @id;

-- name: IsActiveUser :one
SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND NOT suspended);

-- name: CountActiveUsers :one
SELECT count(*)
FROM users
-- подсчет для всех переданных ID пользователей
WHERE id = ANY(@ids::uuid[])
AND NOT suspended;
