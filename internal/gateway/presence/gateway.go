package presence

import (
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Maintainer struct {
	rdb       *redis.Client
	gatewayID uuid.UUID
}

func NewMaintainer(rdb *redis.Client, gatewayID uuid.UUID) *Maintainer {
	return &Maintainer{
		rdb:       rdb,
		gatewayID: gatewayID,
	}
}

func (m *Maintainer) Perform() {

}
