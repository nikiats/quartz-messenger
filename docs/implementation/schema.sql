-- Схема базы данных мессенджера.

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS citext;


-- ============================================================
--  Перечисления
-- ============================================================

CREATE TYPE chat_type   AS ENUM ('private', 'group');

-- ============================================================
--  Пользователи
-- ============================================================

CREATE TABLE users (
    id            UUID        PRIMARY KEY,
    login         CITEXT      NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    username      CITEXT      UNIQUE,
    visible_name  TEXT        NOT NULL DEFAULT 'Пользователь',        
    suspended     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);


-- ============================================================
--  Сессии
-- ============================================================

CREATE TABLE sessions (
    id                 UUID        PRIMARY KEY,
    user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT        NOT NULL UNIQUE,  -- хранится хеш, не сам токен
    expires_at         TIMESTAMPTZ NOT NULL,         -- срок жизни токена обновления
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at         TIMESTAMPTZ DEFAULT NULL
);

-- отзыв всех сессий пользователя
CREATE INDEX sessions_user_id_idx ON sessions (user_id);


-- ============================================================
--  Чаты
-- ============================================================

CREATE TABLE chats (
    id              UUID        PRIMARY KEY,
    type            chat_type   NOT NULL,
    title           TEXT,                              -- пусто у личных чатов
    user_a          UUID        REFERENCES users(id),  -- участники личного чата;
    user_b          UUID        REFERENCES users(id),  -- пусто у групповых
    last_seqnum     BIGINT      NOT NULL DEFAULT 0,    -- счетчик seqnum данного чата
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT now(),  -- сортировка списка чатов
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    admin           UUID        REFERENCES users(id),  -- пусто у личных чатов
    deleted         BOOLEAN     NOT NULL DEFAULT FALSE,

    -- участники заданы тогда и только тогда, когда чат личный
    CHECK ( (type = 'private') = (user_a IS NOT NULL AND user_b IS NOT NULL) ),
    -- администратор задан тогда и только тогда, когда чат групповой
    CHECK ( (type = 'group') = (admin IS NOT NULL) ),
    -- канонический порядок пары: идентификатор чата не зависит от инициатора
    CHECK ( user_a IS NULL OR user_a < user_b )
);

-- один личный чат на пару участников
CREATE UNIQUE INDEX chats_private_pair_idx
    ON chats (user_a, user_b) WHERE type = 'private' AND NOT deleted;

-- ============================================================
--  Участники чатов
-- ============================================================

CREATE TABLE chat_members (
    chat_id   UUID        NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (chat_id, user_id)
);

-- список чатов пользователя
CREATE INDEX chat_members_user_id_idx ON chat_members (user_id);

-- ============================================================
--  Сообщения
-- ============================================================

CREATE TABLE messages (
    chat_id           UUID        NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    seqnum            BIGINT      NOT NULL,
    sender_id         UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content           TEXT        NOT NULL,
    client_message_id UUID        NOT NULL,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- сообщение идентифицируется парой чата и порядкового номера;
    -- этот же индекс обслуживает пагинацию и подгрузку неполученных сообщений
    PRIMARY KEY (chat_id, seqnum),

    -- идемпотентность повторной отправки
    UNIQUE (sender_id, client_message_id)
);

-- поиск по подстроке в пределах чата
CREATE INDEX messages_content_trgm_idx ON messages USING GIN (content gin_trgm_ops);


-- ============================================================
--  Добавление сообщения в чат
-- ============================================================
--
--  Блокирующий запрос отбирает только живой чат. Ноль строк означает,
--  что чата нет либо он удален: транзакция прерывается, следующий
--  запрос не выполняется.
--
-- BEGIN;

-- SELECT last_seqnum FROM chats WHERE id = $1 AND NOT deleted FOR UPDATE;

-- WITH dup AS (
--     SELECT seqnum, sent_at FROM messages
--      WHERE sender_id = $2 AND client_message_id = $4
-- ), updchat AS (
--     UPDATE chats SET last_seqnum = last_seqnum + 1, last_message_at = NOW()
--      WHERE id = $1 AND NOT EXISTS (SELECT 1 FROM dup)
--     RETURNING last_seqnum
-- ), ins AS (
--     INSERT INTO messages (chat_id, seqnum, sender_id, content, client_message_id)
--     SELECT $1, updchat.last_seqnum, $2, $3, $4 FROM updchat
--     RETURNING seqnum, sent_at
-- )
-- SELECT * FROM ins UNION ALL SELECT * FROM dup;

-- COMMIT;
--

-- ============================================================
--  Удаление участника из чата
-- ============================================================
--
--  Условие на администратора включено в сам запрос: удалить админа,
--  не передав права, невозможно. Ноль затронутых строк означает либо
--  отсутствие участника, либо попытку удалить администратора.
--
-- BEGIN;

-- SELECT 1 FROM chats WHERE id = $1 FOR UPDATE;
-- DELETE FROM chat_members
--  WHERE chat_id = $1
--    AND user_id = $2
--    AND user_id <> (SELECT admin FROM chats WHERE id = $1);

-- COMMIT;


-- ============================================================
--  Выход администратора из группового чата
-- ============================================================
--
--  Права передаются псевдослучайно выбранному участнику. Если
--  преемника нет, значит администратор -- последний участник: чат
--  отмечается удаленным, а администратор остается прежним, поскольку
--  ограничение требует непустого администратора у группового чата.
--
--  Преемник выбирается в CTE: обращений к нему два, а `random()`
--  должен отработать один раз.
--
--  Строка чата блокируется до выбора преемника: иначе параллельный
--  выход выбранного участника оставит чат с администратором, который
--  уже не состоит в чате.
--
-- BEGIN;
--
-- SELECT 1 FROM chats WHERE id = $1 AND admin = $2 AND NOT deleted FOR UPDATE;
--
-- WITH successor AS (
--     SELECT user_id FROM chat_members
--      WHERE chat_id = $1 AND user_id <> $2
--      ORDER BY random() LIMIT 1
-- )
-- UPDATE chats
--    SET admin   = COALESCE((SELECT user_id FROM successor), admin),
--        deleted = NOT EXISTS (SELECT 1 FROM successor)
--  WHERE id = $1;
--
-- DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2;
--
-- COMMIT;
