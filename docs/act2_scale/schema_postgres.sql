-- PostgreSQL: пользователи и сессии.

CREATE EXTENSION IF NOT EXISTS citext;

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
    user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    id                 UUID        NOT NULL,
    refresh_token_hash TEXT        NOT NULL,         -- хранится хеш
    expires_at         TIMESTAMPTZ NOT NULL,         -- срок жизни токена
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at         TIMESTAMPTZ DEFAULT NULL,

    PRIMARY KEY (user_id, id)
);
