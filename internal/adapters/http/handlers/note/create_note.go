package note

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/note"
)

type CreateNoteHandler struct {
	repo note.Repository
}

func NewCreateNoteHandler(repo note.Repository) *CreateNoteHandler {
	return &CreateNoteHandler{repo: repo}
}

func (h *CreateNoteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowId")
	if workflowIDStr == "" {
		workflowIDStr = r.URL.Query().Get("resource_id")
	}
	if workflowIDStr == "" {
		workflowIDStr = req.ResourceID
	}

	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	// Validate
	if req.Content == "" {
		common.BadRequest(w, "content is required")
		return
	}
	if len(req.Content) > 10000 {
		common.BadRequest(w, "content too long (max 10000 characters)")
		return
	}

	// Create note
	c, err := note.NewNote(workspaceID, workflowID, userID, req.Content)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if req.NodeID != nil {
		c.WithNodeID(*req.NodeID)
	}

	if err := h.repo.Create(r.Context(), c); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToResponse(c))
}
