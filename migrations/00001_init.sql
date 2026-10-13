-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id            UUID        PRIMARY KEY,
    login         CITEXT      NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    username      CITEXT      UNIQUE,
    visible_name  TEXT        NOT NULL DEFAULT 'Пользователь',
    suspended     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id                 UUID        NOT NULL,
    user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT        NOT NULL,
    expires_at         TIMESTAMPTZ NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at         TIMESTAMPTZ DEFAULT NULL,

    PRIMARY KEY (user_id, id)
);

-- +goose Down
DROP TABLE sessions;
DROP TABLE users;
