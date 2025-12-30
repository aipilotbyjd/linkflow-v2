package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type EnvironmentVariableHandler struct {
	envSvc *services.EnvironmentVariableService
}

func NewEnvironmentVariableHandler(envSvc *services.EnvironmentVariableService) *EnvironmentVariableHandler {
	return &EnvironmentVariableHandler{envSvc: envSvc}
}

func (h *EnvironmentVariableHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	env := dto.QueryString(r, "environment")
	var envPtr *string
	if env != "" {
		envPtr = &env
	}

	vars, err := h.envSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, envPtr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list environment variables")
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/variables"

	type EnvVarWithActions struct {
		ID          string       `json:"id"`
		Name        string       `json:"name"`
		Value       string       `json:"value,omitempty"`
		IsSecret    bool         `json:"is_secret"`
		Environment *string      `json:"environment,omitempty"`
		Description *string      `json:"description,omitempty"`
		CreatedAt   int64        `json:"created_at"`
		UpdatedAt   int64        `json:"updated_at"`
		Actions     []dto.Action `json:"actions,omitempty"`
	}

	response := make([]EnvVarWithActions, len(vars))
	for i, v := range vars {
		varID := v.ID.String()
		varPath := basePath + "/" + varID

		actions := []dto.Action{
			{Name: "edit", Method: "PUT", Href: varPath, Label: "Edit Variable"},
			{Name: "delete", Method: "DELETE", Href: varPath, Label: "Delete"},
		}

		var value string
		if !v.IsSecret {
			value = "[REDACTED]"
		}

		response[i] = EnvVarWithActions{
			ID:          varID,
			Name:        v.Name,
			Value:       value,
			IsSecret:    v.IsSecret,
			Environment: v.Environment,
			Description: v.Description,
			CreatedAt:   v.CreatedAt.Unix(),
			UpdatedAt:   v.UpdatedAt.Unix(),
			Actions:     actions,
		}
	}

	data := dto.SelectFields(r, response)

	dto.NewResponse(data).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *EnvironmentVariableHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Value       string  `json:"value"`
		IsSecret    bool    `json:"is_secret"`
		Environment *string `json:"environment,omitempty"`
		Description *string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	envVar, err := h.envSvc.Create(r.Context(), services.CreateEnvVarInput{
		WorkspaceID: wsCtx.WorkspaceID,
		CreatedBy:   claims.UserID,
		Name:        req.Name,
		Value:       req.Value,
		IsSecret:    req.IsSecret,
		Environment: req.Environment,
		Description: req.Description,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create environment variable")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":          envVar.ID,
		"name":        envVar.Name,
		"is_secret":   envVar.IsSecret,
		"environment": envVar.Environment,
	})
}

func (h *EnvironmentVariableHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.ParseUUID(w, r, "varId")
	if !ok {
		return
	}

	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if err := h.envSvc.Update(r.Context(), id, req.Value); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update environment variable")
		return
	}

	dto.OK(w, map[string]string{"message": "environment variable updated"})
}

func (h *EnvironmentVariableHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.ParseUUID(w, r, "varId")
	if !ok {
		return
	}

	if err := h.envSvc.Delete(r.Context(), id); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete environment variable")
		return
	}

	dto.NoContent(w)
}
