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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/linkflow-ai/linkflow/internal/pkg/streams"
	"github.com/rs/zerolog/log"
)

type WebhookHandler struct {
	workflowSvc   *services.WorkflowService
	executionSvc  *services.ExecutionService
	queueClient   *queue.Client
	webhookStream *streams.WebhookStream // Redis Streams buffer
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

// SetWebhookStream enables Redis Streams buffering for webhooks
func (h *WebhookHandler) SetWebhookStream(stream *streams.WebhookStream) {
	h.webhookStream = stream
}

// MaxWebhookBodySize is the maximum allowed size for webhook request bodies (5MB)
const MaxWebhookBodySize = 5 * 1024 * 1024

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	endpointID := chi.URLParam(r, "endpointID")

	// SECURITY: Limit request body size to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, MaxWebhookBodySize)

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			dto.ErrorResponse(w, http.StatusRequestEntityTooLarge, "request body too large (max 5MB)")
			return
		}
		dto.ErrorResponse(w, http.StatusBadRequest, "failed to read body")
		return
	}

	// If Redis Streams buffering is enabled, use fast path
	if h.webhookStream != nil {
		h.handleWithStream(w, r, endpointID, body)
		return
	}

	// Legacy direct processing path
	h.handleDirect(w, r, ctx, endpointID, body)
}

// handleWithStream buffers the webhook to Redis Streams for durable processing
// This is the fast path - accepts webhook immediately, processes asynchronously
func (h *WebhookHandler) handleWithStream(w http.ResponseWriter, r *http.Request, endpointID string, body []byte) {
	ctx := r.Context()

	// Quick validation - just check endpoint exists
	webhook, err := h.workflowSvc.GetWebhookByEndpoint(ctx, endpointID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "webhook endpoint not found")
		return
	}

	// Verify signature if required (must do before buffering)
	if webhook.Secret != "" {
		if !h.verifySignature(r, body, webhook.Secret) {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid signature")
			return
		}
	}

	// Buffer to Redis Stream - this is fast and durable
	event := streams.WebhookEvent{
		EndpointID:  endpointID,
		Method:      r.Method,
		Path:        r.URL.Path,
		Headers:     headerToMap(r.Header),
		Query:       queryToMap(r.URL.Query()),
		Body:        string(body),
		ContentType: r.Header.Get("Content-Type"),
		ReceivedAt:  time.Now(),
	}

	messageID, err := h.webhookStream.Publish(ctx, event)
	if err != nil {
		log.Error().Err(err).Str("endpoint_id", endpointID).Msg("Failed to buffer webhook")
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to accept webhook")
		return
	}

	log.Debug().
		Str("endpoint_id", endpointID).
		Str("message_id", messageID).
		Msg("Webhook buffered successfully")

	// For sync webhooks that need to wait, fall back to direct processing
	if webhook.ResponseMode == "wait" {
		workflow, err := h.workflowSvc.GetByID(ctx, webhook.WorkflowID)
		if err != nil || workflow.Status != models.WorkflowStatusActive {
			dto.ErrorResponse(w, http.StatusNotFound, "workflow not active")
			return
		}

		triggerData := h.buildTriggerData(r, body)
		result, err := h.waitForExecution(ctx, workflow, triggerData, webhook.ResponseTimeout)
		if err != nil {
			dto.ErrorResponse(w, http.StatusInternalServerError, "execution failed: "+err.Error())
			return
		}
		dto.JSON(w, http.StatusOK, result)
		return
	}

	// Async webhook - return immediately
	dto.Accepted(w, map[string]interface{}{
		"status":     "accepted",
		"message":    "Webhook received and queued",
		"message_id": messageID,
	})
}

// handleDirect processes webhook directly without buffering (legacy path)
func (h *WebhookHandler) handleDirect(w http.ResponseWriter, r *http.Request, ctx context.Context, endpointID string, body []byte) {
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

	triggerData := h.buildTriggerData(r, body)

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
		result, err := h.waitForExecution(ctx, workflow, triggerData, webhook.ResponseTimeout)
		if err != nil {
			log.Error().Err(err).Msg("Execution failed or timed out")
			dto.ErrorResponse(w, http.StatusInternalServerError, "execution failed: "+err.Error())
			return
		}
		dto.JSON(w, http.StatusOK, result)
	} else {
		dto.Accepted(w, map[string]interface{}{
			"status":  "accepted",
			"message": "Webhook received",
		})
	}
}

func (h *WebhookHandler) buildTriggerData(r *http.Request, body []byte) models.JSON {
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

	return triggerData
}

func (h *WebhookHandler) waitForExecution(ctx context.Context, workflow *models.Workflow, triggerData models.JSON, timeout int) (map[string]interface{}, error) {
	if timeout <= 0 || timeout > 30 {
		timeout = 30 // Default 30 second timeout for sync webhooks
	}

	// Create execution directly for sync waiting
	execution, err := h.executionSvc.Create(ctx, services.CreateExecutionInput{
		WorkflowID:  workflow.ID,
		WorkspaceID: workflow.WorkspaceID,
		TriggerType: "webhook",
		TriggerData: triggerData,
	})
	if err != nil {
		return nil, err
	}

	// Queue and wait
	payload := queue.WorkflowExecutionPayload{
		WorkflowID:  workflow.ID,
		WorkspaceID: workflow.WorkspaceID,
		ExecutionID: execution.ID,
		TriggerType: "webhook",
		InputData:   triggerData,
	}
	_, _ = h.queueClient.EnqueueWorkflowExecution(ctx, payload)

	// Poll for completion
	ticker := NewPollTicker(100, timeout*1000)
	for ticker.Next() {
		exec, err := h.executionSvc.GetByID(ctx, execution.ID)
		if err != nil {
			continue
		}
		
		switch exec.Status {
		case models.ExecutionStatusCompleted:
			// Return the output from the last node or respond node
			output := make(map[string]interface{})
			if exec.OutputData != nil {
				output["data"] = exec.OutputData
			}
			output["execution_id"] = exec.ID.String()
			output["status"] = "completed"
			return output, nil
			
		case models.ExecutionStatusFailed, models.ExecutionStatusCancelled:
			errMsg := exec.Status
			if exec.ErrorMessage != nil {
				errMsg = *exec.ErrorMessage
			}
			return nil, executionError(exec.Status, errMsg)
		}
	}

	return nil, executionError("timeout", "execution timed out")
}

type PollTicker struct {
	intervalMs int
	maxMs      int
	elapsed    int
}

func NewPollTicker(intervalMs, maxMs int) *PollTicker {
	return &PollTicker{intervalMs: intervalMs, maxMs: maxMs}
}

func (p *PollTicker) Next() bool {
	if p.elapsed >= p.maxMs {
		return false
	}
	if p.elapsed > 0 {
		<-time.After(time.Duration(p.intervalMs) * time.Millisecond)
	}
	p.elapsed += p.intervalMs
	return true
}

func executionError(status, msg string) error {
	if msg == "" {
		msg = status
	}
	return &ExecutionError{Status: status, Message: msg}
}

type ExecutionError struct {
	Status  string
	Message string
}

func (e *ExecutionError) Error() string {
	return e.Message
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
