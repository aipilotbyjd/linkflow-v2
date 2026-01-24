package websocket

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisSubscriber struct {
	hub    *Hub
	client *redis.Client
}

func NewRedisSubscriber(hub *Hub, client *redis.Client) *RedisSubscriber {
	return &RedisSubscriber{
		hub:    hub,
		client: client,
	}
}

func (s *RedisSubscriber) Start(ctx context.Context) error {
	// Subscribe to all WebSocket patterns
	pubsub := s.client.PSubscribe(ctx, "ws:*")

	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				s.HandlePayload([]byte(msg.Payload))
			}
		}
	}()

	return nil
}

func (s *RedisSubscriber) HandlePayload(payload []byte) {
	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	s.hub.Broadcast(&msg)
}
