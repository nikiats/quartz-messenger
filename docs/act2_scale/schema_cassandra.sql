-- Scylla / Cassandra: 

-- общие настройки базы данных
CREATE KEYSPACE IF NOT EXISTS quaztz
    WITH replication = {'class': 'NetworkTopologyStrategy', 'replication_factor': 1};

USE quaztz;

-- ============================================================
--  Сообщения
-- ============================================================

-- bucket - номер окна фиксированного размера по времени
CREATE TABLE IF NOT EXISTS messages (
    chat_id           bigint,     -- snowflake ID
    bucket            int,
    message_id        bigint,
    sender_id         uuid,
    content           text,
    edited_at         timestamp,
    deleted           boolean,    -- удаление отметкой

    -- партиция - bucket в одном чате
	-- сортировка - по message_id
    PRIMARY KEY ((chat_id, bucket), message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);

-- ============================================================
--  Непустые бакеты чата (оглавление чата)
-- ============================================================

CREATE TABLE IF NOT EXISTS message_buckets (
    chat_id bigint,
    bucket  int,

    PRIMARY KEY ((chat_id), bucket)
) WITH CLUSTERING ORDER BY (bucket DESC);

-- ============================================================
--  Чаты и их участники
-- ============================================================

-- участники чата, партиция - один чат

-- чат описывается static столбцами:
-- * chat_id
-- * type
-- * title
-- * user_a
-- * user_b
-- * admin_id
-- * deleted

-- участник чата описывается
-- * user_id
-- * joined_at

-- для согласованности изменений хранится "версия" чата
-- * version

CREATE TABLE IF NOT EXISTS chat_members (
    chat_id   bigint,
    user_id   uuid,
    joined_at timestamp,

    type      text    static,   -- 'private' | 'group'
    title     text    static,   -- название (группа)
    user_a    uuid    static,   -- участник 1 (личный чат)
    user_b    uuid    static,   -- участник 2 (личный чат)
    admin_id  uuid    static,   -- администратор (группа)
    deleted   boolean static,   -- мягкое удаление
    version   int     static,   -- номер редакции чата для LWT

    -- партиция - чат
    -- сортировка - по user_id
    PRIMARY KEY ((chat_id), user_id)
);

-- ============================================================
--  Последнее сообщение чата
-- ============================================================

-- для отображения preview чатов
CREATE TABLE IF NOT EXISTS chat_activity (
    chat_id                bigint PRIMARY KEY,
    last_message_id        bigint,
    last_message_sender_id uuid,
    last_message_preview   text
);

-- для отображения чатов пользователя
CREATE TABLE IF NOT EXISTS user_chats (
    user_id   uuid,
    chat_id   bigint,
    joined_at timestamp,

    PRIMARY KEY ((user_id), chat_id)
);

-- ============================================================
--  Личные чаты
-- ============================================================

-- выполняет роль уникального индекса
-- user_a < user_b
CREATE TABLE IF NOT EXISTS private_chats (
    user_a  uuid,
    user_b  uuid,
    chat_id bigint,

    PRIMARY KEY ((user_a, user_b))
);
