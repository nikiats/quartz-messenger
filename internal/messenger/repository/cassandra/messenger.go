package cassandra

import (
	"quaztz/internal/messenger/domain"
	"quaztz/internal/platform/snowflake"
	"time"

	"github.com/scylladb/gocqlx/v3"
)

const (
	bucketSize       = 10 * 24 * time.Hour
	previewLength    = 100
	chatsConcurrency = 16
)

type Messenger struct {
	session gocqlx.Session
}

func NewMessenger(session gocqlx.Session) *Messenger {
	return &Messenger{
		session: session,
	}
}

func (r *Messenger) execBatchCAS(batch *gocqlx.Batch) (bool, error) {
	applied, iter, err := r.session.MapExecuteBatchCAS(batch, map[string]any{})
	if err != nil {
		return false, err
	}
	return applied, iter.Close()
}

func checkVersion(applied bool, err error) error {
	if err != nil {
		return err
	}
	if !applied {
		return domain.ErrChatConflict
	}
	return nil
}

func bucket(messageID int64) int32 {
	return int32(snowflake.Time(messageID).UnixMilli() / bucketSize.Milliseconds())
}
