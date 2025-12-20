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
	"github.com/linkflow-ai/linkflow/internal/worker/events"
	"github.com/linkflow-ai/linkflow/internal/worker/executor"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	cfg           *config.Config
	server        *queue.Server
	executor      *executor.Executor
	executionSvc  *services.ExecutionService
	credentialSvc *services.CredentialService
	emailSvc      *email.Service
	publisher     *events.Publisher
}

func New(
	cfg *config.Config,
	executionSvc *services.ExecutionService,
	credentialSvc *services.CredentialService,
	workflowSvc *services.WorkflowService,
	redisClient *redis.Client,
	emailSvc *email.Service,
) *Worker {
	server := queue.NewServer(&cfg.Redis, 10)
	publisher := events.NewPublisher(redisClient)

	exec := executor.NewWithPublisher(executionSvc, credentialSvc, workflowSvc, publisher)

	w := &Worker{
		cfg:           cfg,
		server:        server,
		executor:      exec,
		executionSvc:  executionSvc,
		credentialSvc: credentialSvc,
		emailSvc:      emailSvc,
		publisher:     publisher,
	}

	// Register handlers
	server.HandleFunc(queue.TypeWorkflowExecution, w.handleWorkflowExecution)
	server.HandleFunc(queue.TypeNotification, w.handleNotification)
	server.HandleFunc(queue.TypeWebhookDelivery, w.handleWebhookDelivery)
	server.HandleFunc("email:send", w.handleEmailSend)

	return w
}

func (w *Worker) Start() error {
	log.Info().Msg("Starting worker...")
	return w.server.Start()
}

func (w *Worker) Shutdown() {
	w.server.Shutdown()
}

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
