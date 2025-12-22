package streams

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	WebhookStreamKey     = "stream:webhooks"
	WebhookConsumerGroup = "webhook-processors"
	WebhookDLQKey        = "stream:webhooks:dlq" // Dead letter queue
)

// WebhookEvent represents a buffered webhook event
type WebhookEvent struct {
	ID          string            `json:"id"`
	EndpointID  string            `json:"endpoint_id"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Query       map[string]string `json:"query"`
	Body        string            `json:"body"`
	ContentType string            `json:"content_type"`
	ReceivedAt  time.Time         `json:"received_at"`
	RetryCount  int               `json:"retry_count"`
}

// WebhookStream handles Redis Streams operations for webhook buffering
type WebhookStream struct {
	client *redis.Client
}

// NewWebhookStream creates a new webhook stream handler
func NewWebhookStream(client *redis.Client) *WebhookStream {
	return &WebhookStream{client: client}
}

// Initialize creates the consumer group if it doesn't exist
func (s *WebhookStream) Initialize(ctx context.Context) error {
	// Create consumer group, ignore error if it already exists
	err := s.client.XGroupCreateMkStream(ctx, WebhookStreamKey, WebhookConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	log.Info().
		Str("stream", WebhookStreamKey).
		Str("group", WebhookConsumerGroup).
		Msg("Webhook stream initialized")

	return nil
}

// Publish adds a webhook event to the stream (called by webhook handler)
func (s *WebhookStream) Publish(ctx context.Context, event WebhookEvent) (string, error) {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event: %w", err)
	}

	// XADD with MAXLEN to prevent unbounded growth (keep last 100k messages)
	messageID, err := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: WebhookStreamKey,
		MaxLen: 100000,
		Approx: true,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Result()

	if err != nil {
		return "", fmt.Errorf("failed to add to stream: %w", err)
	}

	log.Debug().
		Str("message_id", messageID).
		Str("endpoint_id", event.EndpointID).
		Msg("Webhook buffered to stream")

	return messageID, nil
}

// Consume reads webhook events from the stream (called by consumer worker)
func (s *WebhookStream) Consume(ctx context.Context, consumerName string, count int64, blockDuration time.Duration) ([]WebhookEvent, []string, error) {
	// Read new messages
	streams, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    WebhookConsumerGroup,
		Consumer: consumerName,
		Streams:  []string{WebhookStreamKey, ">"},
		Count:    count,
		Block:    blockDuration,
	}).Result()

	if err == redis.Nil {
		return nil, nil, nil // No messages
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var events []WebhookEvent
	var messageIDs []string

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			data, ok := msg.Values["data"].(string)
			if !ok {
				continue
			}

			var event WebhookEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				log.Error().Err(err).Str("message_id", msg.ID).Msg("Failed to unmarshal webhook event")
				continue
			}

			events = append(events, event)
			messageIDs = append(messageIDs, msg.ID)
		}
	}

	return events, messageIDs, nil
}

// Ack acknowledges successful processing of messages
func (s *WebhookStream) Ack(ctx context.Context, messageIDs ...string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	return s.client.XAck(ctx, WebhookStreamKey, WebhookConsumerGroup, messageIDs...).Err()
}

// MoveToDeadLetter moves a failed message to the dead letter queue
func (s *WebhookStream) MoveToDeadLetter(ctx context.Context, event WebhookEvent, messageID string, reason string) error {
	event.RetryCount++

	data, _ := json.Marshal(map[string]interface{}{
		"event":      event,
		"message_id": messageID,
		"reason":     reason,
		"moved_at":   time.Now(),
	})

	// Add to DLQ
	err := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: WebhookDLQKey,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to add to DLQ: %w", err)
	}

	// Ack original message
	return s.Ack(ctx, messageID)
}

// ClaimStale claims messages that have been pending too long (for recovery)
func (s *WebhookStream) ClaimStale(ctx context.Context, consumerName string, minIdleTime time.Duration, count int64) ([]WebhookEvent, []string, error) {
	// Get pending messages
	pending, err := s.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: WebhookStreamKey,
		Group:  WebhookConsumerGroup,
		Idle:   minIdleTime,
		Start:  "-",
		End:    "+",
		Count:  count,
	}).Result()

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pending: %w", err)
	}

	if len(pending) == 0 {
		return nil, nil, nil
	}

	// Collect message IDs to claim
	ids := make([]string, len(pending))
	for i, p := range pending {
		ids[i] = p.ID
	}

	// Claim the messages
	messages, err := s.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   WebhookStreamKey,
		Group:    WebhookConsumerGroup,
		Consumer: consumerName,
		MinIdle:  minIdleTime,
		Messages: ids,
	}).Result()

	if err != nil {
		return nil, nil, fmt.Errorf("failed to claim messages: %w", err)
	}

	var events []WebhookEvent
	var messageIDs []string

	for _, msg := range messages {
		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var event WebhookEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		events = append(events, event)
		messageIDs = append(messageIDs, msg.ID)
	}

	return events, messageIDs, nil
}

// GetStats returns stream statistics
func (s *WebhookStream) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Stream length
	length, err := s.client.XLen(ctx, WebhookStreamKey).Result()
	if err != nil {
		return nil, err
	}

	// Pending messages
	pending, err := s.client.XPending(ctx, WebhookStreamKey, WebhookConsumerGroup).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// DLQ length
	dlqLength, _ := s.client.XLen(ctx, WebhookDLQKey).Result()

	stats := map[string]interface{}{
		"stream_length":    length,
		"dlq_length":       dlqLength,
		"consumer_group":   WebhookConsumerGroup,
	}

	if pending != nil {
		stats["pending_count"] = pending.Count
		stats["consumers"] = pending.Consumers
	}

	return stats, nil
}

// ReplayFromDLQ replays messages from dead letter queue
func (s *WebhookStream) ReplayFromDLQ(ctx context.Context, count int64) (int, error) {
	messages, err := s.client.XRange(ctx, WebhookDLQKey, "-", "+").Result()
	if err != nil {
		return 0, err
	}

	replayed := 0
	for i, msg := range messages {
		if int64(i) >= count {
			break
		}

		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var dlqEntry struct {
			Event WebhookEvent `json:"event"`
		}
		if err := json.Unmarshal([]byte(data), &dlqEntry); err != nil {
			continue
		}

		// Re-publish to main stream
		_, err := s.Publish(ctx, dlqEntry.Event)
		if err != nil {
			log.Error().Err(err).Msg("Failed to replay webhook")
			continue
		}

		// Remove from DLQ
		s.client.XDel(ctx, WebhookDLQKey, msg.ID)
		replayed++
	}

	return replayed, nil
}

// Trim removes old processed messages from the stream
func (s *WebhookStream) Trim(ctx context.Context, maxLen int64) (int64, error) {
	return s.client.XTrimMaxLen(ctx, WebhookStreamKey, maxLen).Result()
}
