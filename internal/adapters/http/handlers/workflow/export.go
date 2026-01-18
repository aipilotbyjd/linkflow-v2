package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ExportedWorkflow struct {
	Name        string          `json:"name" validate:"required"`
	Description *string         `json:"description,omitempty"`
	Nodes       types.JSONArray `json:"nodes" validate:"required"`
	Connections types.JSONArray `json:"connections"`
	Settings    types.JSON      `json:"settings,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Version     int             `json:"version"`
	ExportedAt  string          `json:"exported_at"`
}

type ExportHandler struct {
	workflowRepo workflow.Repository
}

func NewExportHandler(workflowRepo workflow.Repository) *ExportHandler {
	return &ExportHandler{workflowRepo: workflowRepo}
}

func (h *ExportHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	wf, err := h.workflowRepo.FindByID(r.Context(), workflowID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	exported := ExportedWorkflow{
		Name:        wf.Name,
		Description: wf.Description,
		Nodes:       wf.Nodes,
		Connections: wf.Connections,
		Settings:    wf.Settings,
		Tags:        wf.Tags,
		Version:     wf.Version,
		ExportedAt:  wf.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=workflow-"+wf.ID.String()+".json")
	json.NewEncoder(w).Encode(exported)
}
