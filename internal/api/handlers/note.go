package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type NoteHandler struct {
	noteSvc *services.NoteService
}

// NewNoteHandler creates a new NoteHandler
func NewNoteHandler(noteSvc *services.NoteService) *NoteHandler {
	return &NoteHandler{noteSvc: noteSvc}
}

// List returns all notes in a workspace with optional filters
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	pg := dto.ParsePagination(r)
	filters := dto.ParseNoteFilters(r)

	notes, total, err := h.noteSvc.GetWithFilters(r.Context(), wsCtx.WorkspaceID, filters.ToRepoFilter(), pg.Opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list notes")
		return
	}

	response := make([]dto.NoteResponse, len(notes))
	for i, note := range notes {
		response[i] = dto.BuildNoteResponse(&note)
	}

	wsID := wsCtx.WorkspaceID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/notes"
	filterQS := filters.ToQueryString()

	links := &dto.Links{
		Self: fmt.Sprintf("%s?page=%d&per_page=%d%s", basePath, pg.Page, pg.PerPage, filterQS),
	}
	meta := pg.NewMeta(total)
	if pg.Page < meta.TotalPages {
		links.Next = fmt.Sprintf("%s?page=%d&per_page=%d%s", basePath, pg.Page+1, pg.PerPage, filterQS)
	}
	if pg.Page > 1 {
		links.Prev = fmt.Sprintf("%s?page=%d&per_page=%d%s", basePath, pg.Page-1, pg.PerPage, filterQS)
	}
	links.First = fmt.Sprintf("%s?page=1&per_page=%d%s", basePath, pg.PerPage, filterQS)
	if meta.TotalPages > 0 {
		links.Last = fmt.Sprintf("%s?page=%d&per_page=%d%s", basePath, meta.TotalPages, pg.PerPage, filterQS)
	}

	data := dto.SelectFields(r, response)

	dto.NewResponse(data).
		WithLinks(links).
		WithMeta(meta).
		Send(w)
}

// Create creates a new note
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	var req dto.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid resource_id")
		return
	}

	var color string
	if req.Color != nil {
		color = *req.Color
	}

	note, err := h.noteSvc.Create(r.Context(), services.CreateNoteInput{
		WorkspaceID:  wsCtx.WorkspaceID,
		ResourceID:   resourceID,
		ResourceName: req.ResourceName,
		CreatedBy:    claims.UserID,
		Content:      req.Content,
		Position:     *req.Position,
		Size:         *req.Size,
		Color:        color,
	})
	if err != nil {
		if errors.Is(err, services.ErrNoteContentRequired) || errors.Is(err, services.ErrNoteResourceRequired) {
			dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create note")
		return
	}

	dto.Created(w, dto.BuildNoteResponse(note))
}

// Get returns a single note by ID
func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	noteID, ok := middleware.ParseUUID(w, r, "noteID")
	if !ok {
		return
	}

	note, err := h.noteSvc.GetByID(r.Context(), noteID)
	if err != nil {
		if errors.Is(err, services.ErrNoteNotFound) {
			dto.ErrorResponse(w, http.StatusNotFound, "note not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get note")
		return
	}

	// Verify workspace ownership
	if note.WorkspaceID != wsCtx.WorkspaceID {
		dto.ErrorResponse(w, http.StatusNotFound, "note not found")
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	noteIDStr := note.ID.String()

	dto.NewResponse(dto.BuildNoteResponse(note)).
		WithLinks(&dto.Links{Self: "/api/v1/workspaces/" + wsID + "/notes/" + noteIDStr}).
		Send(w)
}

// Update updates an existing note
func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	noteID, ok := middleware.ParseUUID(w, r, "noteID")
	if !ok {
		return
	}

	// Verify note exists and belongs to workspace
	existingNote, err := h.noteSvc.GetByID(r.Context(), noteID)
	if err != nil {
		if errors.Is(err, services.ErrNoteNotFound) {
			dto.ErrorResponse(w, http.StatusNotFound, "note not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get note")
		return
	}

	if existingNote.WorkspaceID != wsCtx.WorkspaceID {
		dto.ErrorResponse(w, http.StatusNotFound, "note not found")
		return
	}

	var req dto.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	note, err := h.noteSvc.Update(r.Context(), noteID, services.UpdateNoteInput{
		Content:  req.Content,
		Position: req.Position,
		Size:     req.Size,
		Color:    req.Color,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update note")
		return
	}

	dto.JSON(w, http.StatusOK, dto.BuildNoteResponse(note))
}

// Delete deletes a note
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	noteID, ok := middleware.ParseUUID(w, r, "noteID")
	if !ok {
		return
	}

	// Verify note exists and belongs to workspace
	existingNote, err := h.noteSvc.GetByID(r.Context(), noteID)
	if err != nil {
		if errors.Is(err, services.ErrNoteNotFound) {
			dto.ErrorResponse(w, http.StatusNotFound, "note not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get note")
		return
	}

	if existingNote.WorkspaceID != wsCtx.WorkspaceID {
		dto.ErrorResponse(w, http.StatusNotFound, "note not found")
		return
	}

	if err := h.noteSvc.Delete(r.Context(), noteID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete note")
		return
	}

	dto.NoContent(w)
}
