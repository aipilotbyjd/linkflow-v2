package streams

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// WebhookHandler processes webhook events
type WebhookHandler interface {
	HandleWebhook(ctx context.Context, event *WebhookEvent) error
}

// WebhookConsumer consumes webhook events from the stream
type WebhookConsumer struct {
	client       *redis.Client
	stream       string
	group        string
	consumer     string
	handler      WebhookHandler
	batchSize    int64
	blockTimeout time.Duration
	maxRetries   int
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// WebhookConsumerConfig holds consumer configuration
type WebhookConsumerConfig struct {
	ConsumerName string
	BatchSize    int64
	BlockTimeout time.Duration
	MaxRetries   int
}

// DefaultWebhookConsumerConfig returns default consumer configuration
func DefaultWebhookConsumerConfig(consumerName string) WebhookConsumerConfig {
	return WebhookConsumerConfig{
		ConsumerName: consumerName,
		BatchSize:    10,
		BlockTimeout: 5 * time.Second,
		MaxRetries:   3,
	}
}

// NewWebhookConsumer creates a new webhook consumer
func NewWebhookConsumer(client *redis.Client, handler WebhookHandler, cfg WebhookConsumerConfig) *WebhookConsumer {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.BlockTimeout <= 0 {
		cfg.BlockTimeout = 5 * time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}

	return &WebhookConsumer{
		client:       client,
		stream:       WebhookStreamName,
		group:        WebhookGroupName,
		consumer:     cfg.ConsumerName,
		handler:      handler,
		batchSize:    cfg.BatchSize,
		blockTimeout: cfg.BlockTimeout,
		maxRetries:   cfg.MaxRetries,
		stopCh:       make(chan struct{}),
	}
}

// Start starts consuming events
func (c *WebhookConsumer) Start(ctx context.Context) error {
	c.wg.Add(1)
	go c.consume(ctx)

	// Also process pending messages
	c.wg.Add(1)
	go c.processPending(ctx)

	return nil
}

// Stop stops the consumer
func (c *WebhookConsumer) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

func (c *WebhookConsumer) consume(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
			c.readAndProcess(ctx, ">") // ">" means only new messages
		}
	}
}

func (c *WebhookConsumer) processPending(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.claimAndProcess(ctx)
		}
	}
}

func (c *WebhookConsumer) readAndProcess(ctx context.Context, startID string) {
	result, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.group,
		Consumer: c.consumer,
		Streams:  []string{c.stream, startID},
		Count:    c.batchSize,
		Block:    c.blockTimeout,
	}).Result()

	if err != nil {
		if err != redis.Nil {
			// Log error but continue
			time.Sleep(time.Second)
		}
		return
	}

	for _, stream := range result {
		for _, msg := range stream.Messages {
			c.processMessage(ctx, msg)
		}
	}
}

func (c *WebhookConsumer) claimAndProcess(ctx context.Context) {
	// Claim messages that have been pending for more than 1 minute
	minIdleTime := time.Minute

	pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: c.stream,
		Group:  c.group,
		Start:  "-",
		End:    "+",
		Count:  c.batchSize,
	}).Result()

	if err != nil {
		return
	}

	for _, p := range pending {
		if p.Idle < minIdleTime {
			continue
		}

		// Check retry count
		if p.RetryCount >= int64(c.maxRetries) {
			// Move to dead letter or delete
			c.client.XAck(ctx, c.stream, c.group, p.ID)
			continue
		}

		// Claim the message
		claimed, err := c.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   c.stream,
			Group:    c.group,
			Consumer: c.consumer,
			MinIdle:  minIdleTime,
			Messages: []string{p.ID},
		}).Result()

		if err != nil {
			continue
		}

		for _, msg := range claimed {
			c.processMessage(ctx, msg)
		}
	}
}

func (c *WebhookConsumer) processMessage(ctx context.Context, msg redis.XMessage) {
	data, ok := msg.Values["data"].(string)
	if !ok {
		c.client.XAck(ctx, c.stream, c.group, msg.ID)
		return
	}

	var event WebhookEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		c.client.XAck(ctx, c.stream, c.group, msg.ID)
		return
	}
	event.ID = msg.ID

	if err := c.handler.HandleWebhook(ctx, &event); err != nil {
		// Don't ack - will be retried
		return
	}

	// Acknowledge successful processing
	c.client.XAck(ctx, c.stream, c.group, msg.ID)
}

// Pending returns the number of pending messages for this consumer
func (c *WebhookConsumer) Pending(ctx context.Context) (int64, error) {
	info, err := c.client.XPending(ctx, c.stream, c.group).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get pending info: %w", err)
	}
	return info.Count, nil
}
