package worker

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/queue"
	"github.com/linkflow-ai/linkflow/internal/worker/executor"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	cfg          *config.Config
	server       *queue.Server
	executor     *executor.Executor
	executionSvc *services.ExecutionService
	credentialSvc *services.CredentialService
}

func New(
	cfg *config.Config,
	executionSvc *services.ExecutionService,
	credentialSvc *services.CredentialService,
	workflowSvc *services.WorkflowService,
) *Worker {
	server := queue.NewServer(&cfg.Redis, 10)

	exec := executor.New(executionSvc, credentialSvc, workflowSvc)

	w := &Worker{
		cfg:          cfg,
		server:       server,
		executor:     exec,
		executionSvc: executionSvc,
		credentialSvc: credentialSvc,
	}

	// Register handlers
	server.HandleFunc(queue.TypeWorkflowExecution, w.handleWorkflowExecution)
	server.HandleFunc(queue.TypeNotification, w.handleNotification)
	server.HandleFunc(queue.TypeWebhookDelivery, w.handleWebhookDelivery)

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

	// TODO: Implement notification sending based on type
	switch payload.Type {
	case "email":
		// Send email
	case "slack":
		// Send Slack message
	case "discord":
		// Send Discord message
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

	// TODO: Implement webhook delivery

	return nil
}
