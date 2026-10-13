-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE chat_type   AS ENUM ('private', 'group');

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
    id                 UUID        PRIMARY KEY,
    user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT        NOT NULL UNIQUE,
    expires_at         TIMESTAMPTZ NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at         TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);


CREATE TABLE chats (
    id              UUID        PRIMARY KEY,
    type            chat_type   NOT NULL,
    title           TEXT,
    user_a          UUID        REFERENCES users(id),
    user_b          UUID        REFERENCES users(id),
    last_seqnum     BIGINT      NOT NULL DEFAULT 0,
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    admin           UUID        REFERENCES users(id),
    deleted         BOOLEAN     NOT NULL DEFAULT FALSE,

    CHECK ( (type = 'private') = (user_a IS NOT NULL AND user_b IS NOT NULL) ),
    CHECK ( (type = 'group') = (admin IS NOT NULL) ),
    CHECK ( (type = 'group') = (title IS NOT NULL) ),
    CHECK ( user_a IS NULL OR user_a < user_b )
);

CREATE UNIQUE INDEX chats_private_pair_idx
    ON chats (user_a, user_b) WHERE type = 'private' AND NOT deleted;

CREATE TABLE chat_members (
    chat_id   UUID        NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX chat_members_user_id_idx ON chat_members (user_id);

CREATE TABLE messages (
    chat_id           UUID        NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    seqnum            BIGINT      NOT NULL,
    sender_id         UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content           TEXT        NOT NULL,
    client_message_id UUID        NOT NULL,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),

    PRIMARY KEY (chat_id, seqnum),

    UNIQUE (chat_id, sender_id, client_message_id)
);

-- +goose Down
DROP TABLE messages;
DROP TABLE chat_members;
DROP TABLE chats;
DROP TABLE sessions;
DROP TABLE users;
DROP TYPE chat_type;
