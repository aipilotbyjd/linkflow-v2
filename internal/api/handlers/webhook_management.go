package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type WebhookManagementHandler struct {
	webhookMgr *services.WebhookManager
}

func NewWebhookManagementHandler(webhookMgr *services.WebhookManager) *WebhookManagementHandler {
	return &WebhookManagementHandler{webhookMgr: webhookMgr}
}

// GenerateWebhookRequest represents a request to generate a webhook
type GenerateWebhookRequest struct {
	NodeID     string `json:"node_id"`
	Method     string `json:"method,omitempty"`
	CustomPath string `json:"custom_path,omitempty"`
}

// Generate creates a new webhook endpoint
func (h *WebhookManagementHandler) Generate(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	var req GenerateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NodeID == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "node_id required")
		return
	}

	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "workspace required")
		return
	}

	webhook, err := h.webhookMgr.GenerateWebhook(r.Context(), services.GenerateWebhookInput{
		WorkflowID:  workflowID,
		WorkspaceID: wsCtx.WorkspaceID,
		NodeID:      req.NodeID,
		Method:      req.Method,
		CustomPath:  req.CustomPath,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusCreated, webhook)
}

// List returns all webhooks for a workflow
func (h *WebhookManagementHandler) List(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	webhooks, err := h.webhookMgr.GetWebhooksByWorkflow(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"webhooks": webhooks,
		"count":    len(webhooks),
	})
}

// RegenerateSecret generates a new secret for a webhook
func (h *WebhookManagementHandler) RegenerateSecret(w http.ResponseWriter, r *http.Request) {
	webhookIDStr := chi.URLParam(r, "webhookID")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid webhook ID")
		return
	}

	secret, err := h.webhookMgr.RegenerateSecret(r.Context(), webhookID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Secret regenerated",
		"secret":  secret,
	})
}

// Activate activates a webhook
func (h *WebhookManagementHandler) Activate(w http.ResponseWriter, r *http.Request) {
	webhookIDStr := chi.URLParam(r, "webhookID")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid webhook ID")
		return
	}

	if err := h.webhookMgr.ActivateWebhook(r.Context(), webhookID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Webhook activated",
	})
}

// Deactivate deactivates a webhook
func (h *WebhookManagementHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	webhookIDStr := chi.URLParam(r, "webhookID")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid webhook ID")
		return
	}

	if err := h.webhookMgr.DeactivateWebhook(r.Context(), webhookID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Webhook deactivated",
	})
}
