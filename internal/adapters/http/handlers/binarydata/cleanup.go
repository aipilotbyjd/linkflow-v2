package binarydata

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
)

// CleanupHandler handles cleanup old files request
type CleanupHandler struct {
	repo binarydata.Repository
}

// NewCleanupHandler creates a new handler
func NewCleanupHandler(repo binarydata.Repository) *CleanupHandler {
	return &CleanupHandler{repo: repo}
}

// Handle handles the cleanup old files request
func (h *CleanupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30 // Default 30 days
	}

	deleted, err := h.repo.DeleteOlderThan(r.Context(), workspaceID, days)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"deleted": deleted,
		"message": "Cleanup completed successfully",
	})
}
