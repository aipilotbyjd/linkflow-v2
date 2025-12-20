package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/rs/zerolog/log"
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
	ctx := r.Context()
	endpointID := chi.URLParam(r, "endpointID")

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "failed to read body")
		return
	}

	// Look up webhook by endpoint ID
	webhook, err := h.workflowSvc.GetWebhookByEndpoint(ctx, endpointID)
	if err != nil {
		log.Warn().Str("endpoint_id", endpointID).Msg("Webhook endpoint not found")
		dto.ErrorResponse(w, http.StatusNotFound, "webhook endpoint not found")
		return
	}

	// Verify webhook signature if secret is set
	if webhook.Secret != "" {
		if !h.verifySignature(r, body, webhook.Secret) {
			log.Warn().Str("endpoint_id", endpointID).Msg("Invalid webhook signature")
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid signature")
			return
		}
	}

	// Check if workflow is active
	workflow, err := h.workflowSvc.GetByID(ctx, webhook.WorkflowID)
	if err != nil || workflow.Status != models.WorkflowStatusActive {
		log.Warn().
			Str("endpoint_id", endpointID).
			Str("workflow_id", webhook.WorkflowID.String()).
			Msg("Workflow not found or inactive")
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not active")
		return
	}

	// Build trigger data
	triggerData := models.JSON{
		"method":      r.Method,
		"path":        r.URL.Path,
		"headers":     headerToMap(r.Header),
		"query":       queryToMap(r.URL.Query()),
		"body":        string(body),
		"contentType": r.Header.Get("Content-Type"),
	}

	// Parse JSON body if applicable
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") && len(body) > 0 {
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			triggerData["json"] = jsonBody
		}
	}

	// Queue workflow execution
	err = h.queueWorkflowExecution(ctx, workflow, triggerData)
	if err != nil {
		log.Error().Err(err).Str("workflow_id", workflow.ID.String()).Msg("Failed to queue workflow")
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to queue execution")
		return
	}

	log.Info().
		Str("endpoint_id", endpointID).
		Str("workflow_id", workflow.ID.String()).
		Msg("Webhook triggered workflow execution")

	// Return response based on webhook config
	if webhook.ResponseMode == "wait" {
		// TODO: Wait for execution to complete and return result
		dto.Accepted(w, map[string]interface{}{
			"status":  "processing",
			"message": "Workflow execution queued",
		})
	} else {
		dto.Accepted(w, map[string]interface{}{
			"status":  "accepted",
			"message": "Webhook received",
		})
	}
}

func (h *WebhookHandler) verifySignature(r *http.Request, body []byte, secret string) bool {
	// Check common signature headers
	signatures := map[string]string{
		"X-Hub-Signature-256":   "sha256=",  // GitHub
		"X-Signature":           "",         // Generic
		"Stripe-Signature":      "v1=",      // Stripe (simplified)
		"X-Slack-Signature":     "v0=",      // Slack
		"X-Twilio-Signature":    "",         // Twilio
	}

	for header, prefix := range signatures {
		sig := r.Header.Get(header)
		if sig == "" {
			continue
		}

		// Remove prefix if present
		sig = strings.TrimPrefix(sig, prefix)

		// Compute expected signature
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))

		if hmac.Equal([]byte(sig), []byte(expected)) {
			return true
		}
	}

	// If no signature header found, allow if no signature required
	for header := range signatures {
		if r.Header.Get(header) != "" {
			return false
		}
	}

	return true
}

func (h *WebhookHandler) queueWorkflowExecution(ctx context.Context, workflow *models.Workflow, triggerData models.JSON) error {
	payload := queue.WorkflowExecutionPayload{
		WorkflowID:  workflow.ID,
		WorkspaceID: workflow.WorkspaceID,
		TriggerType: "webhook",
		InputData:   triggerData,
	}

	_, err := h.queueClient.EnqueueWorkflowExecution(ctx, payload)
	return err
}

// HandleTest handles test webhook requests (for testing during workflow creation)
func (h *WebhookHandler) HandleTest(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")

	wfID, err := uuid.Parse(workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "failed to read body")
		return
	}

	triggerData := models.JSON{
		"method":      r.Method,
		"path":        r.URL.Path,
		"headers":     headerToMap(r.Header),
		"query":       queryToMap(r.URL.Query()),
		"body":        string(body),
		"contentType": r.Header.Get("Content-Type"),
		"isTest":      true,
	}

	// Parse JSON body
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") && len(body) > 0 {
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			triggerData["json"] = jsonBody
		}
	}

	// Get workflow
	workflow, err := h.workflowSvc.GetByID(r.Context(), wfID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}

	// Queue execution
	err = h.queueWorkflowExecution(r.Context(), workflow, triggerData)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to queue execution")
		return
	}

	dto.Accepted(w, map[string]interface{}{
		"status":  "accepted",
		"message": "Test webhook received, execution queued",
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

func queryToMap(q map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range q {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
