package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type ProjectHandler struct {
	projectSvc *services.ProjectService
}

func NewProjectHandler(projectSvc *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectSvc: projectSvc}
}

type CreateProjectRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
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

	projects, err := h.projectSvc.List(r.Context(), wsCtx.WorkspaceID, parentID)
	if err != nil {
		dto.InternalServerError(w, "failed to list projects")
		return
	}

	dto.NewResponse(projects).
		WithMeta(&dto.Meta{Total: int64(len(projects))}).
		Send(w)
}

func (h *ProjectHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	tree, err := h.projectSvc.GetTree(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.InternalServerError(w, "failed to get project tree")
		return
	}

	dto.NewResponse(tree).Send(w)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		dto.BadRequest(w, "name is required")
		return
	}

	input := services.CreateProjectInput{
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

	project, err := h.projectSvc.Create(r.Context(), input)
	if err != nil {
		dto.BadRequest(w, err.Error())
		return
	}

	dto.Created(w, project)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	projectID, ok := middleware.ParseUUID(w, r, "projectID")
	if !ok {
		return
	}

	project, err := h.projectSvc.Get(r.Context(), projectID)
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "project")
			return
		}
		dto.InternalServerError(w, "failed to get project")
		return
	}

	if project.WorkspaceID != wsCtx.WorkspaceID {
		dto.Forbidden(w, "access denied")
		return
	}

	dto.NewResponse(project).Send(w)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	projectID, ok := middleware.ParseUUID(w, r, "projectID")
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

	input := services.UpdateProjectInput{
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

	project, err := h.projectSvc.Update(r.Context(), projectID, wsCtx.WorkspaceID, input)
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "project")
			return
		}
		if err == services.ErrForbidden {
			dto.Forbidden(w, "access denied")
			return
		}
		dto.BadRequest(w, err.Error())
		return
	}

	dto.NewResponse(project).Send(w)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	projectID, ok := middleware.ParseUUID(w, r, "projectID")
	if !ok {
		return
	}

	if err := h.projectSvc.Delete(r.Context(), projectID, wsCtx.WorkspaceID); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "project")
			return
		}
		if err == services.ErrForbidden {
			dto.Forbidden(w, "access denied")
			return
		}
		dto.BadRequest(w, err.Error())
		return
	}

	dto.NewResponse(map[string]string{"message": "Project deleted"}).Send(w)
}
