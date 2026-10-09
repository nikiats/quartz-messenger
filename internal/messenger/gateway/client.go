package gateway

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"quaztz/internal/api/bus"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/platform/rediskey"
)

// события для шлюзов через NATS и общее с ними состояние в Redis
type Client struct {
	nc         *nats.Conn
	rdb        *redis.Client
	revokedTTL time.Duration
}

func NewClient(nc *nats.Conn, rdb *redis.Client, revokedTTL time.Duration) *Client {
	return &Client{nc: nc, rdb: rdb, revokedTTL: revokedTTL}
}

// отправка сообщения на все шлюзы с получателями
func (c *Client) DeliverMessage(ctx context.Context, recipientIDs []uuid.UUID, message domain.Message) error {
	pipe := c.rdb.Pipeline()

	// кол-во шлюзов <= количеству получателей (худший случай - каждый получатель на своем шлюзе)
	cmds := make([]*redis.StringSliceCmd, len(recipientIDs))
	for i, userID := range recipientIDs {
		// команда на получение всех шлюзов на которых сидит получатель
		cmds[i] = pipe.SMembers(ctx, rediskey.Conn(userID))
	}
	// пакетное выполнение запроса
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	// для каждого шлюза список его пользователей, которым доставить сообщение
	byGateway := map[string][]uuid.UUID{}
	for i, cmd := range cmds {
		for _, gatewayID := range cmd.Val() {
			byGateway[gatewayID] = append(byGateway[gatewayID], recipientIDs[i])
		}
	}

	// каждому рассылаем сообщение + список получателей
	// шлюз не имеет информацию о том, кто есть в каком чате
	for gatewayID, userIDs := range byGateway {
		id, err := uuid.Parse(gatewayID)
		if err != nil {
			continue
		}

		data, err := json.Marshal(bus.MessageEvent{
			ChatID:       message.ChatID,
			SenderID:     message.SenderID,
			Seqnum:       message.Seqnum,
			Content:      bus.MessageContent(message.Content),
			SentAt:       message.SentAt,
			RecipientIDs: userIDs,
		})
		if err != nil {
			return err
		}

		if err := c.nc.Publish(bus.SubjectGatewayMessage(id), data); err != nil {
			return err
		}
	}

	return nil
}

// запись об отзыве сессии + рассылка события отзыва всем шлюзам
func (c *Client) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := c.rdb.Set(ctx, rediskey.Revoked(sessionID), "1", c.revokedTTL).Err(); err != nil {
		return err
	}

	data, err := json.Marshal(bus.SessionRevokedEvent{SessionID: sessionID})
	if err != nil {
		return err
	}

	return c.nc.Publish(bus.SubjectSessionRevoked, data)
}

// проверка отзыва сессии, нужна для каждого запроса с авторизацией
func (c *Client) IsRevoked(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	n, err := c.rdb.Exists(ctx, rediskey.Revoked(sessionID)).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
