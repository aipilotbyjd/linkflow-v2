package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type APIKeyHandler struct {
	apiKeySvc *services.APIKeyService
}

func NewAPIKeyHandler(apiKeySvc *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeySvc: apiKeySvc}
}

type CreateAPIKeyRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Scopes      []string `json:"scopes,omitempty"`
	WorkspaceID *string  `json:"workspace_id,omitempty"`
	ExpiresIn   *int     `json:"expires_in_days,omitempty"`
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	input := services.CreateAPIKeyInput{
		UserID: claims.UserID,
		Name:   req.Name,
		Scopes: req.Scopes,
	}

	if req.WorkspaceID != nil {
		wsID, ok := middleware.ParseUUIDString(w, *req.WorkspaceID, "workspace_id")
		if !ok {
			return
		}
		input.WorkspaceID = &wsID
	}

	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expiresAt := time.Now().AddDate(0, 0, *req.ExpiresIn)
		input.ExpiresAt = &expiresAt
	}

	result, err := h.apiKeySvc.Create(r.Context(), input)
	if err != nil {
		dto.InternalServerError(w, "failed to create API key")
		return
	}

	dto.NewResponse(map[string]interface{}{
		"id":         result.ID,
		"name":       result.Name,
		"key":        result.Key,
		"key_prefix": result.KeyPrefix,
		"scopes":     result.Scopes,
		"expires_at": result.ExpiresAt,
		"created_at": result.CreatedAt,
		"message":    "Store this key securely. It will not be shown again.",
	}).Status(http.StatusCreated).Send(w)
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	keys, err := h.apiKeySvc.List(r.Context(), claims.UserID)
	if err != nil {
		dto.InternalServerError(w, "failed to list API keys")
		return
	}

	dto.NewResponse(keys).
		WithMeta(&dto.Meta{Total: int64(len(keys))}).
		Send(w)
}

func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	keyID, ok := middleware.ParseUUID(w, r, "keyID")
	if !ok {
		return
	}

	if err := h.apiKeySvc.Revoke(r.Context(), claims.UserID, keyID); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "API key")
			return
		}
		if err == services.ErrForbidden {
			dto.Forbidden(w, "cannot revoke this API key")
			return
		}
		dto.InternalServerError(w, "failed to revoke API key")
		return
	}

	dto.NewResponse(map[string]string{
		"message": "API key revoked successfully",
	}).Send(w)
}
