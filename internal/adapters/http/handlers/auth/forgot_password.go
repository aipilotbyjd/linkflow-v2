package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordHandler struct{}

func NewForgotPasswordHandler() *ForgotPasswordHandler {
	return &ForgotPasswordHandler{}
}

func (h *ForgotPasswordHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// TODO: Implement password reset token generation
	// 1. Find user by email
	// 2. Generate reset token
	// 3. Send reset email

	// Always return success to prevent email enumeration
	common.Success(w, map[string]string{
		"message": "If an account exists with this email, a password reset link has been sent",
	})
}
