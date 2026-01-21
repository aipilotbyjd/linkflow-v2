package alerts

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// CreateHandler handles creating usage alerts
type CreateHandler struct {
	repo billing.UsageAlertRepository
}

func NewCreateHandler(repo billing.UsageAlertRepository) *CreateHandler {
	return &CreateHandler{repo: repo}
}

func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	var req CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// Validate threshold
	validThresholds := map[int]bool{50: true, 75: true, 90: true, 100: true}
	if !validThresholds[req.Threshold] {
		common.BadRequest(w, "Invalid threshold. Must be 50, 75, 90, or 100")
		return
	}

	// Validate alert type
	validTypes := map[string]bool{
		"operations":    true,
		"ai_credits":    true,
		"storage":       true,
		"data_transfer": true,
	}
	if !validTypes[req.AlertType] {
		common.BadRequest(w, "Invalid alert type")
		return
	}

	alert := billing.NewUsageAlert(workspaceID, billing.UsageAlertType(req.AlertType), req.Threshold)
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

	if err := h.repo.Create(ctx, alert); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToAlertResponse(alert))
}
