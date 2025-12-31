package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type WorkflowVariableHandler struct {
	varSvc *services.WorkflowVariableService
}

func NewWorkflowVariableHandler(varSvc *services.WorkflowVariableService) *WorkflowVariableHandler {
	return &WorkflowVariableHandler{varSvc: varSvc}
}

type CreateVariableRequest struct {
	Name        string  `json:"name" validate:"required"`
	Key         string  `json:"key" validate:"required"`
	Type        string  `json:"type,omitempty"`
	Value       string  `json:"value,omitempty"`
	Default     string  `json:"default,omitempty"`
	Description *string `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
}

func (h *WorkflowVariableHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	variables, err := h.varSvc.List(r.Context(), workflowID)
	if err != nil {
		dto.InternalServerError(w, "failed to list variables")
		return
	}

	dto.NewResponse(variables).
		WithMeta(&dto.Meta{Total: int64(len(variables))}).
		Send(w)
}

func (h *WorkflowVariableHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	var req CreateVariableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" || req.Key == "" {
		dto.BadRequest(w, "name and key are required")
		return
	}

	variable, err := h.varSvc.Create(r.Context(), services.CreateVariableInput{
		WorkflowID:  workflowID,
		Name:        req.Name,
		Key:         req.Key,
		Type:        req.Type,
		Value:       req.Value,
		Default:     req.Default,
		Description: req.Description,
		Required:    req.Required,
	})
	if err != nil {
		dto.BadRequest(w, err.Error())
		return
	}

	dto.Created(w, variable)
}

func (h *WorkflowVariableHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	varID, ok := middleware.ParseUUID(w, r, "variableID")
	if !ok {
		return
	}

	var req struct {
		Name        *string `json:"name,omitempty"`
		Value       *string `json:"value,omitempty"`
		Default     *string `json:"default,omitempty"`
		Description *string `json:"description,omitempty"`
		Required    *bool   `json:"required,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	variable, err := h.varSvc.Update(r.Context(), varID, services.UpdateVariableInput{
		Name:        req.Name,
		Value:       req.Value,
		Default:     req.Default,
		Description: req.Description,
		Required:    req.Required,
	})
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "variable")
			return
		}
		dto.InternalServerError(w, "failed to update variable")
		return
	}

	dto.NewResponse(variable).Send(w)
}

func (h *WorkflowVariableHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	varID, ok := middleware.ParseUUID(w, r, "variableID")
	if !ok {
		return
	}

	if err := h.varSvc.Delete(r.Context(), varID); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "variable")
			return
		}
		dto.InternalServerError(w, "failed to delete variable")
		return
	}

	dto.NewResponse(map[string]string{"message": "Variable deleted"}).Send(w)
}
