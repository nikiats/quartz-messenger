-- Схема базы данных мессенджера.

CREATE EXTENSION IF NOT EXISTS pg_trgm;


-- ============================================================
--  Перечисления
-- ============================================================

CREATE TYPE chat_type   AS ENUM ('private', 'group');
CREATE TYPE member_role AS ENUM ('admin', 'regular');


-- ============================================================
--  Пользователи
-- ============================================================

CREATE TABLE users (
    id            UUID        PRIMARY KEY,
    login         TEXT        NOT NULL UNIQUE,   -- для входа
    password_hash TEXT        NOT NULL,
    username      TEXT        UNIQUE,            -- для упоминаний
    visible_name  TEXT        NOT NULL,
	suspended     BOOLEAN     DEFAULT FALSE,
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
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
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

    -- участники заданы тогда и только тогда, когда чат личный
    CHECK ( (type = 'private') = (user_a IS NOT NULL AND user_b IS NOT NULL) ),

    -- канонический порядок пары: идентификатор чата не зависит от инициатора
    CHECK ( user_a IS NULL OR user_a < user_b )
);

-- один личный чат на пару участников
CREATE UNIQUE INDEX chats_private_pair_idx
    ON chats (user_a, user_b) WHERE type = 'private';


-- ============================================================
--  Участники чатов
-- ============================================================

CREATE TABLE chat_members (
    chat_id   UUID        NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role      member_role NOT NULL DEFAULT 'regular',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (chat_id, user_id)
);

-- список чатов пользователя
CREATE INDEX chat_members_user_id_idx ON chat_members (user_id);

-- не более одного администратора в чате
CREATE UNIQUE INDEX chat_members_single_admin_idx
    ON chat_members (chat_id) WHERE role = 'admin';


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
--  Выдача порядкового номера
-- ============================================================
--
--  BEGIN;
--      -- повтор: отдать ранее сохраненное, счетчик не трогать
--      SELECT seqnum, sent_at FROM messages
--       WHERE sender_id = $1 AND client_message_id = $2;
--
--      -- новое сообщение: инкремент под блокировкой строки чата
--      UPDATE chats
--         SET last_seqnum = last_seqnum + 1,
--             last_message_at = now()
--       WHERE id = $3
--       RETURNING last_seqnum, last_message_at;
--
--      INSERT INTO messages (chat_id, seqnum, sender_id, content, client_message_id)
--      VALUES ($3, $4, $1, $5, $2);
--  COMMIT;
--
--  Блокировка строки чата удерживается до конца транзакции, поэтому порядок
--  выдачи номеров совпадает с порядком фиксации. При откате счетчик
--  восстанавливается, и нумерация остается непрерывной.
