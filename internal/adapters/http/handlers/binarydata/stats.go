package binarydata

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
)

// StatsHandler handles storage stats request
type StatsHandler struct {
	repo binarydata.Repository
}

// NewStatsHandler creates a new handler
func NewStatsHandler(repo binarydata.Repository) *StatsHandler {
	return &StatsHandler{repo: repo}
}

// Handle handles the storage stats request
func (h *StatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	stats, err := h.repo.GetStats(r.Context(), workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToStorageStatsResponse(stats))
}
