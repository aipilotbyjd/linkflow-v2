package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
)

// MustWorkspace returns the workspace context or writes an error response.
// Returns nil if workspace context is not available.
func MustWorkspace(w http.ResponseWriter, r *http.Request) *WorkspaceContext {
	wsCtx := GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.Forbidden(w, "workspace context required")
		return nil
	}
	return wsCtx
}

// MustUser returns the user claims or writes an error response.
// Returns nil if user is not authenticated.
func MustUser(w http.ResponseWriter, r *http.Request) *crypto.Claims {
	claims := GetUserFromContext(r.Context())
	if claims == nil {
		dto.Unauthorized(w, "authentication required")
		return nil
	}
	return claims
}

// MustUserAndWorkspace returns both user claims and workspace context.
// Returns nil, nil if either is not available.
func MustUserAndWorkspace(w http.ResponseWriter, r *http.Request) (*crypto.Claims, *WorkspaceContext) {
	claims := MustUser(w, r)
	if claims == nil {
		return nil, nil
	}
	wsCtx := MustWorkspace(w, r)
	if wsCtx == nil {
		return nil, nil
	}
	return claims, wsCtx
}

// ParseUUID parses a UUID from URL parameter.
// Returns uuid.Nil and writes error response if invalid.
func ParseUUID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		dto.BadRequest(w, "invalid "+param)
		return uuid.Nil, false
	}
	return id, true
}

// ParseUUIDQuery parses a UUID from query parameter.
// Returns nil if not provided, uuid.Nil with error if invalid.
func ParseUUIDQuery(w http.ResponseWriter, r *http.Request, param string) (*uuid.UUID, bool) {
	idStr := r.URL.Query().Get(param)
	if idStr == "" {
		return nil, true
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		dto.BadRequest(w, "invalid "+param)
		return nil, false
	}
	return &id, true
}
