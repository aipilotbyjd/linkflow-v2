package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type PinnedDataHandler struct {
	pinnedRepo *repositories.PinnedDataRepository
}

func NewPinnedDataHandler(pinnedRepo *repositories.PinnedDataRepository) *PinnedDataHandler {
	return &PinnedDataHandler{pinnedRepo: pinnedRepo}
}

// GetByWorkflow returns all pinned data for a workflow
func (h *PinnedDataHandler) GetByWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	pinnedList, err := h.pinnedRepo.FindByWorkflow(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/pinned-data"

	dto.NewResponse(map[string]interface{}{
		"pinned_data": pinnedList,
		"count":       len(pinnedList),
	}).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

// GetByNode returns pinned data for a specific node
func (h *PinnedDataHandler) GetByNode(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	nodeID := chi.URLParam(r, "nodeID")
	if nodeID == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "node ID required")
		return
	}

	pinned, err := h.pinnedRepo.FindByWorkflowAndNode(r.Context(), workflowID, nodeID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "pinned data not found")
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/pinned-data/" + nodeID

	dto.NewResponse(pinned).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(
			dto.Action{Name: "update", Method: "PUT", Href: basePath, Label: "Update"},
			dto.Action{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"},
		).
		Send(w)
}

// SetPinnedDataRequest represents a request to set pinned data
type SetPinnedDataRequest struct {
	NodeID string      `json:"node_id"`
	Name   string      `json:"name,omitempty"`
	Data   models.JSON `json:"data"`
}

// Set creates or updates pinned data for a node
func (h *PinnedDataHandler) Set(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	var req SetPinnedDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NodeID == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "node_id required")
		return
	}

	userCtx := middleware.MustUser(w, r)
	if userCtx == nil {
		return
	}

	pinned := &models.PinnedData{
		WorkflowID: workflowID,
		NodeID:     req.NodeID,
		Name:       req.Name,
		Data:       req.Data,
		CreatedBy:  userCtx.UserID,
	}

	if err := h.pinnedRepo.Upsert(r.Context(), pinned); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/pinned-data/" + req.NodeID

	dto.NewResponse(map[string]interface{}{
		"message":     "Pinned data saved",
		"pinned_data": pinned,
	}).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(
			dto.Action{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"},
		).
		Send(w)
}

// Delete removes pinned data for a node
func (h *PinnedDataHandler) Delete(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	nodeID := chi.URLParam(r, "nodeID")
	if nodeID == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "node ID required")
		return
	}

	if err := h.pinnedRepo.DeleteByWorkflowAndNode(r.Context(), workflowID, nodeID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/pinned-data"

	dto.NewResponse(map[string]interface{}{
		"message": "Pinned data deleted",
	}).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}
