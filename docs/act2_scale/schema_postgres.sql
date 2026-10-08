-- PostgreSQL / Citus: пользователи и сессии.

CREATE EXTENSION IF NOT EXISTS citext;

-- ============================================================
--  Пользователи
-- ============================================================

-- распределение по id
CREATE TABLE users (
    id            UUID        PRIMARY KEY,
    login         CITEXT      NOT NULL,
    password_hash TEXT        NOT NULL,
    username      CITEXT,
    visible_name  TEXT        NOT NULL DEFAULT 'Пользователь',
    suspended     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
--  Уникальные индексы пользователей
-- ============================================================

-- распределение по login, уникальность логина
CREATE TABLE user_logins (
    login   CITEXT PRIMARY KEY,
    user_id UUID   NOT NULL
);

-- распределение по username, уникальность юзернейма
CREATE TABLE usernames (
    username CITEXT PRIMARY KEY,
    user_id  UUID   NOT NULL
);

-- ============================================================
--  Сессии
-- ============================================================

CREATE TABLE sessions (
    user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    id                 UUID        NOT NULL,
    refresh_token_hash TEXT        NOT NULL,         -- хранится хеш
    expires_at         TIMESTAMPTZ NOT NULL,         -- срок жизни токена
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at         TIMESTAMPTZ DEFAULT NULL,

    PRIMARY KEY (user_id, id)
);

-- ============================================================
--  Распределение (только Citus)
-- ============================================================

-- SELECT create_distributed_table('users',       'id');
-- SELECT create_distributed_table('sessions',    'user_id', colocate_with => 'users');
-- SELECT create_distributed_table('user_logins', 'login');
-- SELECT create_distributed_table('usernames',   'username');
