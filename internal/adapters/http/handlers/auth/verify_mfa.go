package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
)

type VerifyMFAHandler struct {
	userRepo user.Repository
	cache    cache.Cache
	otp      *crypto.OTP
}

func NewVerifyMFAHandler(userRepo user.Repository, cache cache.Cache) *VerifyMFAHandler {
	return &VerifyMFAHandler{
		userRepo: userRepo,
		cache:    cache,
		otp:      crypto.NewOTP("LinkFlow"),
	}
}

func (h *VerifyMFAHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req VerifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	// Get pending MFA secret from cache
	cacheKey := "mfa_setup:" + claims.UserID.String()
	secretBytes, err := h.cache.Get(r.Context(), cacheKey)
	if err != nil {
		common.BadRequest(w, "MFA setup not initiated or expired. Please start setup again.")
		return
	}
	secret := string(secretBytes)

	// Validate TOTP code
	if !h.otp.Validate(secret, req.Code) {
		common.BadRequest(w, "Invalid verification code")
		return
	}

	// Enable MFA for user
	if err := h.userRepo.EnableMFA(r.Context(), claims.UserID, secret); err != nil {
		common.HandleError(w, err)
		return
	}

	// Generate backup codes
	backupCodes, err := crypto.GenerateRecoveryCodes(10)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Delete the temporary secret from cache
	_ = h.cache.Delete(r.Context(), cacheKey)

	common.Success(w, map[string]interface{}{
		"message":      "MFA enabled successfully",
		"backup_codes": backupCodes,
	})
}
