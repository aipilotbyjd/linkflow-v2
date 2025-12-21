package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/email"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/linkflow-ai/linkflow/internal/worker/cache"
	"github.com/linkflow-ai/linkflow/internal/worker/events"
	"github.com/linkflow-ai/linkflow/internal/worker/executor"
	"github.com/linkflow-ai/linkflow/internal/worker/middleware"
	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
	"github.com/linkflow-ai/linkflow/internal/worker/processor"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Worker handles background job processing
type Worker struct {
	cfg          *config.Config
	server       *queue.Server
	processor    *processor.Processor
	executor     *executor.Executor
	cancellation *processor.CancellationManager
	emailSvc     *email.Service
	publisher    *events.Publisher
	metrics      *middleware.MetricsCollector
	redisClient  *redis.Client
}

// Dependencies holds all external dependencies for the worker
type Dependencies struct {
	ExecutionSvc  *services.ExecutionService
	CredentialSvc *services.CredentialService
	WorkflowSvc   *services.WorkflowService
	RedisClient   *redis.Client
	EmailSvc      *email.Service
	QueueClient   *queue.Client
}

// New creates a new worker with the scalable architecture
func New(
	cfg *config.Config,
	executionSvc *services.ExecutionService,
	credentialSvc *services.CredentialService,
	workflowSvc *services.WorkflowService,
	redisClient *redis.Client,
	emailSvc *email.Service,
) *Worker {
	// Create queue server
	server := queue.NewServer(&cfg.Redis, 10)

	// Create event publisher
	publisher := events.NewPublisher(redisClient)

	// Create queue client for node dependencies
	queueClient := queue.NewClient(&cfg.Redis)

	// Set global dependencies for nodes that need them
	nodes.SetGlobalDependencies(&nodes.Dependencies{
		QueueClient: queueClient,
		RedisClient: redisClient,
	})

	// Create middleware chain
	middlewareChain := middleware.NewChain(
		middleware.NewRecoveryMiddleware(middleware.RecoveryConfig{
			LogStackTrace: cfg.App.Debug,
		}),
		middleware.NewLoggingMiddleware(middleware.LoggingOptions{
			LogInput:    cfg.App.Debug,
			LogOutput:   cfg.App.Debug,
			LogDuration: true,
		}),
		middleware.NewMetricsMiddleware(),
		middleware.NewTimeoutMiddleware(middleware.TimeoutConfig{
			DefaultTimeout: 5 * time.Minute,
			NodeTimeouts:   middleware.DefaultTimeouts(),
		}),
		middleware.NewRateLimitMiddleware(middleware.DefaultRateLimitConfig()),
	)

	// Create result cache
	resultCache := cache.NewResultCache(redisClient, cache.ResultCacheConfig{
		TTL: 1 * time.Hour,
	})

	// Create metrics collector
	metricsCollector := middleware.NewMetricsCollector()

	// Create processor
	proc := processor.New(processor.Config{
		Middleware: middlewareChain,
		Cache:      resultCache,
		Metrics:    metricsCollector,
	})

	// Create cancellation manager
	cancellation := processor.NewCancellationManager(redisClient)

	// Create credential cache
	credCache := cache.NewCredentialCache(redisClient, cache.CredentialCacheConfig{
		RedisTTL:  5 * time.Minute,
		MemoryTTL: 1 * time.Minute,
	})

	// Create executor
	exec := executor.New(
		proc,
		executionSvc,
		credentialSvc,
		workflowSvc,
		publisher,
		cancellation,
		credCache,
		redisClient,
	)

	w := &Worker{
		cfg:          cfg,
		server:       server,
		processor:    proc,
		executor:     exec,
		cancellation: cancellation,
		emailSvc:     emailSvc,
		publisher:    publisher,
		metrics:      metricsCollector,
		redisClient:  redisClient,
	}

	// Register handlers
	server.HandleFunc(queue.TypeWorkflowExecution, w.handleWorkflowExecution)
	server.HandleFunc(queue.TypeNotification, w.handleNotification)
	server.HandleFunc(queue.TypeWebhookDelivery, w.handleWebhookDelivery)
	server.HandleFunc("email:send", w.handleEmailSend)
	server.HandleFunc("workflow:cancel", w.handleWorkflowCancel)

	return w
}

