package messenger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"

	"quartz/internal/api/bus"
)

const requestTimeout = 10 * time.Second

var ErrSendFailed = errors.New("message send failed")

type Client struct {
	nc *nats.Conn
}

func NewClient(natsConnection *nats.Conn) *Client {
	return &Client{
		nc: natsConnection,
	}
}

func (c *Client) SendMessage(ctx context.Context, request bus.SendMessageRequest) (bus.SendMessageResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return bus.SendMessageResponse{}, fmt.Errorf("marshal send message request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	reply, err := c.nc.RequestWithContext(ctx, bus.SubjectSendMessage, data)
	if err != nil {
		return bus.SendMessageResponse{}, fmt.Errorf("send message request: %w", err)
	}

	switch reply.Header.Get(micro.ErrorCodeHeader) {
	case "":
	case bus.CodeForbidden:
		return bus.SendMessageResponse{}, ErrSendFailed
	default:
		return bus.SendMessageResponse{}, fmt.Errorf("send message: %s", reply.Header.Get(micro.ErrorHeader))
	}

	var response bus.SendMessageResponse
	if err := json.Unmarshal(reply.Data, &response); err != nil {
		return bus.SendMessageResponse{}, fmt.Errorf("unmarshal send message response: %w", err)
	}

	return response, nil
}
