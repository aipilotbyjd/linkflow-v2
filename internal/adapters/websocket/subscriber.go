package websocket

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type EventPublisher interface {
	Publish(ctx context.Context, event string, data interface{}) error
	PublishToWorkspace(ctx context.Context, workspaceID uuid.UUID, event string, data interface{}) error
	PublishToUser(ctx context.Context, userID uuid.UUID, event string, data interface{}) error
}

type Subscriber struct {
	hub *Hub
}

func NewSubscriber(hub *Hub) *Subscriber {
	return &Subscriber{hub: hub}
}

func (s *Subscriber) Publish(ctx context.Context, event string, data interface{}) error {
	s.hub.Broadcast(&Message{
		Event: event,
		Data:  data,
	})
	return nil
}

func (s *Subscriber) PublishToWorkspace(ctx context.Context, workspaceID uuid.UUID, event string, data interface{}) error {
	s.hub.BroadcastToWorkspace(workspaceID, event, data)
	return nil
}

func (s *Subscriber) PublishToUser(ctx context.Context, userID uuid.UUID, event string, data interface{}) error {
	s.hub.BroadcastToUser(userID, event, data)
	return nil
}

// RedisSubscriber subscribes to Redis pub/sub for cross-instance communication
type RedisSubscriber struct {
	hub *Hub
}

func NewRedisSubscriber(hub *Hub) *RedisSubscriber {
	return &RedisSubscriber{hub: hub}
}

func (s *RedisSubscriber) Subscribe(ctx context.Context, channel string) error {
	// TODO: Implement Redis subscription
	// This would use Redis pub/sub to receive messages from other instances
	return nil
}

func (s *RedisSubscriber) HandleMessage(channel string, payload []byte) {
	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	s.hub.Broadcast(&msg)
}

// ChannelName returns the Redis channel name for a given scope
func ChannelName(scope string, id uuid.UUID) string {
	if id == uuid.Nil {
		return "ws:" + scope
	}
	return "ws:" + scope + ":" + id.String()
}
