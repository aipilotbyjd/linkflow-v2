package streams

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Event represents a generic domain event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventStream handles domain event streaming via Redis Streams
type EventStream struct {
	client *redis.Client
	prefix string
	maxLen int64
}

// EventStreamConfig holds configuration
type EventStreamConfig struct {
	Prefix string
	MaxLen int64
}

// NewEventStream creates a new event stream
func NewEventStream(client *redis.Client, cfg EventStreamConfig) *EventStream {
	if cfg.Prefix == "" {
		cfg.Prefix = "linkflow:events:"
	}
	if cfg.MaxLen <= 0 {
		cfg.MaxLen = 10000
	}

	return &EventStream{
		client: client,
		prefix: cfg.Prefix,
		maxLen: cfg.MaxLen,
	}
}

// Publish publishes an event to a stream
func (s *EventStream) Publish(ctx context.Context, streamName string, event *Event) (string, error) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event: %w", err)
	}

	result, err := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: s.prefix + streamName,
		MaxLen: s.maxLen,
		Approx: true,
		Values: map[string]interface{}{
			"type": event.Type,
			"data": string(data),
		},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("failed to publish event: %w", err)
	}

	return result, nil
}

// PublishWorkflowEvent publishes a workflow-related event
func (s *EventStream) PublishWorkflowEvent(ctx context.Context, workflowID string, eventType string, data map[string]interface{}) (string, error) {
	event := &Event{
		Type:      eventType,
		Source:    "workflow:" + workflowID,
		Data:      data,
		Timestamp: time.Now(),
	}
	return s.Publish(ctx, "workflow:"+workflowID, event)
}

// PublishExecutionEvent publishes an execution-related event
func (s *EventStream) PublishExecutionEvent(ctx context.Context, executionID string, eventType string, data map[string]interface{}) (string, error) {
	event := &Event{
		Type:      eventType,
		Source:    "execution:" + executionID,
		Data:      data,
		Timestamp: time.Now(),
	}
	return s.Publish(ctx, "execution:"+executionID, event)
}

// PublishWorkspaceEvent publishes a workspace-related event
func (s *EventStream) PublishWorkspaceEvent(ctx context.Context, workspaceID string, eventType string, data map[string]interface{}) (string, error) {
	event := &Event{
		Type:      eventType,
		Source:    "workspace:" + workspaceID,
		Data:      data,
		Timestamp: time.Now(),
	}
	return s.Publish(ctx, "workspace:"+workspaceID, event)
}

// Subscribe subscribes to events from a stream
func (s *EventStream) Subscribe(ctx context.Context, streamName string, lastID string, handler func(*Event) error) error {
	if lastID == "" {
		lastID = "$" // Only new messages
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			result, err := s.client.XRead(ctx, &redis.XReadArgs{
				Streams: []string{s.prefix + streamName, lastID},
				Count:   10,
				Block:   5 * time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					continue
				}
				return fmt.Errorf("failed to read from stream: %w", err)
			}

			for _, stream := range result {
				for _, msg := range stream.Messages {
					data, ok := msg.Values["data"].(string)
					if !ok {
						continue
					}

					var event Event
					if err := json.Unmarshal([]byte(data), &event); err != nil {
						continue
					}
					event.ID = msg.ID

					if err := handler(&event); err != nil {
						// Log error but continue processing
					}

					lastID = msg.ID
				}
			}
		}
	}
}

// Read reads events from a stream without blocking
func (s *EventStream) Read(ctx context.Context, streamName string, startID string, count int64) ([]*Event, error) {
	if startID == "" {
		startID = "0"
	}

	result, err := s.client.XRange(ctx, s.prefix+streamName, startID, "+").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	events := make([]*Event, 0, len(result))
	for _, msg := range result {
		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var event Event
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		event.ID = msg.ID
		events = append(events, &event)
	}

	return events, nil
}

// ReadLast reads the last N events from a stream
func (s *EventStream) ReadLast(ctx context.Context, streamName string, count int64) ([]*Event, error) {
	if count <= 0 {
		count = 10
	}

	result, err := s.client.XRevRange(ctx, s.prefix+streamName, "+", "-").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	// Limit results
	if int64(len(result)) > count {
		result = result[:count]
	}

	// Reverse to get chronological order
	events := make([]*Event, 0, len(result))
	for i := len(result) - 1; i >= 0; i-- {
		msg := result[i]
		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var event Event
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		event.ID = msg.ID
		events = append(events, &event)
	}

	return events, nil
}

// Len returns the number of events in a stream
func (s *EventStream) Len(ctx context.Context, streamName string) (int64, error) {
	return s.client.XLen(ctx, s.prefix+streamName).Result()
}

// Trim trims a stream to the specified max length
func (s *EventStream) Trim(ctx context.Context, streamName string, maxLen int64) (int64, error) {
	return s.client.XTrimMaxLen(ctx, s.prefix+streamName, maxLen).Result()
}

// Delete deletes a stream
func (s *EventStream) Delete(ctx context.Context, streamName string) error {
	return s.client.Del(ctx, s.prefix+streamName).Err()
}
