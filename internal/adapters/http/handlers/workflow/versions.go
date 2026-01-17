package workflow

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListVersionsHandler struct {
	versionRepo workflow.VersionRepository
}

func NewListVersionsHandler(versionRepo workflow.VersionRepository) *ListVersionsHandler {
	return &ListVersionsHandler{versionRepo: versionRepo}
}

func (h *ListVersionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	versions, total, err := h.versionRepo.FindByWorkflowID(r.Context(), workflowID, types.NewListOptions(page, pageSize))
	if err != nil {
		common.HandleError(w, err)
		return
	}

	var responses []VersionResponse
	for _, v := range versions {
		responses = append(responses, ToVersionResponse(&v))
	}

	common.List(w, responses, types.PageResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		HasMore:    int64(page*pageSize) < total,
	})
}
