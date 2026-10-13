-- +goose Up
DROP TABLE messages;
DROP TABLE chat_members;
DROP TABLE chats;
DROP TYPE chat_type;

CREATE TABLE user_logins (
    login   CITEXT PRIMARY KEY,
    user_id UUID   NOT NULL
);

CREATE TABLE usernames (
    username CITEXT PRIMARY KEY,
    user_id  UUID   NOT NULL
);

INSERT INTO user_logins (login, user_id)
SELECT login, id FROM users;

INSERT INTO usernames (username, user_id)
SELECT username, id FROM users WHERE username IS NOT NULL;

ALTER TABLE users
    DROP CONSTRAINT users_login_key,
    DROP CONSTRAINT users_username_key;

ALTER TABLE sessions
    DROP CONSTRAINT sessions_pkey,
    DROP CONSTRAINT sessions_refresh_token_hash_key,
    ADD PRIMARY KEY (user_id, id);

DROP INDEX sessions_user_id_idx;

-- +goose Down
CREATE INDEX sessions_user_id_idx ON sessions (user_id);

ALTER TABLE sessions
    DROP CONSTRAINT sessions_pkey,
    ADD PRIMARY KEY (id),
    ADD UNIQUE (refresh_token_hash);

ALTER TABLE users
    ADD UNIQUE (login),
    ADD UNIQUE (username);

DROP TABLE usernames;
DROP TABLE user_logins;

CREATE TYPE chat_type AS ENUM ('private', 'group');

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
