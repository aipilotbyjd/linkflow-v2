package auth

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
)

// LogoutHandler handles user logout
type LogoutHandler struct {
	sessionRepo user.SessionRepository
	blacklist   *jwt.Blacklist
}

// NewLogoutHandler creates a new logout handler
func NewLogoutHandler(sessionRepo user.SessionRepository, blacklist *jwt.Blacklist) *LogoutHandler {
	return &LogoutHandler{
		sessionRepo: sessionRepo,
		blacklist:   blacklist,
	}
}

// Handle handles the logout request
func (h *LogoutHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	// Get the token and full claims from context
	token := middleware.GetTokenFromContext(r.Context())
	jwtClaims := middleware.GetClaimsFromContext(r.Context())
	
	if token != "" && h.blacklist != nil && jwtClaims != nil {
		// Blacklist the current token
		_ = h.blacklist.AddWithExpiration(r.Context(), token, jwtClaims.ExpiresAt.Time)
	}

	// Revoke all sessions for the user (optional, could be just current session)
	if h.sessionRepo != nil {
		_ = h.sessionRepo.DeleteByUserID(r.Context(), userClaims.UserID)
	}

	common.Success(w, map[string]string{"message": "logged out successfully"})
}
