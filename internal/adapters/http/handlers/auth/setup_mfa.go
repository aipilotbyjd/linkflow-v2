package auth

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type SetupMFAResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

type SetupMFAHandler struct{}

func NewSetupMFAHandler() *SetupMFAHandler {
	return &SetupMFAHandler{}
}

func (h *SetupMFAHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user from context
	// userID := middleware.GetUserIDFromContext(r.Context())

	// TODO: Implement MFA setup
	// 1. Generate TOTP secret
	// 2. Generate QR code URL
	// 3. Store secret temporarily (not enabled yet)

	response := SetupMFAResponse{
		Secret:    "PLACEHOLDER_SECRET",
		QRCodeURL: "otpauth://totp/LinkFlow:user@example.com?secret=PLACEHOLDER_SECRET&issuer=LinkFlow",
	}

	common.Success(w, response)
}
