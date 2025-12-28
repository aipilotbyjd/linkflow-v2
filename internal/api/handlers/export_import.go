package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type WorkflowExportHandler struct {
	exportSvc *services.WorkflowExportService
}

func NewWorkflowExportHandler(exportSvc *services.WorkflowExportService) *WorkflowExportHandler {
	return &WorkflowExportHandler{exportSvc: exportSvc}
}

func (h *WorkflowExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowId")
	if !ok {
		return
	}

	includeCredentials := dto.ParseBool(r, "include_credentials")

	data, err := h.exportSvc.Export(r.Context(), workflowID, claims.UserID, includeCredentials)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to export workflow")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=workflow-"+workflowID.String()+".json")
	json.NewEncoder(w).Encode(data)
}

func (h *WorkflowExportHandler) Import(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	var data services.WorkflowExportData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		dto.BadRequest(w, "invalid import data")
		return
	}

	workflow, err := h.exportSvc.Import(r.Context(), wsCtx.WorkspaceID, claims.UserID, &data)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to import workflow")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":      workflow.ID,
		"name":    workflow.Name,
		"message": "workflow imported successfully",
	})
}