// Start starts the worker
func (w *Worker) Start() error {
	log.Info().Msg("Starting worker with scalable architecture...")

	// Start cancellation listener
	go w.cancellation.Listen(context.Background())

	// Start credential cache cleanup
	ctx := context.Background()
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Cleanup handled by cache internally
			}
		}
	}()

	return w.server.Run()
}

// Shutdown gracefully shuts down the worker
func (w *Worker) Shutdown() {
	log.Info().Msg("Shutting down worker...")
	w.server.Shutdown()
}

// GetExecutor returns the executor for API access
func (w *Worker) GetExecutor() *executor.Executor {
	return w.executor
}

// GetCancellationManager returns the cancellation manager
func (w *Worker) GetCancellationManager() *processor.CancellationManager {
	return w.cancellation
}

// Handler implementations

func (w *Worker) handleWorkflowExecution(ctx context.Context, task *asynq.Task) error {
	var payload queue.WorkflowExecutionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Info().
		Str("workflow_id", payload.WorkflowID.String()).
		Str("workspace_id", payload.WorkspaceID.String()).
		Msg("Processing workflow execution")

	return w.executor.Execute(context.Background(), payload)
}

func (w *Worker) handleNotification(ctx context.Context, task *asynq.Task) error {
	var payload queue.NotificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Info().
		Str("type", payload.Type).
		Str("recipient", payload.Recipient).
		Msg("Sending notification")

	switch payload.Type {
	case "email":
		if w.emailSvc != nil {
			return w.emailSvc.Send(ctx, &email.Email{
				To:      []string{payload.Recipient},
				Subject: payload.Subject,
				Body:    payload.Message,
			})
		}
	case "webhook":
		return w.deliverWebhook(ctx, payload.Recipient, payload.Data)
	}

	return nil
}

func (w *Worker) handleWebhookDelivery(ctx context.Context, task *asynq.Task) error {
	var payload queue.WebhookDeliveryPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Info().
		Str("url", payload.URL).
		Str("method", payload.Method).
		Msg("Delivering webhook")

	return w.deliverWebhookRequest(ctx, payload)
}

func (w *Worker) handleEmailSend(ctx context.Context, task *asynq.Task) error {
	if w.emailSvc == nil {
		log.Warn().Msg("Email service not configured")
		return nil
	}

	var emailData email.Email
	if err := json.Unmarshal(task.Payload(), &emailData); err != nil {
		return err
	}

	return w.emailSvc.Send(ctx, &emailData)
}

func (w *Worker) handleWorkflowCancel(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		ExecutionID string `json:"execution_id"`
		Reason      string `json:"reason"`
		RequestedBy string `json:"requested_by"`
	}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Info().
		Str("execution_id", payload.ExecutionID).
		Str("reason", payload.Reason).
		Msg("Processing workflow cancellation")

	// The cancellation manager handles this via Redis pubsub
	// This handler is for explicit queue-based cancellation
	return nil
}

func (w *Worker) deliverWebhook(ctx context.Context, url string, data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Warn().
			Str("url", url).
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("Webhook delivery failed")
	}

	return nil
}

func (w *Worker) deliverWebhookRequest(ctx context.Context, payload queue.WebhookDeliveryPayload) error {
	var body io.Reader
	if payload.Body != "" {
		body = bytes.NewReader([]byte(payload.Body))
	}

	req, err := http.NewRequestWithContext(ctx, payload.Method, payload.URL, body)
	if err != nil {
		return err
	}

	for k, v := range payload.Headers {
		req.Header.Set(k, v)
	}

	if payload.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("url", payload.URL).Msg("Webhook delivery failed")
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	log.Info().
		Str("url", payload.URL).
		Int("status", resp.StatusCode).
		Int("response_size", len(respBody)).
		Msg("Webhook delivered")

	return nil
}
