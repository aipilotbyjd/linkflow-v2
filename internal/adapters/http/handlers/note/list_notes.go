package note

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/note"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListNotesHandler struct {
	repo note.Repository
}

func NewListNotesHandler(repo note.Repository) *ListNotesHandler {
	return &ListNotesHandler{repo: repo}
}

func (h *ListNotesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	if workflowIDStr == "" {
		workflowIDStr = r.URL.Query().Get("resource_id")
	}

	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	page := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := types.DefaultPageSize
	if ps := query.Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	opts := &note.ListOptions{
		ListOptions: types.NewListOptions(page, pageSize),
		NodeID:      stringPtrOrNil(query.Get("node_id")),
	}

	if query.Get("resolved") == "true" {
		opts.ResolvedOnly = true
	} else if query.Get("resolved") == "false" {
		opts.UnresolvedOnly = true
	}

	notes, total, err := h.repo.FindByWorkflow(r.Context(), workflowID, opts)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Convert to response
	response := make([]*NoteResponse, len(notes))
	for i, c := range notes {
		response[i] = ToResponse(c)
	}

	common.List(w, response, types.NewPageResponse(page, pageSize, total))
}

func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
