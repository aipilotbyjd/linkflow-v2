package streams

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	WebhookStreamName = "linkflow:webhooks"
	WebhookGroupName  = "webhook-processors"
)

// WebhookEvent represents an incoming webhook event
type WebhookEvent struct {
	ID         string                 `json:"id"`
	EndpointID string                 `json:"endpoint_id"`
	WorkflowID string                 `json:"workflow_id"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	Headers    map[string]string      `json:"headers"`
	Body       []byte                 `json:"body"`
	Query      map[string]string      `json:"query"`
	RemoteAddr string                 `json:"remote_addr"`
	ReceivedAt time.Time              `json:"received_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookStream handles webhook event streaming via Redis Streams
type WebhookStream struct {
	client    *redis.Client
	stream    string
	group     string
	maxLen    int64
	batchSize int64
}

// WebhookStreamConfig holds configuration for the webhook stream
type WebhookStreamConfig struct {
	MaxLen    int64
	BatchSize int64
}

// DefaultWebhookStreamConfig returns default configuration
func DefaultWebhookStreamConfig() WebhookStreamConfig {
	return WebhookStreamConfig{
		MaxLen:    100000,
		BatchSize: 10,
	}
}

// NewWebhookStream creates a new webhook stream
func NewWebhookStream(client *redis.Client, cfg WebhookStreamConfig) *WebhookStream {
	if cfg.MaxLen <= 0 {
		cfg.MaxLen = 100000
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}

	return &WebhookStream{
		client:    client,
		stream:    WebhookStreamName,
		group:     WebhookGroupName,
		maxLen:    cfg.MaxLen,
		batchSize: cfg.BatchSize,
	}
}

// Initialize creates the stream and consumer group if they don't exist
func (s *WebhookStream) Initialize(ctx context.Context) error {
	// Create consumer group (creates stream if it doesn't exist)
	err := s.client.XGroupCreateMkStream(ctx, s.stream, s.group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	return nil
}

// Publish publishes a webhook event to the stream
func (s *WebhookStream) Publish(ctx context.Context, event *WebhookEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event: %w", err)
	}

	result, err := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: s.stream,
		MaxLen: s.maxLen,
		Approx: true,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("failed to publish event: %w", err)
	}

	return result, nil
}

// Read reads events from the stream (for non-consumer-group usage)
func (s *WebhookStream) Read(ctx context.Context, lastID string, count int64) ([]*WebhookEvent, string, error) {
	if lastID == "" {
		lastID = "0"
	}
	if count <= 0 {
		count = s.batchSize
	}

	result, err := s.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{s.stream, lastID},
		Count:   count,
		Block:   time.Second,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, lastID, nil
		}
		return nil, "", fmt.Errorf("failed to read from stream: %w", err)
	}

	if len(result) == 0 || len(result[0].Messages) == 0 {
		return nil, lastID, nil
	}

	events := make([]*WebhookEvent, 0, len(result[0].Messages))
	var newLastID string

	for _, msg := range result[0].Messages {
		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var event WebhookEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		event.ID = msg.ID
		events = append(events, &event)
		newLastID = msg.ID
	}

	return events, newLastID, nil
}

// Len returns the number of events in the stream
func (s *WebhookStream) Len(ctx context.Context) (int64, error) {
	return s.client.XLen(ctx, s.stream).Result()
}

// Trim trims the stream to the specified max length
func (s *WebhookStream) Trim(ctx context.Context, maxLen int64) (int64, error) {
	return s.client.XTrimMaxLen(ctx, s.stream, maxLen).Result()
}

// Delete deletes specific messages from the stream
func (s *WebhookStream) Delete(ctx context.Context, ids ...string) (int64, error) {
	return s.client.XDel(ctx, s.stream, ids...).Result()
}

// Info returns information about the stream
func (s *WebhookStream) Info(ctx context.Context) (*redis.XInfoStream, error) {
	return s.client.XInfoStream(ctx, s.stream).Result()
}

// GroupInfo returns information about consumer groups
func (s *WebhookStream) GroupInfo(ctx context.Context) ([]redis.XInfoGroup, error) {
	return s.client.XInfoGroups(ctx, s.stream).Result()
}
