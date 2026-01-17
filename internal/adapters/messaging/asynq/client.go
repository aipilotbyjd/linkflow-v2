package asynq

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Task types
const (
	TaskExecuteWorkflow = "workflow:execute"
	TaskSendEmail       = "email:send"
	TaskCleanup         = "system:cleanup"
	TaskTokenRefresh    = "oauth:token_refresh"
)

// Config holds Asynq configuration
type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	UseTLS        bool
	Concurrency   int
	Queues        map[string]int
}

// Client wraps Asynq client
type Client struct {
	client *asynq.Client
}

// NewClient creates a new Asynq client
func NewClient(cfg Config) (*Client, error) {
	opts := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	if cfg.UseTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := asynq.NewClient(opts)
	return &Client{client: client}, nil
}

// Close closes the client
func (c *Client) Close() error {
	return c.client.Close()
}

// Enqueue enqueues a task
func (c *Client) Enqueue(ctx context.Context, taskType string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(taskType, data, opts...)
	return c.client.EnqueueContext(ctx, task)
}

// EnqueueIn enqueues a task to run after a delay
func (c *Client) EnqueueIn(ctx context.Context, taskType string, payload interface{}, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.ProcessIn(delay))
	return c.Enqueue(ctx, taskType, payload, opts...)
}

// EnqueueAt enqueues a task to run at a specific time
func (c *Client) EnqueueAt(ctx context.Context, taskType string, payload interface{}, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.ProcessAt(at))
	return c.Enqueue(ctx, taskType, payload, opts...)
}

// ExecuteWorkflowPayload represents the payload for workflow execution
type ExecuteWorkflowPayload struct {
	ExecutionID string                 `json:"execution_id"`
	WorkflowID  string                 `json:"workflow_id"`
	WorkspaceID string                 `json:"workspace_id"`
	TriggerType string                 `json:"trigger_type"`
	TriggerData map[string]interface{} `json:"trigger_data,omitempty"`
	InputData   map[string]interface{} `json:"input_data,omitempty"`
}

// EnqueueWorkflowExecution enqueues a workflow execution task
func (c *Client) EnqueueWorkflowExecution(ctx context.Context, payload ExecuteWorkflowPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.Queue("executions"))
	return c.Enqueue(ctx, TaskExecuteWorkflow, payload, opts...)
}

// SendEmailPayload represents the payload for sending email
type SendEmailPayload struct {
	To       string                 `json:"to"`
	Subject  string                 `json:"subject"`
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// EnqueueEmail enqueues an email task
func (c *Client) EnqueueEmail(ctx context.Context, payload SendEmailPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.Queue("emails"))
	return c.Enqueue(ctx, TaskSendEmail, payload, opts...)
}

// Server wraps Asynq server
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewServer creates a new Asynq server
func NewServer(cfg Config) (*Server, error) {
	opts := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	if cfg.UseTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	queues := cfg.Queues
	if queues == nil {
		queues = map[string]int{
			"critical":   6,
			"executions": 4,
			"default":    3,
			"emails":     2,
			"low":        1,
		}
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	server := asynq.NewServer(opts, asynq.Config{
		Concurrency: concurrency,
		Queues:      queues,
	})

	return &Server{
		server: server,
		mux:    asynq.NewServeMux(),
	}, nil
}

// HandleFunc registers a handler function for a task type
func (s *Server) HandleFunc(taskType string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(taskType, handler)
}

// Start starts the server
func (s *Server) Start() error {
	return s.server.Start(s.mux)
}

// Shutdown shuts down the server
func (s *Server) Shutdown() {
	s.server.Shutdown()
}
