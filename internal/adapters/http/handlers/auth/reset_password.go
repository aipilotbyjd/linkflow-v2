package auth

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type ResetPasswordHandler struct {
	userRepo    user.Repository
	sessionRepo user.SessionRepository
	cache       cache.Cache
}

func NewResetPasswordHandler(userRepo user.Repository, sessionRepo user.SessionRepository, cache cache.Cache) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cache:       cache,
	}
}

func (h *ResetPasswordHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
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

	// Validate reset token
	cacheKey := "password_reset:" + req.Token
	userIDBytes, err := h.cache.Get(r.Context(), cacheKey)
	if err != nil {
		common.BadRequest(w, "Invalid or expired reset token")
		return
	}
	userIDStr := string(userIDBytes)

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid reset token")
		return
	}

	// Validate password strength
	newPassword, err := user.NewPassword(req.NewPassword)
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}

	// Update user password
	if err := h.userRepo.UpdatePassword(r.Context(), userID, newPassword.Hash()); err != nil {
		common.HandleError(w, err)
		return
	}

	// Invalidate all sessions (non-fatal if fails)
	_ = h.sessionRepo.RevokeAllUserSessions(r.Context(), userID)

	// Delete the token from cache (mark as used)
	_ = h.cache.Delete(r.Context(), cacheKey)

	common.Success(w, map[string]string{
		"message": "Password has been reset successfully",
	})
}
