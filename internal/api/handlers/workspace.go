package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/mappers"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type WorkspaceHandler struct {
	workspaceSvc *services.WorkspaceService
	billingSvc   *services.BillingService
}

// NewWorkspaceHandler creates a new WorkspaceHandler for workspace management.
func NewWorkspaceHandler(workspaceSvc *services.WorkspaceService, billingSvc *services.BillingService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceSvc: workspaceSvc,
		billingSvc:   billingSvc,
	}
}

// List returns all workspaces for the authenticated user.
func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	workspaces, err := h.workspaceSvc.GetUserWorkspaces(r.Context(), claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list workspaces")
		return
	}

	// Build response with actions
	type WorkspaceWithActions struct {
		dto.WorkspaceResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}

	response := []WorkspaceWithActions{}
	for _, ws := range workspaces {
		wsID := ws.ID.String()
		basePath := "/api/v1/workspaces/" + wsID

		actions := []dto.Action{
			{Name: "view", Method: "GET", Href: basePath, Label: "View Workspace"},
			{Name: "settings", Method: "GET", Href: basePath + "/settings", Label: "Settings"},
			{Name: "members", Method: "GET", Href: basePath + "/members", Label: "Members"},
			{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"},
		}

		response = append(response, WorkspaceWithActions{
			WorkspaceResponse: mappers.WorkspaceToResponse(&ws),
			Actions:           actions,
		})
	}

	// Apply field selection
	data := dto.SelectFields(r, response)

	dto.NewResponse(data).
		WithLinks(&dto.Links{Self: "/api/v1/workspaces"}).
		Send(w)
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
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
		OwnerID:      claims.UserID,
		Name:         req.Name,
		Slug:         req.Slug,
		Description:  req.Description,
		Timezone:     req.Timezone,
		Language:     req.Language,
		Currency:     req.Currency,
		Country:      req.Country,
		Industry:     req.Industry,
		CompanySize:  req.CompanySize,
		Website:      req.Website,
		BillingEmail: req.BillingEmail,
	})
	if err != nil {
		if err == services.ErrSlugExists {
			dto.ErrorResponse(w, http.StatusConflict, "slug already exists")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create workspace")
		return
	}

	dto.Created(w, mappers.WorkspaceToResponse(workspace))
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	workspace, err := h.workspaceSvc.GetByID(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workspace not found")
		return
	}

	wsID := workspace.ID.String()
	basePath := "/api/v1/workspaces/" + wsID

	actions := []dto.Action{
		{Name: "edit", Method: "PUT", Href: basePath, Label: "Edit Workspace"},
		{Name: "members", Method: "GET", Href: basePath + "/members", Label: "View Members"},
		{Name: "billing", Method: "GET", Href: basePath + "/billing", Label: "View Billing"},
		{Name: "settings", Method: "GET", Href: basePath + "/settings", Label: "Settings"},
	}

	response := struct {
		dto.WorkspaceResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}{
		WorkspaceResponse: mappers.WorkspaceToResponse(workspace),
		Actions:           actions,
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
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
		Name:         req.Name,
		Description:  req.Description,
		LogoURL:      req.LogoURL,
		Timezone:     req.Timezone,
		Language:     req.Language,
		Currency:     req.Currency,
		Country:      req.Country,
		Industry:     req.Industry,
		CompanySize:  req.CompanySize,
		Website:      req.Website,
		BillingEmail: req.BillingEmail,
		Settings:     req.Settings,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update workspace")
		return
	}

	basePath := "/api/v1/workspaces/" + workspace.ID.String()

	dto.NewResponse(mappers.WorkspaceToResponse(workspace)).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	if err := h.workspaceSvc.Delete(r.Context(), wsCtx.WorkspaceID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete workspace")
		return
	}

	dto.NoContent(w)
}

func (h *WorkspaceHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	members, err := h.workspaceSvc.GetMembers(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get members")
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/members"

	type MemberWithActions struct {
		dto.WorkspaceMemberResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}

	response := []MemberWithActions{}
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

		memberID := m.ID.String()
		memberPath := basePath + "/" + memberID

		actions := []dto.Action{
			{Name: "change_role", Method: "PUT", Href: memberPath, Label: "Change Role"},
			{Name: "remove", Method: "DELETE", Href: memberPath, Label: "Remove Member"},
		}

		response = append(response, MemberWithActions{
			WorkspaceMemberResponse: dto.WorkspaceMemberResponse{
				ID: memberID,
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
			},
			Actions: actions,
		})
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		WithMeta(&dto.Meta{Total: int64(len(response)), Page: 1, PerPage: len(response), TotalPages: 1}).
		Send(w)
}

func (h *WorkspaceHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
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
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	userID, ok := middleware.ParseUUID(w, r, "userID")
	if !ok {
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
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	userID, ok := middleware.ParseUUID(w, r, "userID")
	if !ok {
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
