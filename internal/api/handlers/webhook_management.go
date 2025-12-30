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

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
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

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	whID := webhook.ID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/webhooks/" + whID

	dto.NewResponse(webhook).
		Status(http.StatusCreated).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(
			dto.Action{Name: "test", Method: "POST", Href: basePath + "/test", Label: "Test Webhook"},
			dto.Action{Name: "regenerate_secret", Method: "POST", Href: basePath + "/regenerate-secret", Label: "Regenerate Secret"},
			dto.Action{Name: "disable", Method: "POST", Href: basePath + "/disable", Label: "Disable"},
		).
		Send(w)
}

// List returns all webhooks for a workflow
func (h *WebhookManagementHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

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

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/webhooks"

	type WebhookWithActions struct {
		ID       string       `json:"id"`
		Path     string       `json:"path"`
		Method   string       `json:"method"`
		IsActive bool         `json:"is_active"`
		URL      string       `json:"url"`
		Actions  []dto.Action `json:"actions,omitempty"`
	}

	response := make([]WebhookWithActions, len(webhooks))
	for i, wh := range webhooks {
		whID := wh.ID.String()
		whPath := basePath + "/" + whID

		actions := []dto.Action{
			{Name: "regenerate_secret", Method: "POST", Href: whPath + "/regenerate-secret", Label: "Regenerate Secret"},
			{Name: "test", Method: "POST", Href: whPath + "/test", Label: "Test Webhook"},
			{Name: "delete", Method: "DELETE", Href: whPath, Label: "Delete"},
		}
		if wh.IsActive {
			actions = append([]dto.Action{{Name: "disable", Method: "POST", Href: whPath + "/disable", Label: "Disable"}}, actions...)
		} else {
			actions = append([]dto.Action{{Name: "enable", Method: "POST", Href: whPath + "/enable", Label: "Enable"}}, actions...)
		}

		response[i] = WebhookWithActions{
			ID:       whID,
			Path:     wh.Path,
			Method:   wh.Method,
			IsActive: wh.IsActive,
			URL:      wh.URL,
			Actions:  actions,
		}
	}

	data := dto.SelectFields(r, response)

	dto.NewResponse(data).
		WithLinks(&dto.Links{Self: basePath}).
		WithMeta(&dto.Meta{Total: int64(len(webhooks)), Page: 1, PerPage: len(webhooks), TotalPages: 1}).
		Send(w)
}

// RegenerateSecret generates a new secret for a webhook
func (h *WebhookManagementHandler) RegenerateSecret(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowID")
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

	basePath := "/api/v1/workspaces/" + wsCtx.WorkspaceID.String() + "/workflows/" + workflowIDStr + "/webhooks/" + webhookIDStr

	dto.NewResponse(map[string]interface{}{
		"message": "Secret regenerated",
		"secret":  secret,
	}).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

// Activate activates a webhook
func (h *WebhookManagementHandler) Activate(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowID")
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

	basePath := "/api/v1/workspaces/" + wsCtx.WorkspaceID.String() + "/workflows/" + workflowIDStr + "/webhooks/" + webhookIDStr

	dto.NewResponse(map[string]interface{}{
		"message": "Webhook activated",
	}).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(
			dto.Action{Name: "disable", Method: "POST", Href: basePath + "/disable", Label: "Disable"},
		).
		Send(w)
}

// Deactivate deactivates a webhook
func (h *WebhookManagementHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowID")
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

	basePath := "/api/v1/workspaces/" + wsCtx.WorkspaceID.String() + "/workflows/" + workflowIDStr + "/webhooks/" + webhookIDStr

	dto.NewResponse(map[string]interface{}{
		"message": "Webhook deactivated",
	}).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(
			dto.Action{Name: "enable", Method: "POST", Href: basePath + "/enable", Label: "Enable"},
		).
		Send(w)
}
