package alerts

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// UpdateHandler handles updating usage alerts
type UpdateHandler struct {
	repo billing.UsageAlertRepository
}

func NewUpdateHandler(repo billing.UsageAlertRepository) *UpdateHandler {
	return &UpdateHandler{repo: repo}
}

func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)
	
	alertID, err := uuid.Parse(chi.URLParam(r, "alertId"))
	if err != nil {
		common.BadRequest(w, "Invalid alert ID")
		return
	}

	var req UpdateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	alert, err := h.repo.FindByID(ctx, alertID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if alert == nil {
		common.NotFound(w, "Alert")
		return
	}
	if alert.WorkspaceID != workspaceID {
		common.Forbidden(w, "You don't have permission to update this alert")
		return
	}

	// Update fields
	if req.Enabled != nil {
		alert.Enabled = *req.Enabled
	}

	if req.Channels != nil {
		alert.Channels = billing.AlertChannels{
			Email:        req.Channels.Email,
			EmailAddrs:   req.Channels.EmailAddrs,
			Slack:        req.Channels.Slack,
			SlackWebhook: req.Channels.SlackWebhook,
			Webhook:      req.Channels.Webhook,
			WebhookURL:   req.Channels.WebhookURL,
			InApp:        req.Channels.InApp,
			SMS:          req.Channels.SMS,
			PhoneNumbers: req.Channels.PhoneNumbers,
		}
	}

	alert.UpdatedAt = time.Now()

	if err := h.repo.Update(ctx, alert); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToAlertResponse(alert))
}
