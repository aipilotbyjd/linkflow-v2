package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// RefreshHandler handles token refresh
type RefreshHandler struct {
	jwtManager *jwt.Manager
}

// NewRefreshHandler creates a new refresh handler
func NewRefreshHandler(jwtManager *jwt.Manager) *RefreshHandler {
	return &RefreshHandler{jwtManager: jwtManager}
}

// Handle handles the refresh request
func (h *RefreshHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	tokenPair, err := h.jwtManager.RefreshTokens(req.RefreshToken)
	if err != nil {
		common.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired refresh token")
		return
	}

	common.Success(w, RefreshResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		TokenType:    "Bearer",
	})
}
