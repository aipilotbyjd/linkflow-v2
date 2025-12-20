package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
)

const (
	TypeWorkflowExecution = "workflow:execution"
	TypeNotification      = "notification:send"
	TypeWebhookDelivery   = "webhook:delivery"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type Client struct {
	client *asynq.Client
}

func NewClient(cfg *config.RedisConfig) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Client{client: client}
}

func (c *Client) Close() error {
	return c.client.Close()
}

// Workflow Execution
type WorkflowExecutionPayload struct {
	WorkflowID  uuid.UUID   `json:"workflow_id"`
	WorkspaceID uuid.UUID   `json:"workspace_id"`
	ExecutionID uuid.UUID   `json:"execution_id,omitempty"`
	TriggeredBy *uuid.UUID  `json:"triggered_by,omitempty"`
	TriggerType string      `json:"trigger_type"`
	TriggerData models.JSON `json:"trigger_data,omitempty"`
	InputData   models.JSON `json:"input_data,omitempty"`
}

func (c *Client) EnqueueWorkflowExecution(ctx context.Context, payload WorkflowExecutionPayload) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeWorkflowExecution, data,
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
		asynq.Retention(24*time.Hour),
	)

	return c.client.EnqueueContext(ctx, task)
}

func (c *Client) EnqueuePriorityWorkflowExecution(ctx context.Context, payload WorkflowExecutionPayload) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeWorkflowExecution, data,
		asynq.Queue(QueueCritical),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
		asynq.Retention(24*time.Hour),
	)

	return c.client.EnqueueContext(ctx, task)
}

func (c *Client) EnqueueDelayedWorkflowExecution(ctx context.Context, payload WorkflowExecutionPayload, delay time.Duration) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeWorkflowExecution, data,
		asynq.Queue(QueueLow),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
		asynq.Retention(24*time.Hour),
		asynq.ProcessIn(delay),
	)

	return c.client.EnqueueContext(ctx, task)
}

// Notification
type NotificationPayload struct {
	Type       string                 `json:"type"` // email, slack, discord, etc.
	Recipient  string                 `json:"recipient"`
	Subject    string                 `json:"subject,omitempty"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

func (c *Client) EnqueueNotification(ctx context.Context, payload NotificationPayload) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeNotification, data,
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Second),
		asynq.Retention(24*time.Hour),
	)

	return c.client.EnqueueContext(ctx, task)
}

// Webhook Delivery
type WebhookDeliveryPayload struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Secret  string            `json:"secret,omitempty"`
}

func (c *Client) EnqueueWebhookDelivery(ctx context.Context, payload WebhookDeliveryPayload) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeWebhookDelivery, data,
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
		asynq.Retention(24*time.Hour),
	)

	return c.client.EnqueueContext(ctx, task)
}

// Scheduled tasks
func (c *Client) EnqueueAt(ctx context.Context, task *asynq.Task, processAt time.Time) (*asynq.TaskInfo, error) {
	return c.client.EnqueueContext(ctx, task, asynq.ProcessAt(processAt))
}

func (c *Client) EnqueueIn(ctx context.Context, task *asynq.Task, delay time.Duration) (*asynq.TaskInfo, error) {
	return c.client.EnqueueContext(ctx, task, asynq.ProcessIn(delay))
}
