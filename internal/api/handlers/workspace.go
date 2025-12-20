package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type WorkspaceHandler struct {
	workspaceSvc *services.WorkspaceService
	billingSvc   *services.BillingService
}

func NewWorkspaceHandler(workspaceSvc *services.WorkspaceService, billingSvc *services.BillingService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceSvc: workspaceSvc,
		billingSvc:   billingSvc,
	}
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workspaces, err := h.workspaceSvc.GetUserWorkspaces(r.Context(), claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list workspaces")
		return
	}

	var response []dto.WorkspaceResponse
	for _, ws := range workspaces {
		response = append(response, dto.WorkspaceResponse{
			ID:          ws.ID.String(),
			Name:        ws.Name,
			Slug:        ws.Slug,
			Description: ws.Description,
			LogoURL:     ws.LogoURL,
			PlanID:      ws.PlanID,
			CreatedAt:   ws.CreatedAt.Unix(),
		})
	}

	dto.JSON(w, http.StatusOK, response)
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	workspace, err := h.workspaceSvc.Create(r.Context(), services.CreateWorkspaceInput{
		OwnerID:     claims.UserID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	})
	if err != nil {
		if err == services.ErrSlugExists {
			dto.ErrorResponse(w, http.StatusConflict, "slug already exists")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create workspace")
		return
	}

	dto.Created(w, dto.WorkspaceResponse{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Slug:        workspace.Slug,
		Description: workspace.Description,
		PlanID:      workspace.PlanID,
		CreatedAt:   workspace.CreatedAt.Unix(),
	})
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	workspace, err := h.workspaceSvc.GetByID(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workspace not found")
		return
	}

	dto.JSON(w, http.StatusOK, dto.WorkspaceResponse{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Slug:        workspace.Slug,
		Description: workspace.Description,
		LogoURL:     workspace.LogoURL,
		PlanID:      workspace.PlanID,
		CreatedAt:   workspace.CreatedAt.Unix(),
	})
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	var req dto.UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	workspace, err := h.workspaceSvc.Update(r.Context(), wsCtx.WorkspaceID, services.UpdateWorkspaceInput{
		Name:        req.Name,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		Settings:    req.Settings,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update workspace")
		return
	}

	dto.JSON(w, http.StatusOK, dto.WorkspaceResponse{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Slug:        workspace.Slug,
		Description: workspace.Description,
		LogoURL:     workspace.LogoURL,
		PlanID:      workspace.PlanID,
		CreatedAt:   workspace.CreatedAt.Unix(),
	})
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	if err := h.workspaceSvc.Delete(r.Context(), wsCtx.WorkspaceID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete workspace")
		return
	}

	dto.NoContent(w)
}

func (h *WorkspaceHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	members, err := h.workspaceSvc.GetMembers(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get members")
		return
	}

	var response []dto.WorkspaceMemberResponse
	for _, m := range members {
		var joinedAt, invitedAt *int64
		if m.JoinedAt != nil {
			ts := m.JoinedAt.Unix()
			joinedAt = &ts
		}
		if m.InvitedAt != nil {
			ts := m.InvitedAt.Unix()
			invitedAt = &ts
		}

		response = append(response, dto.WorkspaceMemberResponse{
			ID: m.ID.String(),
			User: dto.UserResponse{
				ID:        m.User.ID.String(),
				Email:     m.User.Email,
				FirstName: m.User.FirstName,
				LastName:  m.User.LastName,
				AvatarURL: m.User.AvatarURL,
			},
			Role:      m.Role,
			JoinedAt:  joinedAt,
			InvitedAt: invitedAt,
		})
	}

	dto.JSON(w, http.StatusOK, response)
}

func (h *WorkspaceHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	var req dto.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	invitation, err := h.workspaceSvc.InviteMember(r.Context(), services.InviteMemberInput{
		WorkspaceID: wsCtx.WorkspaceID,
		Email:       req.Email,
		Role:        req.Role,
		InvitedBy:   claims.UserID,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to invite member")
		return
	}

	dto.Created(w, map[string]string{
		"id":    invitation.ID.String(),
		"email": invitation.Email,
		"role":  invitation.Role,
	})
}

func (h *WorkspaceHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.workspaceSvc.UpdateMemberRole(r.Context(), wsCtx.WorkspaceID, userID, req.Role); err != nil {
		if err == services.ErrCannotRemoveOwner {
			dto.ErrorResponse(w, http.StatusForbidden, "cannot change owner's role")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update member role")
		return
	}

	dto.NoContent(w)
}

func (h *WorkspaceHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.workspaceSvc.RemoveMember(r.Context(), wsCtx.WorkspaceID, userID); err != nil {
		if err == services.ErrCannotRemoveOwner {
			dto.ErrorResponse(w, http.StatusForbidden, "cannot remove workspace owner")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to remove member")
		return
	}

	dto.NoContent(w)
}
