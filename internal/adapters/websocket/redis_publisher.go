package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}

func (p *RedisPublisher) Publish(ctx context.Context, event string, data interface{}) error {
	msg := &Message{
		Event: event,
		Data:  data,
	}
	return p.publishMessage(ctx, "ws:global", msg)
}

func (p *RedisPublisher) PublishToWorkspace(ctx context.Context, workspaceID uuid.UUID, event string, data interface{}) error {
	msg := &Message{
		WorkspaceID: workspaceID,
		Event:       event,
		Data:        data,
	}
	return p.publishMessage(ctx, fmt.Sprintf("ws:workspace:%s", workspaceID), msg)
}

func (p *RedisPublisher) PublishToUser(ctx context.Context, userID uuid.UUID, event string, data interface{}) error {
	msg := &Message{
		UserID: userID,
		Event:  event,
		Data:   data,
	}
	return p.publishMessage(ctx, fmt.Sprintf("ws:user:%s", userID), msg)
}

func (p *RedisPublisher) publishMessage(ctx context.Context, channel string, msg *Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, channel, payload).Err()
}
