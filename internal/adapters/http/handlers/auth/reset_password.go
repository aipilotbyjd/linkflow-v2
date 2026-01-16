package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type ResetPasswordHandler struct{}

func NewResetPasswordHandler() *ResetPasswordHandler {
	return &ResetPasswordHandler{}
}

func (h *ResetPasswordHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// TODO: Implement password reset
	// 1. Validate reset token
	// 2. Update user password
	// 3. Invalidate all sessions
	// 4. Mark token as used

	common.Success(w, map[string]string{
		"message": "Password has been reset successfully",
	})
}
