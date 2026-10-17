-- Scylla / Cassandra: 

CREATE KEYSPACE IF NOT EXISTS quartz_messenger
WITH replication = {'class': 'NetworkTopologyStrategy', 'replication_factor': 1};

USE quartz_messenger;

CREATE TABLE
	IF NOT EXISTS messages (
		chat_id bigint,
		bucket int,
		message_id bigint,
		sender_id uuid,
		content text,
		edited_at timestamp,
		deleted boolean,
		PRIMARY KEY ((chat_id, bucket), message_id)
	)
WITH
	CLUSTERING
ORDER BY
	(message_id DESC);

CREATE TABLE
	IF NOT EXISTS message_buckets (
		chat_id bigint,
		bucket int,
		PRIMARY KEY ((chat_id), bucket)
	)
WITH
	CLUSTERING
ORDER BY
	(bucket DESC);

CREATE TABLE
	IF NOT EXISTS chat_members (
		chat_id bigint,
		user_id uuid,
		joined_at timestamp,
		type text static,
		title text static,
		user_a uuid static,
		user_b uuid static,
		admin_id uuid static,
		deleted boolean static,
		version int static,
		PRIMARY KEY ((chat_id), user_id)
	);

CREATE TABLE
	IF NOT EXISTS chat_activity (
		chat_id bigint PRIMARY KEY,
		last_message_id bigint,
		last_message_sender_id uuid,
		last_message_preview text
	);

CREATE TABLE
	IF NOT EXISTS user_chats (
		user_id uuid,
		chat_id bigint,
		joined_at timestamp,
		PRIMARY KEY ((user_id), chat_id)
	);

CREATE TABLE
	IF NOT EXISTS private_chats (
		user_a uuid,
		user_b uuid,
		chat_id bigint,
		PRIMARY KEY ((user_a, user_b))
	);