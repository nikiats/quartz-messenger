CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE
	users (
		id UUID PRIMARY KEY,
		login CITEXT NOT NULL,
		password_hash TEXT NOT NULL,
		username CITEXT,
		visible_name TEXT NOT NULL DEFAULT 'Пользователь',
		suspended BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now ()
	);

CREATE TABLE
	user_logins (login CITEXT PRIMARY KEY, user_id UUID NOT NULL);

CREATE TABLE
	usernames (
		username CITEXT PRIMARY KEY,
		user_id UUID NOT NULL
	);

CREATE TABLE
	sessions (
		user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
		id UUID NOT NULL,
		refresh_token_hash TEXT NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
		revoked_at TIMESTAMPTZ DEFAULT NULL,
		PRIMARY KEY (user_id, id)
	);