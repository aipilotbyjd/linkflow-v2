package websocket

import (
	"context"

	"github.com/google/uuid"
)

// EventPublisher defines the interface for publishing real-time events
type EventPublisher interface {
	Publish(ctx context.Context, event string, data interface{}) error
	PublishToWorkspace(ctx context.Context, workspaceID uuid.UUID, event string, data interface{}) error
	PublishToUser(ctx context.Context, userID uuid.UUID, event string, data interface{}) error
}

// Subscriber is a local adapter that publishes events directly to the Hub
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

// ChannelName returns the Redis channel name for a given scope
func ChannelName(scope string, id uuid.UUID) string {
	if id == uuid.Nil {
		return "ws:" + scope
	}
	return "ws:" + scope + ":" + id.String()
}
