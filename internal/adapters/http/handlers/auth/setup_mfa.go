package auth

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
)

type SetupMFAHandler struct {
	userRepo user.Repository
	cache    cache.Cache
	otp      *crypto.OTP
}

func NewSetupMFAHandler(userRepo user.Repository, cache cache.Cache) *SetupMFAHandler {
	return &SetupMFAHandler{
		userRepo: userRepo,
		cache:    cache,
		otp:      crypto.NewOTP("LinkFlow"),
	}
}

func (h *SetupMFAHandler) Handle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	// Get user to check if MFA is already enabled
	u, err := h.userRepo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if u.MFAEnabled {
		common.BadRequest(w, "MFA is already enabled")
		return
	}

	// Generate TOTP secret
	secret, err := h.otp.GenerateSecret()
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Store secret temporarily in cache (expires in 10 minutes)
	cacheKey := "mfa_setup:" + claims.UserID.String()
	if err := h.cache.Set(r.Context(), cacheKey, []byte(secret), 10*time.Minute); err != nil {
		common.HandleError(w, err)
		return
	}

	// Generate QR code URL
	qrCodeURL := h.otp.GenerateURI(secret, u.Email)

	response := SetupMFAResponse{
		Secret:    secret,
		QRCodeURL: qrCodeURL,
	}

	common.Success(w, response)
}
