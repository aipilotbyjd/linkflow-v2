package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
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

// List returns all credentials for a workspace.
func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	pg := dto.ParsePagination(r)
	credentials, total, err := h.credentialSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, pg.Opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list credentials")
		return
	}

	response := []dto.CredentialResponse{}
	for _, cred := range credentials {
		var lastUsedAt *int64
		if cred.LastUsedAt != nil {
			ts := cred.LastUsedAt.Unix()
			lastUsedAt = &ts
		}

		response = append(response, dto.CredentialResponse{
			ID:          cred.ID.String(),
			Name:        cred.Name,
			Type:        cred.Type,
			Description: cred.Description,
			LastUsedAt:  lastUsedAt,
			CreatedAt:   cred.CreatedAt.Unix(),
		})
	}

	dto.JSONWithMeta(w, http.StatusOK, response, pg.NewMeta(total))
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

	credential, err := h.credentialSvc.Create(r.Context(), services.CreateCredentialInput{
		WorkspaceID: wsCtx.WorkspaceID,
		CreatedBy:   claims.UserID,
		Name:        req.Name,
		Type:        req.Type,
		Data:        req.Data,
		Description: req.Description,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create credential")
		return
	}

	dto.Created(w, dto.CredentialResponse{
		ID:          credential.ID.String(),
		Name:        credential.Name,
		Type:        credential.Type,
		Description: credential.Description,
		CreatedAt:   credential.CreatedAt.Unix(),
	})
}

func (h *CredentialHandler) Get(w http.ResponseWriter, r *http.Request) {
	credentialID, ok := middleware.ParseUUID(w, r, "credentialID")
	if !ok {
		return
	}

	credential, err := h.credentialSvc.GetByID(r.Context(), credentialID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "credential not found")
		return
	}

	// SECURITY: Validate workspace ownership to prevent cross-tenant access
	if !ValidateWorkspaceOwnership(w, r, credential) {
		return
	}

	var lastUsedAt *int64
	if credential.LastUsedAt != nil {
		ts := credential.LastUsedAt.Unix()
		lastUsedAt = &ts
	}

	dto.JSON(w, http.StatusOK, dto.CredentialResponse{
		ID:          credential.ID.String(),
		Name:        credential.Name,
		Type:        credential.Type,
		Description: credential.Description,
		LastUsedAt:  lastUsedAt,
		CreatedAt:   credential.CreatedAt.Unix(),
	})
}

func (h *CredentialHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	credential, err := h.credentialSvc.Update(r.Context(), credentialID, services.UpdateCredentialInput{
		Name:        req.Name,
		Data:        req.Data,
		Description: req.Description,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update credential")
		return
	}

	dto.JSON(w, http.StatusOK, dto.CredentialResponse{
		ID:          credential.ID.String(),
		Name:        credential.Name,
		Type:        credential.Type,
		Description: credential.Description,
		CreatedAt:   credential.CreatedAt.Unix(),
	})
}

func (h *CredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.credentialSvc.Delete(r.Context(), credentialID); err != nil {
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
