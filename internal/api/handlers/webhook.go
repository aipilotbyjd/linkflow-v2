package handlers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
)

type WebhookHandler struct {
	workflowSvc  *services.WorkflowService
	executionSvc *services.ExecutionService
	queueClient  *queue.Client
}

func NewWebhookHandler(
	workflowSvc *services.WorkflowService,
	executionSvc *services.ExecutionService,
	queueClient *queue.Client,
) *WebhookHandler {
	return &WebhookHandler{
		workflowSvc:  workflowSvc,
		executionSvc: executionSvc,
		queueClient:  queueClient,
	}
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	endpointID := chi.URLParam(r, "endpointID")
	_ = endpointID

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "failed to read body")
		return
	}

	// TODO: Look up webhook endpoint by ID
	// TODO: Verify webhook signature if secret is set
	// TODO: Queue workflow execution

	triggerData := models.JSON{
		"method":  r.Method,
		"path":    r.URL.Path,
		"headers": headerToMap(r.Header),
		"query":   r.URL.Query(),
		"body":    string(body),
	}

	_ = triggerData

	dto.Accepted(w, map[string]string{
		"status": "accepted",
	})
}

func headerToMap(h http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
