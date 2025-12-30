package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type VersionDiffHandler struct {
	diffSvc *services.VersionDiffService
}

func NewVersionDiffHandler(diffSvc *services.VersionDiffService) *VersionDiffHandler {
	return &VersionDiffHandler{diffSvc: diffSvc}
}

func (h *VersionDiffHandler) Compare(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	fromVersion, err := strconv.Atoi(dto.QueryString(r, "from"))
	if err != nil {
		dto.BadRequest(w, "invalid from version")
		return
	}

	toVersion, err := strconv.Atoi(dto.QueryString(r, "to"))
	if err != nil {
		dto.BadRequest(w, "invalid to version")
		return
	}

	diff, err := h.diffSvc.Compare(r.Context(), workflowID, fromVersion, toVersion)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to compare versions")
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()

	dto.NewResponse(diff).
		WithLinks(&dto.Links{Self: "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/versions/compare"}).
		Send(w)
}

func (h *VersionDiffHandler) CompareWithCurrent(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		dto.BadRequest(w, "invalid version")
		return
	}

	diff, err := h.diffSvc.CompareWithCurrent(r.Context(), workflowID, version)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to compare with current")
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()

	dto.NewResponse(diff).
		WithLinks(&dto.Links{Self: "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/versions/" + strconv.Itoa(version) + "/diff"}).
		Send(w)
}
