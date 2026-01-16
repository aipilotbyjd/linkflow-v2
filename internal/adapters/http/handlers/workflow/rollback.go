package workflow

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type RollbackHandler struct {
	workflowRepo workflow.Repository
	versionRepo  workflow.VersionRepository
}

func NewRollbackHandler(workflowRepo workflow.Repository, versionRepo workflow.VersionRepository) *RollbackHandler {
	return &RollbackHandler{workflowRepo: workflowRepo, versionRepo: versionRepo}
}

func (h *RollbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	versionStr := chi.URLParam(r, "version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		common.BadRequest(w, "invalid version number")
		return
	}

	wf, err := h.workflowRepo.FindByID(r.Context(), workflowID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	v, err := h.versionRepo.FindByWorkflowAndVersion(r.Context(), workflowID, version)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	wf.Nodes = v.Nodes
	wf.Connections = v.Connections
	wf.Settings = v.Settings

	if err := h.workflowRepo.Update(r.Context(), wf); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toWorkflowResponse(wf))
}
