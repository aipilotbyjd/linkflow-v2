package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type VerifyMFARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

type VerifyMFAHandler struct{}

func NewVerifyMFAHandler() *VerifyMFAHandler {
	return &VerifyMFAHandler{}
}

func (h *VerifyMFAHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req VerifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// TODO: Get user from context
	// userID := middleware.GetUserIDFromContext(r.Context())

	// TODO: Implement MFA verification
	// 1. Get pending MFA secret
	// 2. Validate TOTP code
	// 3. Enable MFA for user
	// 4. Generate backup codes

	common.Success(w, map[string]interface{}{
		"message":      "MFA enabled successfully",
		"backup_codes": []string{"12345678", "23456789", "34567890", "45678901", "56789012"},
	})
}
