package streams

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/rs/zerolog/log"
)

const (
	MaxRetries       = 3
	StaleMessageTime = 5 * time.Minute
	BatchSize        = 10
	BlockDuration    = 5 * time.Second
)

// WebhookConsumer processes webhooks from the Redis Stream
type WebhookConsumer struct {
	stream       *WebhookStream
	workflowSvc  *services.WorkflowService
	queueClient  *queue.Client
	consumerName string
	wg           sync.WaitGroup
	stopCh       chan struct{}
}

// NewWebhookConsumer creates a new webhook consumer
func NewWebhookConsumer(
	stream *WebhookStream,
	workflowSvc *services.WorkflowService,
	queueClient *queue.Client,
	consumerName string,
) *WebhookConsumer {
	if consumerName == "" {
		consumerName = "consumer-" + uuid.New().String()[:8]
	}

	return &WebhookConsumer{
		stream:       stream,
		workflowSvc:  workflowSvc,
		queueClient:  queueClient,
		consumerName: consumerName,
		stopCh:       make(chan struct{}),
	}
}

// Start begins consuming webhooks from the stream
func (c *WebhookConsumer) Start(ctx context.Context) error {
	log.Info().
		Str("consumer", c.consumerName).
		Msg("Starting webhook stream consumer")

	// Initialize the stream
	if err := c.stream.Initialize(ctx); err != nil {
		return err
	}

	// Start main consumer loop
	c.wg.Add(1)
	go c.consumeLoop(ctx)

	// Start stale message recovery loop
	c.wg.Add(1)
	go c.recoveryLoop(ctx)

	return nil
}

// Stop gracefully stops the consumer
func (c *WebhookConsumer) Stop() {
	log.Info().Str("consumer", c.consumerName).Msg("Stopping webhook stream consumer")
	close(c.stopCh)
	c.wg.Wait()
}

func (c *WebhookConsumer) consumeLoop(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			c.processBatch(ctx)
		}
	}
}

func (c *WebhookConsumer) processBatch(ctx context.Context) {
	events, messageIDs, err := c.stream.Consume(ctx, c.consumerName, BatchSize, BlockDuration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to consume from webhook stream")
		time.Sleep(time.Second) // Back off on error
		return
	}

	if len(events) == 0 {
		return
	}

	log.Debug().Int("count", len(events)).Msg("Processing webhook batch")

	for i, event := range events {
		err := c.processWebhook(ctx, event)
		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint_id", event.EndpointID).
				Int("retry_count", event.RetryCount).
				Msg("Failed to process webhook")

			// Move to DLQ if max retries exceeded
			if event.RetryCount >= MaxRetries {
				if dlqErr := c.stream.MoveToDeadLetter(ctx, event, messageIDs[i], err.Error()); dlqErr != nil {
					log.Error().Err(dlqErr).Msg("Failed to move to DLQ")
				}
			} else {
				// Re-queue with incremented retry count
				event.RetryCount++
				if _, pubErr := c.stream.Publish(ctx, event); pubErr != nil {
					log.Error().Err(pubErr).Msg("Failed to re-queue webhook")
				}
				// Ack the original
				_ = c.stream.Ack(ctx, messageIDs[i])
			}
			continue
		}

		// Successfully processed - acknowledge
		if ackErr := c.stream.Ack(ctx, messageIDs[i]); ackErr != nil {
			log.Error().Err(ackErr).Msg("Failed to ack message")
		}
	}
}

func (c *WebhookConsumer) processWebhook(ctx context.Context, event WebhookEvent) error {
	// Look up webhook configuration
	webhook, err := c.workflowSvc.GetWebhookByEndpoint(ctx, event.EndpointID)
	if err != nil {
		return err
	}

	// Get workflow
	workflow, err := c.workflowSvc.GetByID(ctx, webhook.WorkflowID)
	if err != nil {
		return err
	}

	// Check if workflow is active
	if workflow.Status != models.WorkflowStatusActive {
		log.Debug().
			Str("workflow_id", workflow.ID.String()).
			Msg("Skipping webhook - workflow not active")
		return nil // Not an error, just skip
	}

	// Build trigger data
	triggerData := models.JSON{
		"method":      event.Method,
		"path":        event.Path,
		"headers":     event.Headers,
		"query":       event.Query,
		"body":        event.Body,
		"contentType": event.ContentType,
		"receivedAt":  event.ReceivedAt.Format(time.RFC3339),
		"eventId":     event.ID,
	}

	// Parse JSON body if applicable
	if strings.Contains(event.ContentType, "application/json") && len(event.Body) > 0 {
		var jsonBody interface{}
		if err := json.Unmarshal([]byte(event.Body), &jsonBody); err == nil {
			triggerData["json"] = jsonBody
		}
	}

	// Queue to Asynq for execution
	payload := queue.WorkflowExecutionPayload{
		WorkflowID:  workflow.ID,
		WorkspaceID: workflow.WorkspaceID,
		TriggerType: "webhook",
		InputData:   triggerData,
	}

	_, err = c.queueClient.EnqueueWorkflowExecution(ctx, payload)
	if err != nil {
		return err
	}

	log.Info().
		Str("endpoint_id", event.EndpointID).
		Str("workflow_id", workflow.ID.String()).
		Str("event_id", event.ID).
		Msg("Webhook queued for execution")

	return nil
}

func (c *WebhookConsumer) recoveryLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.recoverStaleMessages(ctx)
		}
	}
}

func (c *WebhookConsumer) recoverStaleMessages(ctx context.Context) {
	events, messageIDs, err := c.stream.ClaimStale(ctx, c.consumerName, StaleMessageTime, BatchSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to claim stale messages")
		return
	}

	if len(events) == 0 {
		return
	}

	log.Info().Int("count", len(events)).Msg("Recovering stale webhook messages")

	for i, event := range events {
		err := c.processWebhook(ctx, event)
		if err != nil {
			// Move to DLQ - these have already been pending too long
			_ = c.stream.MoveToDeadLetter(ctx, event, messageIDs[i], "stale: "+err.Error())
		} else {
			_ = c.stream.Ack(ctx, messageIDs[i])
		}
	}
}
