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
    id                 UUID        PRIMARY KEY,
    user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT        NOT NULL UNIQUE,  -- хранится хеш, не сам токен
    expires_at         TIMESTAMPTZ NOT NULL,         -- срок жизни токена обновления
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at         TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
