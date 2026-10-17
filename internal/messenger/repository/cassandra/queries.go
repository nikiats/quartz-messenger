package cassandra

import (
	"context"
	"quartz/internal/messenger/repository/cassandra/db"

	"github.com/scylladb/gocqlx/v3"
	"github.com/scylladb/gocqlx/v3/qb"
)

// заранее собираемые запросы CQL
type cqlQuery struct {
	cql   string
	names []string
}

func newQuery(b qb.Builder) cqlQuery {
	cql, names := b.ToCql()
	return cqlQuery{
		cql:   cql,
		names: names,
	}
}

func (s cqlQuery) query(ctx context.Context, session gocqlx.Session) *gocqlx.Queryx {
	return session.ContextQuery(ctx, s.cql, s.names)
}

var (
	insertMessage = newQuery(
		qb.Insert(
			db.Messages.Name(),
		).Columns("chat_id", "bucket", "message_id", "sender_id", "content"),
	)
	getMessage = newQuery(
		db.Messages.GetBuilder(),
	)
	getMessagesBefore = newQuery(
		db.Messages.SelectBuilder().
			Where(qb.Lt("message_id")).
			LimitNamed("limit"),
	)
	getMessagesAfter = newQuery(
		db.Messages.SelectBuilder().
			Where(qb.Gt("message_id")).
			OrderBy("message_id", qb.ASC).
			LimitNamed("limit"),
	)

	insertBucket = newQuery(
		db.MessageBuckets.InsertBuilder(),
	)
	getBucketsBefore = newQuery(
		db.MessageBuckets.SelectBuilder("bucket").
			Where(qb.LtOrEq("bucket")),
	)
	getBucketsAfter = newQuery(
		db.MessageBuckets.SelectBuilder("bucket").
			Where(qb.GtOrEq("bucket")).
			OrderBy("bucket", qb.ASC),
	)

	upsertChatActivity = newQuery(
		db.ChatActivity.InsertBuilder(),
	)
	getChatActivity = newQuery(
		db.ChatActivity.GetBuilder(),
	)
)

var (
	insertChat = newQuery(
		qb.Update(
			db.ChatMembers.Name(),
		).Set("type", "title", "user_a", "user_b", "admin_id", "deleted", "version").
			Where(qb.Eq("chat_id")).
			If(qb.EqLit("version", "null")),
	)
	insertChatMember = newQuery(
		qb.Insert(
			db.ChatMembers.Name(),
		).Columns("chat_id", "user_id", "joined_at"),
	)
	getChat = newQuery(
		db.ChatMembers.GetBuilder(),
	)
	getChatMembers = newQuery(
		db.ChatMembers.SelectBuilder("user_id", "joined_at", "admin_id"),
	)
	deleteChatMember = newQuery(
		db.ChatMembers.DeleteBuilder(),
	)
	addChatMember = newQuery(
		db.ChatMembers.UpdateBuilder("joined_at", "version").
			If(qb.EqNamed("version", "expected_version")),
	)
	updateChat = newQuery(
		qb.Update(
			db.ChatMembers.Name(),
		).Set("admin_id", "deleted", "version").
			Where(qb.Eq("chat_id")).
			If(qb.EqNamed("version", "expected_version")),
	)

	insertPrivateChat = newQuery(
		db.PrivateChats.InsertBuilder().Unique(),
	)
	deletePrivateChat = newQuery(
		db.PrivateChats.DeleteBuilder().
			If(qb.Eq("chat_id")),
	)
)

var (
	insertUserChat = newQuery(
		db.UserChats.InsertBuilder(),
	)
	getUserChats = newQuery(
		db.UserChats.SelectBuilder("chat_id"),
	)
	deleteUserChat = newQuery(
		db.UserChats.DeleteBuilder(),
	)
)
