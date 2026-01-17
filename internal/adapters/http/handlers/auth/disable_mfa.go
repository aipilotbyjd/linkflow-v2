package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
)

type DisableMFARequest struct {
	Code     string `json:"code" validate:"required,len=6"`
	Password string `json:"password" validate:"required"`
}

type DisableMFAHandler struct {
	userRepo user.Repository
	otp      *crypto.OTP
}

func NewDisableMFAHandler(userRepo user.Repository) *DisableMFAHandler {
	return &DisableMFAHandler{
		userRepo: userRepo,
		otp:      crypto.NewOTP("LinkFlow"),
	}
}

func (h *DisableMFAHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req DisableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	// Get user
	u, err := h.userRepo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if !u.MFAEnabled {
		common.BadRequest(w, "MFA is not enabled")
		return
	}

	// Verify password
	password := user.NewPasswordFromHash(u.PasswordHash)
	if !password.Verify(req.Password) {
		common.BadRequest(w, "Invalid password")
		return
	}

	// Verify TOTP code
	if u.MFASecret == nil || !h.otp.Validate(*u.MFASecret, req.Code) {
		common.BadRequest(w, "Invalid verification code")
		return
	}

	// Disable MFA for user
	if err := h.userRepo.DisableMFA(r.Context(), claims.UserID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]string{
		"message": "MFA disabled successfully",
	})
}
