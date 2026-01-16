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

type VersionResponse struct {
	ID          string          `json:"id"`
	WorkflowID  string          `json:"workflow_id"`
	Version     int             `json:"version"`
	Nodes       types.JSONArray `json:"nodes"`
	Connections types.JSONArray `json:"connections"`
	Settings    types.JSON      `json:"settings,omitempty"`
	ChangeNote  *string         `json:"change_note,omitempty"`
	CreatedBy   string          `json:"created_by"`
	CreatedAt   string          `json:"created_at"`
}

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
		responses = append(responses, toVersionResponse(&v))
	}

	common.List(w, responses, types.PageResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		HasMore:    int64(page*pageSize) < total,
	})
}

func toVersionResponse(v *workflow.Version) VersionResponse {
	createdBy := ""
	if v.CreatedBy != nil {
		createdBy = v.CreatedBy.String()
	}
	return VersionResponse{
		ID:          v.ID.String(),
		WorkflowID:  v.WorkflowID.String(),
		Version:     v.Version,
		Nodes:       v.Nodes,
		Connections: v.Connections,
		Settings:    v.Settings,
		ChangeNote:  v.ChangeMessage,
		CreatedBy:   createdBy,
		CreatedAt:   v.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
