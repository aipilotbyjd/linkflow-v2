package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type UpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
}

type UpdateHandler struct {
	workspaceRepo workspace.Repository
}

func NewUpdateHandler(workspaceRepo workspace.Repository) *UpdateHandler {
	return &UpdateHandler{workspaceRepo: workspaceRepo}
}

func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workspace ID")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	ws, err := h.workspaceRepo.FindByID(r.Context(), workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if req.Name != nil {
		ws.Name = *req.Name
	}
	if req.Description != nil {
		ws.Description = req.Description
	}
	if req.LogoURL != nil {
		ws.LogoURL = req.LogoURL
	}
	if req.Timezone != nil {
		ws.Timezone = *req.Timezone
	}

	if err := h.workspaceRepo.Update(r.Context(), ws); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toWorkspaceResponse(ws))
}
