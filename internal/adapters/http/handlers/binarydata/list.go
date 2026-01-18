package binarydata

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
)

// ListHandler handles list binary data request
type ListHandler struct {
	repo binarydata.Repository
}

// NewListHandler creates a new handler
func NewListHandler(repo binarydata.Repository) *ListHandler {
	return &ListHandler{repo: repo}
}

// Handle handles the list binary data request
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	files, total, err := h.repo.FindByWorkspace(r.Context(), workspaceID, limit, offset)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"files": ToBinaryDataResponseList(files),
		"total": total,
	})
}
