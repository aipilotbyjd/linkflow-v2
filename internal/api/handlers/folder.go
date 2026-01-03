package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type FolderHandler struct {
	folderSvc *services.FolderService
}

func NewFolderHandler(folderSvc *services.FolderService) *FolderHandler {
	return &FolderHandler{folderSvc: folderSvc}
}

type CreateFolderRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
}

func (h *FolderHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	var parentID *uuid.UUID
	if pidStr := r.URL.Query().Get("parent_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			dto.BadRequest(w, "invalid parent_id")
			return
		}
		parentID = &pid
	}

	folders, err := h.folderSvc.List(r.Context(), wsCtx.WorkspaceID, parentID)
	if err != nil {
		dto.InternalServerError(w, "failed to list folders")
		return
	}

	dto.NewResponse(folders).
		WithMeta(&dto.Meta{Total: int64(len(folders))}).
		Send(w)
}

func (h *FolderHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	tree, err := h.folderSvc.GetTree(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.InternalServerError(w, "failed to get folder tree")
		return
	}

	dto.NewResponse(tree).Send(w)
}

func (h *FolderHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	var req CreateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		dto.BadRequest(w, "name is required")
		return
	}

	input := services.CreateFolderInput{
		WorkspaceID: wsCtx.WorkspaceID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}

	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			dto.BadRequest(w, "invalid parent_id")
			return
		}
		input.ParentID = &pid
	}

	folder, err := h.folderSvc.Create(r.Context(), input)
	if err != nil {
		dto.BadRequest(w, err.Error())
		return
	}

	dto.Created(w, folder)
}

func (h *FolderHandler) Get(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	folderID, ok := middleware.ParseUUID(w, r, "folderID")
	if !ok {
		return
	}

	folder, err := h.folderSvc.Get(r.Context(), folderID)
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "folder")
			return
		}
		dto.InternalServerError(w, "failed to get folder")
		return
	}

	if folder.WorkspaceID != wsCtx.WorkspaceID {
		dto.Forbidden(w, "access denied")
		return
	}

	dto.NewResponse(folder).Send(w)
}

func (h *FolderHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	folderID, ok := middleware.ParseUUID(w, r, "folderID")
	if !ok {
		return
	}

	var req struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		Color       *string `json:"color,omitempty"`
		Icon        *string `json:"icon,omitempty"`
		ParentID    *string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	input := services.UpdateFolderInput{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}

	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			dto.BadRequest(w, "invalid parent_id")
			return
		}
		input.ParentID = &pid
	}

	folder, err := h.folderSvc.Update(r.Context(), folderID, wsCtx.WorkspaceID, input)
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "folder")
			return
		}
		if err == services.ErrForbidden {
			dto.Forbidden(w, "access denied")
			return
		}
		dto.BadRequest(w, err.Error())
		return
	}

	dto.NewResponse(folder).Send(w)
}

func (h *FolderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	folderID, ok := middleware.ParseUUID(w, r, "folderID")
	if !ok {
		return
	}

	if err := h.folderSvc.Delete(r.Context(), folderID, wsCtx.WorkspaceID); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "folder")
			return
		}
		if err == services.ErrForbidden {
			dto.Forbidden(w, "access denied")
			return
		}
		dto.BadRequest(w, err.Error())
		return
	}

	dto.NewResponse(map[string]string{"message": "Folder deleted"}).Send(w)
}
