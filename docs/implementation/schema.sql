-- Схема базы данных мессенджера.

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
    -- название задано тогда и только тогда, когда чат групповой
    CHECK ( (type = 'group') = (title IS NOT NULL) ),
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
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),

    -- сообщение идентифицируется парой чата и порядкового номера;
    -- этот же индекс обслуживает пагинацию и подгрузку неполученных сообщений
    PRIMARY KEY (chat_id, seqnum),

    -- идемпотентность повторной отправки в пределах чата
    UNIQUE (chat_id, sender_id, client_message_id)
);
