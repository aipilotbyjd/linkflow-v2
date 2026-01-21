package alerts

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// HistoryHandler handles listing alert history
type HistoryHandler struct {
	repo billing.UsageAlertLogRepository
}

func NewHistoryHandler(repo billing.UsageAlertLogRepository) *HistoryHandler {
	return &HistoryHandler{repo: repo}
}

func (h *HistoryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	// Parse pagination
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, total, err := h.repo.FindByWorkspaceID(ctx, workspaceID, limit, offset)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	responses := make([]AlertLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = ToAlertLogResponse(log)
	}

	common.Success(w, map[string]interface{}{
		"alerts": responses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
