package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type WorkflowShareHandler struct {
	shareSvc *services.WorkflowShareService
}

func NewWorkflowShareHandler(shareSvc *services.WorkflowShareService) *WorkflowShareHandler {
	return &WorkflowShareHandler{shareSvc: shareSvc}
}

type ShareWorkflowRequest struct {
	WorkflowID        string  `json:"workflow_id"`
	TargetWorkspaceID string  `json:"target_workspace_id"`
	Permission        string  `json:"permission"`
	ExpiresAt         *string `json:"expires_at,omitempty"`
}

func (h *WorkflowShareHandler) Share(w http.ResponseWriter, r *http.Request) {
	userCtx, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if userCtx == nil {
		return
	}

	var req ShareWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	workflowID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		dto.BadRequest(w, "invalid workflow_id")
		return
	}

	targetWorkspaceID, err := uuid.Parse(req.TargetWorkspaceID)
	if err != nil {
		dto.BadRequest(w, "invalid target_workspace_id")
		return
	}

	input := services.ShareWorkflowInput{
		WorkflowID:        workflowID,
		SourceWorkspaceID: wsCtx.WorkspaceID,
		TargetWorkspaceID: targetWorkspaceID,
		SharedBy:          userCtx.UserID,
		Permission:        req.Permission,
	}

	if req.ExpiresAt != nil {
		expiresAt, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			dto.BadRequest(w, "invalid expires_at format (use RFC3339)")
			return
		}
		input.ExpiresAt = &expiresAt
	}

	share, err := h.shareSvc.Share(r.Context(), input)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.Created(w, share)
}

func (h *WorkflowShareHandler) Accept(w http.ResponseWriter, r *http.Request) {
	userCtx, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if userCtx == nil {
		return
	}

	shareID, err := uuid.Parse(chi.URLParam(r, "shareID"))
	if err != nil {
		dto.BadRequest(w, "invalid share ID")
		return
	}

	share, err := h.shareSvc.Accept(r.Context(), shareID, userCtx.UserID, wsCtx.WorkspaceID)
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "share not found")
			return
		}
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, share)
}

func (h *WorkflowShareHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	shareID, err := uuid.Parse(chi.URLParam(r, "shareID"))
	if err != nil {
		dto.BadRequest(w, "invalid share ID")
		return
	}

	if err := h.shareSvc.Revoke(r.Context(), shareID, wsCtx.WorkspaceID); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "share not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.NoContent(w)
}

func (h *WorkflowShareHandler) ListSharedByMe(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	shares, err := h.shareSvc.ListSharedByWorkspace(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"shares": shares,
		"count":  len(shares),
	})
}

func (h *WorkflowShareHandler) ListSharedWithMe(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	shares, err := h.shareSvc.ListSharedWithWorkspace(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"shares": shares,
		"count":  len(shares),
	})
}

func (h *WorkflowShareHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	shares, err := h.shareSvc.GetPendingShares(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"shares": shares,
		"count":  len(shares),
	})
}

func (h *WorkflowShareHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	shareID, err := uuid.Parse(chi.URLParam(r, "shareID"))
	if err != nil {
		dto.BadRequest(w, "invalid share ID")
		return
	}

	var req struct {
		Permission string `json:"permission"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if err := h.shareSvc.UpdatePermission(r.Context(), shareID, wsCtx.WorkspaceID, req.Permission); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "share not found")
			return
		}
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]string{"message": "permission updated"})
}
