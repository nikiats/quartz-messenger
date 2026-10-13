package cassandra

import (
	"quaztz/internal/messenger/repository/cassandra/db"

	"github.com/scylladb/gocqlx/v3/qb"
)

var (
	getMessage = db.Messages.GetBuilder()

	insertMessage = qb.Insert(db.Messages.Name()).
			Columns("chat_id", "bucket", "message_id", "sender_id", "content")

	getMessagesBefore = db.Messages.SelectBuilder().
				Where(qb.Lt("message_id")).
				LimitNamed("limit")

	getMessagesAfter = db.Messages.SelectBuilder().
				Where(qb.Gt("message_id")).
				OrderBy("message_id", qb.ASC).
				LimitNamed("limit")

	getBucketsBefore = db.MessageBuckets.SelectBuilder("bucket").
				Where(qb.LtOrEq("bucket"))

	getBucketsAfter = db.MessageBuckets.SelectBuilder("bucket").
			Where(qb.GtOrEq("bucket")).
			OrderBy("bucket", qb.ASC)

	insertBucket = db.MessageBuckets.InsertBuilder()

	getChatActivity    = db.ChatActivity.GetBuilder()
	upsertChatActivity = db.ChatActivity.InsertBuilder()
)

var (
	getChat        = db.ChatMembers.GetBuilder()
	getChatMembers = db.ChatMembers.SelectBuilder("user_id", "joined_at", "admin_id")

	insertChatMember = db.ChatMembers.InsertBuilder().Unique()
	deleteChatMember = db.ChatMembers.DeleteBuilder()

	addChatMember = db.ChatMembers.UpdateBuilder("joined_at", "version").
			If(qb.EqNamed("version", "expected_version"))

	updateChat = qb.Update(db.ChatMembers.Name()).
			Set("admin_id", "deleted", "version").
			Where(qb.Eq("chat_id")).
			If(qb.EqNamed("version", "expected_version"))

	insertPrivateChat = db.PrivateChats.InsertBuilder().Unique()
	deletePrivateChat = db.PrivateChats.DeleteBuilder().If(qb.Eq("chat_id"))
)

var (
	getUserChats   = db.UserChats.SelectBuilder("chat_id")
	insertUserChat = db.UserChats.InsertBuilder()
	deleteUserChat = db.UserChats.DeleteBuilder()
)
