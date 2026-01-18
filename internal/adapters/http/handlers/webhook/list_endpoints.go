package webhook

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListEndpointsHandler struct {
	webhookRepo webhook.Repository
	baseURL     string
}

func NewListEndpointsHandler(webhookRepo webhook.Repository, baseURL string) *ListEndpointsHandler {
	return &ListEndpointsHandler{
		webhookRepo: webhookRepo,
		baseURL:     baseURL,
	}
}

func (h *ListEndpointsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}
	workspaceID := wsCtx.WorkspaceID

	// Parse pagination params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	opts := types.NewListOptions(page, pageSize)

	endpoints, total, err := h.webhookRepo.FindByWorkspaceID(r.Context(), workspaceID, opts)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Transform to response format
	result := make([]map[string]interface{}, 0, len(endpoints))
	for _, ep := range endpoints {
		result = append(result, map[string]interface{}{
			"id":             ep.ID.String(),
			"workflow_id":    ep.WorkflowID.String(),
			"path":           ep.Path,
			"method":         ep.Method,
			"is_active":      ep.IsActive,
			"call_count":     ep.CallCount,
			"last_called_at": ep.LastCalledAt,
			"url":            ep.GetURL(h.baseURL),
			"created_at":     ep.CreatedAt,
		})
	}

	common.Success(w, map[string]interface{}{
		"endpoints": result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
