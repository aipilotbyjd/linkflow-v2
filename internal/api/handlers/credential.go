package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type CredentialHandler struct {
	credentialSvc *services.CredentialService
}

// NewCredentialHandler creates a new CredentialHandler for credential management.
func NewCredentialHandler(credentialSvc *services.CredentialService) *CredentialHandler {
	return &CredentialHandler{credentialSvc: credentialSvc}
}

// List returns all credentials the user can access in a workspace.
func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	pg := dto.ParsePagination(r)
	filters := dto.ParseCredentialFilters(r)

	// Convert DTO filters to repository filters
	repoFilter := &repositories.CredentialFilter{
		Type:   filters.Type,
		Search: filters.Search,
		SortBy: filters.SortBy,
		Order:  filters.Order,
	}

	// Use access-controlled query - only returns credentials user can see
	credentials, total, err := h.credentialSvc.GetAccessibleByUser(r.Context(), wsCtx.WorkspaceID, claims.UserID, repoFilter, pg.Opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list credentials")
		return
	}

	// Build response with actions based on user permissions
	type CredentialWithActions struct {
		dto.CredentialResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}

	response := []CredentialWithActions{}
	wsID := wsCtx.WorkspaceID.String()

	for _, cred := range credentials {
		credID := cred.ID.String()
		basePath := "/api/v1/workspaces/" + wsID + "/credentials/" + credID

		// Build response with permission flags
		credResp := buildCredentialResponseWithPermissions(&cred, claims.UserID)

		// Only include edit/delete actions if user is owner
		var actions []dto.Action
		if credResp.CanEdit {
			actions = append(actions,
				dto.Action{Name: "edit", Method: "PUT", Href: basePath, Label: "Edit Credential"},
				dto.Action{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"},
			)
		}
		if credResp.CanShare {
			actions = append(actions,
				dto.Action{Name: "share", Method: "POST", Href: basePath + "/share", Label: "Share"},
			)
		}
		actions = append(actions,
			dto.Action{Name: "test", Method: "POST", Href: basePath + "/test", Label: "Test Connection"},
		)

		response = append(response, CredentialWithActions{
			CredentialResponse: credResp,
			Actions:            actions,
		})
	}

	// Build links with filter query string preservation
	basePath := "/api/v1/workspaces/" + wsID + "/credentials"
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

	// Apply field selection
	data := dto.SelectFields(r, response)

	dto.NewResponse(data).
		WithLinks(links).
		WithMeta(meta).
		Send(w)
}

