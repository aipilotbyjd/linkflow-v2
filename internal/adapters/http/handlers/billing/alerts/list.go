package alerts

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// ListHandler handles listing usage alerts
type ListHandler struct {
	repo billing.UsageAlertRepository
}

func NewListHandler(repo billing.UsageAlertRepository) *ListHandler {
	return &ListHandler{repo: repo}
}

func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	alerts, err := h.repo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	responses := make([]AlertResponse, len(alerts))
	for i, alert := range alerts {
		responses[i] = ToAlertResponse(alert)
	}

	common.Success(w, map[string]interface{}{
		"alerts": responses,
		"total":  len(responses),
	})
}
