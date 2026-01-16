package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type DisableMFARequest struct {
	Code     string `json:"code" validate:"required,len=6"`
	Password string `json:"password" validate:"required"`
}

type DisableMFAHandler struct{}

func NewDisableMFAHandler() *DisableMFAHandler {
	return &DisableMFAHandler{}
}

func (h *DisableMFAHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req DisableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// TODO: Get user from context
	// userID := middleware.GetUserIDFromContext(r.Context())

	// TODO: Implement MFA disable
	// 1. Verify password
	// 2. Verify TOTP code
	// 3. Disable MFA for user
	// 4. Invalidate backup codes

	common.Success(w, map[string]string{
		"message": "MFA disabled successfully",
	})
}
