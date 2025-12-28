package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type WorkflowCommentHandler struct {
	commentSvc *services.WorkflowCommentService
}

func NewWorkflowCommentHandler(commentSvc *services.WorkflowCommentService) *WorkflowCommentHandler {
	return &WorkflowCommentHandler{commentSvc: commentSvc}
}

func (h *WorkflowCommentHandler) List(w http.ResponseWriter, r *http.Request) {
	workflowID, ok := middleware.ParseUUID(w, r, "workflowId")
	if !ok {
		return
	}

	nodeID := dto.QueryString(r, "node_id")

	var comments interface{}
	var err error
	if nodeID != "" {
		comments, err = h.commentSvc.GetByNode(r.Context(), workflowID, nodeID)
	} else {
		comments, err = h.commentSvc.GetByWorkflow(r.Context(), workflowID)
	}

	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list comments")
		return
	}

	dto.OK(w, comments)
}

func (h *WorkflowCommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowId")
	if !ok {
		return
	}

	var req struct {
		NodeID   *string `json:"node_id,omitempty"`
		ParentID *string `json:"parent_id,omitempty"`
		Content  string  `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			dto.BadRequest(w, "invalid parent_id")
			return
		}
		parentID = &id
	}

	comment, err := h.commentSvc.Create(r.Context(), services.CreateCommentInput{
		WorkflowID:  workflowID,
		WorkspaceID: wsCtx.WorkspaceID,
		NodeID:      req.NodeID,
		ParentID:    parentID,
		CreatedBy:   claims.UserID,
		Content:     req.Content,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	dto.Created(w, comment)
}

func (h *WorkflowCommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.ParseUUID(w, r, "commentId")
	if !ok {
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if err := h.commentSvc.Update(r.Context(), id, req.Content); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update comment")
		return
	}

	dto.OK(w, map[string]string{"message": "comment updated"})
}

func (h *WorkflowCommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.ParseUUID(w, r, "commentId")
	if !ok {
		return
	}

	if err := h.commentSvc.Delete(r.Context(), id); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete comment")
		return
	}

	dto.NoContent(w)
}

func (h *WorkflowCommentHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	id, ok := middleware.ParseUUID(w, r, "commentId")
	if !ok {
		return
	}

	if err := h.commentSvc.Resolve(r.Context(), id, claims.UserID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to resolve comment")
		return
	}

	dto.OK(w, map[string]string{"message": "comment resolved"})
}
