package alerts

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
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

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
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