func (h *CredentialHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	var req dto.CreateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	// Parse sharing scope
	var sharingScope models.SharingScope
	if req.SharingScope != nil {
		sharingScope = models.SharingScope(*req.SharingScope)
	}

	credential, err := h.credentialSvc.Create(r.Context(), services.CreateCredentialInput{
		WorkspaceID:  wsCtx.WorkspaceID,
		CreatedBy:    claims.UserID,
		Name:         req.Name,
		Type:         req.Type,
		Data:         req.Data,
		Description:  req.Description,
		SharingScope: sharingScope,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create credential")
		return
	}

	dto.Created(w, buildCredentialResponseWithPermissions(credential, claims.UserID))
}

func (h *CredentialHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	// Use access-controlled fetch
	credential, err := h.credentialSvc.GetByIDWithAccessCheck(r.Context(), credentialID, claims.UserID, true)
	if err != nil {
		if errors.Is(err, services.ErrCredentialAccessDenied) {
			dto.ErrorResponse(w, http.StatusForbidden, "access denied to this credential")
			return
		}
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}

	// SECURITY: Validate workspace ownership to prevent cross-tenant access
	if !ValidateWorkspaceOwnership(w, r, credential) {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	credID := credential.ID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/credentials/" + credID

	// Build response with permission flags
	credResp := buildCredentialResponseWithPermissions(credential, claims.UserID)

	// Build actions based on permissions
	var actions []dto.Action
	if credResp.CanEdit {
		actions = append(actions,
			dto.Action{Name: "edit", Method: "PUT", Href: basePath, Label: "Edit Credential"},
			dto.Action{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"},
		)
	}
	if credResp.CanShare {
		actions = append(actions,
			dto.Action{Name: "share", Method: "POST", Href: basePath + "/share", Label: "Share"},
			dto.Action{Name: "shares", Method: "GET", Href: basePath + "/shares", Label: "View Shares"},
		)
	}
	actions = append(actions,
		dto.Action{Name: "test", Method: "POST", Href: basePath + "/test", Label: "Test Connection"},
	)

	response := struct {
		dto.CredentialResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}{
		CredentialResponse: credResp,
		Actions:            actions,
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *CredentialHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	// SECURITY: First fetch and validate ownership before any modification
	existing, err := h.credentialSvc.GetByID(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	var req dto.UpdateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	// Use access-controlled update (only owner can edit)
	credential, err := h.credentialSvc.UpdateWithAccessCheck(r.Context(), credentialID, claims.UserID, services.UpdateCredentialInput{
		Name:        req.Name,
		Data:        req.Data,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, services.ErrCredentialEditDenied) {
			dto.ErrorResponse(w, http.StatusForbidden, "only the owner can edit this credential")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update credential")
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/credentials/" + credential.ID.String()

	dto.NewResponse(buildCredentialResponseWithPermissions(credential, claims.UserID)).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *CredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before deletion
	existing, err := h.credentialSvc.GetByID(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	// Use access-controlled delete (only owner can delete)
	if err := h.credentialSvc.DeleteWithAccessCheck(r.Context(), credentialID, claims.UserID); err != nil {
		if errors.Is(err, services.ErrCredentialEditDenied) {
			dto.ErrorResponse(w, http.StatusForbidden, "only the owner can delete this credential")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete credential")
		return
	}

	dto.NoContent(w)
}

func (h *CredentialHandler) Test(w http.ResponseWriter, r *http.Request) {
	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before testing
	existing, err := h.credentialSvc.GetByID(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	success, err := h.credentialSvc.TestConnection(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to test credential")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]bool{"success": success})
}

// Share shares a credential with specific users
func (h *CredentialHandler) Share(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	// Validate workspace ownership
	existing, err := h.credentialSvc.GetByID(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	var req dto.ShareCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	// Convert string UUIDs to uuid.UUID
	var userIDs []uuid.UUID
	for _, id := range req.UserIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			dto.ErrorResponse(w, http.StatusBadRequest, "invalid user ID: "+id)
			return
		}
		userIDs = append(userIDs, uid)
	}

	shares, err := h.credentialSvc.ShareCredential(r.Context(), services.ShareCredentialInput{
		CredentialID: credentialID,
		OwnerID:      claims.UserID,
		UserIDs:      userIDs,
	})
	if err != nil {
		if errors.Is(err, services.ErrCredentialShareDenied) {
			dto.ErrorResponse(w, http.StatusForbidden, "only the owner can share this credential")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to share credential")
		return
	}

	// Build response
	shareResponses := make([]dto.CredentialShareResponse, len(shares))
	for i, share := range shares {
		shareResponses[i] = buildCredentialShareResponse(&share)
	}

	dto.Created(w, shareResponses)
}

// Unshare removes sharing for a specific user
func (h *CredentialHandler) Unshare(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	userID, ok := middleware.ParseUUID(w, r, "userID")
	if !ok {
		return
	}

	// Validate workspace ownership
	existing, err := h.credentialSvc.GetByID(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	if err := h.credentialSvc.UnshareCredential(r.Context(), credentialID, claims.UserID, userID); err != nil {
		if errors.Is(err, services.ErrCredentialShareDenied) {
			dto.ErrorResponse(w, http.StatusForbidden, "only the owner can manage sharing")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to remove share")
		return
	}

	dto.NoContent(w)
}

// GetShares returns all shares for a credential
func (h *CredentialHandler) GetShares(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	// Validate workspace ownership
	existing, err := h.credentialSvc.GetByID(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	shares, err := h.credentialSvc.GetCredentialShares(r.Context(), credentialID, claims.UserID)
	if err != nil {
		if errors.Is(err, services.ErrCredentialShareDenied) {
			dto.ErrorResponse(w, http.StatusForbidden, "only the owner can view shares")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get shares")
		return
	}

	// Build response
	shareResponses := make([]dto.CredentialShareResponse, len(shares))
	for i, share := range shares {
		shareResponses[i] = buildCredentialShareResponse(&share)
	}

	dto.JSON(w, http.StatusOK, shareResponses)
}

// UpdateSharingScope updates the sharing scope of a credential
func (h *CredentialHandler) UpdateSharingScope(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	// Validate workspace ownership
	existing, err := h.credentialSvc.GetByID(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	var req dto.UpdateSharingScopeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	scope := models.SharingScope(req.SharingScope)
	if err := h.credentialSvc.UpdateSharingScope(r.Context(), credentialID, claims.UserID, scope); err != nil {
		if errors.Is(err, services.ErrCredentialEditDenied) {
			dto.ErrorResponse(w, http.StatusForbidden, "only the owner can change sharing scope")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update sharing scope")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]string{"sharing_scope": req.SharingScope})
}

// buildCredentialResponse creates a CredentialResponse from a Credential model (basic, no permissions)
func buildCredentialResponse(c *models.Credential) dto.CredentialResponse {
	return buildCredentialResponseWithPermissions(c, uuid.Nil)
}

// buildCredentialResponseWithPermissions creates a CredentialResponse with permission flags
func buildCredentialResponseWithPermissions(c *models.Credential, userID uuid.UUID) dto.CredentialResponse {
	var lastUsedAt *int64
	if c.LastUsedAt != nil {
		ts := c.LastUsedAt.Unix()
		lastUsedAt = &ts
	}

	var tokenExpiresAt *int64
	if c.TokenExpiresAt != nil {
		ts := c.TokenExpiresAt.Unix()
		tokenExpiresAt = &ts
	}

	// Build shares if available
	var shares []dto.CredentialShareResponse
	for _, share := range c.Shares {
		shares = append(shares, buildCredentialShareResponse(&share))
	}

	// Determine sharing scope string
	sharingScope := string(c.SharingScope)
	if sharingScope == "" {
		sharingScope = "workspace" // Default
	}

	return dto.CredentialResponse{
		ID:                c.ID.String(),
		WorkspaceID:       c.WorkspaceID.String(),
		CreatedBy:         c.CreatedBy.String(),
		Name:              c.Name,
		Type:              c.Type,
		Description:       c.Description,
		Provider:          c.Provider,
		ProviderAccountID: c.ProviderAccountID,
		TokenExpiresAt:    tokenExpiresAt,
		SharingScope:      sharingScope,
		IsOwner:           c.IsOwner(userID),
		CanEdit:           c.CanUserEdit(userID),
		CanShare:          c.CanUserShare(userID),
		Shares:            shares,
		LastUsedAt:        lastUsedAt,
		CreatedAt:         c.CreatedAt.Unix(),
		UpdatedAt:         c.UpdatedAt.Unix(),
	}
}

// buildCredentialShareResponse creates a CredentialShareResponse from a CredentialShare model
func buildCredentialShareResponse(s *models.CredentialShare) dto.CredentialShareResponse {
	resp := dto.CredentialShareResponse{
		ID:         s.ID.String(),
		UserID:     s.UserID.String(),
		Permission: s.Permission,
		SharedBy:   s.SharedBy.String(),
		CreatedAt:  s.CreatedAt.Unix(),
	}

	// Include user info if available
	if s.User.ID != uuid.Nil {
		resp.User = &dto.UserSummary{
			ID:        s.User.ID.String(),
			Email:     s.User.Email,
			FirstName: s.User.FirstName,
			LastName:  s.User.LastName,
		}
	}

	return resp
}
