-- ============================================================
--  Транзакции над чатом
-- ============================================================
--
--  Во всех запросах ниже $1 -- идентификатор чата, $2 -- инициатор
--  операции (пользователь из токена).
--
--  Операция над существующим чатом начинается с двух шагов. Ноль строк
--  на любом из них прерывает транзакцию, следующий запрос не выполняется.
--
--  1. Блокировка строки чата. Ноль строк -- чата нет либо он удален: 404.
--
      SELECT type, admin FROM chats WHERE id = $1 AND NOT deleted FOR UPDATE;
--
--  2. Проверка членства инициатора. Ноль строк: 404.
--
      SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2;
--
--  Права проверяются по полям type и admin, полученным на шаге 1.


-- ============================================================
--  Создание личного чата
-- ============================================================
--
--  $3 -- собеседник. Пара упорядочивается средствами PostgreSQL в
--  соответствии с ограничением user_a < user_b.
--
--  Совпадение $2 и $3 отсекается сервисом до транзакции (400),
--  иначе запрос нарушит CHECK (user_a < user_b).
--
BEGIN;

-- ноль строк: собеседника нет либо он заблокирован (404)
SELECT 1 FROM users WHERE id = $3 AND NOT suspended;

INSERT INTO chats (id, type, user_a, user_b)
VALUES ($1, 'private', LEAST($2::uuid, $3::uuid), GREATEST($2::uuid, $3::uuid))
ON CONFLICT (user_a, user_b) WHERE type = 'private' AND NOT deleted DO NOTHING
RETURNING id;

-- строка вставлена: чат создан (201)
INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3);

-- ноль строк: чат уже существует (200)
SELECT id FROM chats
 WHERE type = 'private' AND NOT deleted
   AND user_a = LEAST($2::uuid, $3::uuid)
   AND user_b = GREATEST($2::uuid, $3::uuid);

COMMIT;


-- ============================================================
--  Создание группового чата
-- ============================================================
--
--  $3 -- название, $4 -- memberIds (uuid[]). Создатель ($2)
--  становится администратором.
--
--  До транзакции сервис проверяет $4: нет повторов, нет создателя,
--  не более 100 элементов. Иначе 400.
--
BEGIN;

-- найдено меньше, чем cardinality($4): кого-то нет
-- либо он заблокирован (404)
SELECT count(*) FROM users WHERE id = ANY($4::uuid[]) AND NOT suspended;

INSERT INTO chats (id, type, title, admin) VALUES ($1, 'group', $3, $2);

INSERT INTO chat_members (chat_id, user_id)
SELECT $1, u FROM unnest(array_append($4::uuid[], $2::uuid)) AS u;

COMMIT;


-- ============================================================
--  Добавление сообщения в чат
-- ============================================================
--
--  $3 -- текст, $4 -- clientMessageId.
--
--  Принадлежность к чату отправителя проверяется после блокировки.
--
BEGIN;

SELECT 1 FROM chats WHERE id = $1 AND NOT deleted FOR UPDATE;

SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2;

WITH dup AS (
    SELECT seqnum, sent_at FROM messages
     WHERE chat_id = $1 AND sender_id = $2 AND client_message_id = $4
), updchat AS (
    UPDATE chats SET last_seqnum = last_seqnum + 1, last_message_at = NOW()
     WHERE id = $1 AND NOT EXISTS (SELECT 1 FROM dup)
    RETURNING last_seqnum
), ins AS (
    INSERT INTO messages (chat_id, seqnum, sender_id, content, client_message_id)
    SELECT $1, updchat.last_seqnum, $2, $3, $4 FROM updchat
    RETURNING seqnum, sent_at
)
SELECT * FROM ins UNION ALL SELECT * FROM dup;

COMMIT;


-- ============================================================
--  Добавление участника в чат
-- ============================================================
--
--  $3 -- добавляемый пользователь.
--
--  Права (групповой чат, инициатор -- администратор) проверяются по
--  результату шага 1, иначе 403.
--
BEGIN;

SELECT type, admin FROM chats WHERE id = $1 AND NOT deleted FOR UPDATE;

SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2;

-- ноль строк: пользователя нет либо он заблокирован (404)
SELECT 1 FROM users WHERE id = $3 AND NOT suspended;

-- ноль строк: пользователь уже участник (409 already_member)
INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $3)
ON CONFLICT DO NOTHING
RETURNING user_id;

COMMIT;


-- ============================================================
--  Удаление участника из чата
-- ============================================================
--
--  $3 -- удаляемый участник; совпадение с $2 означает выход из чата.
--
--
BEGIN;

SELECT type, admin FROM chats WHERE id = $1 AND NOT deleted FOR UPDATE;

SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2;

DELETE FROM chat_members
 WHERE chat_id = $1
   AND user_id = $3
   AND user_id <> (SELECT admin FROM chats WHERE id = $1);

COMMIT;


-- ============================================================
--  Выход администратора из группового чата
-- ============================================================
--
BEGIN;

SELECT type, admin FROM chats WHERE id = $1 AND NOT deleted FOR UPDATE;

SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2;

WITH successor AS (
    SELECT user_id FROM chat_members
     WHERE chat_id = $1 AND user_id <> $2
     ORDER BY random() LIMIT 1
)
UPDATE chats
   SET admin   = COALESCE((SELECT user_id FROM successor), admin),
       deleted = NOT EXISTS (SELECT 1 FROM successor)
 WHERE id = $1;

DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2;

COMMIT;


-- ============================================================
--  Удаление чата
-- ============================================================
--
--
BEGIN;

SELECT type, admin FROM chats WHERE id = $1 AND NOT deleted FOR UPDATE;

SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2;

UPDATE chats SET deleted = TRUE WHERE id = $1;

COMMIT;
