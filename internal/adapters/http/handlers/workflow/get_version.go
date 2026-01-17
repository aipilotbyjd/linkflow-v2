package workflow

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type GetVersionHandler struct {
	versionRepo workflow.VersionRepository
}

func NewGetVersionHandler(versionRepo workflow.VersionRepository) *GetVersionHandler {
	return &GetVersionHandler{versionRepo: versionRepo}
}

func (h *GetVersionHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	v, err := h.versionRepo.FindByWorkflowAndVersion(r.Context(), workflowID, version)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToVersionResponse(v))
}
