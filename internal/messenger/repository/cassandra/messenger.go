package cassandra

import (
	"quaztz/internal/common/snowflake"
	"quaztz/internal/messenger/domain"
	"time"

	"github.com/scylladb/gocqlx/v3"
)

const (
	bucketSize       = 10 * 24 * time.Hour
	previewLength    = 100
	queryConcurrency = 16
)

type Messenger struct {
	session gocqlx.Session
}

func NewMessenger(session gocqlx.Session) *Messenger {
	return &Messenger{
		session: session,
	}
}

func (r *Messenger) executeBatch(batch *gocqlx.Batch) (bool, error) {
	applied, iter, err := r.session.MapExecuteBatchCAS(batch, map[string]any{})
	if err != nil {
		return false, err
	}
	return applied, iter.Close()
}

func ensureVersion(applied bool) error {
	if !applied {
		return domain.ErrChatConflict
	}
	return nil
}

func bucketForMessage(messageID int64) int32 {
	ts := snowflake.Time(messageID).UnixMilli()
	return int32(ts / bucketSize.Milliseconds())
}
