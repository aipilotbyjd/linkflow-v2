package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/email"
)

type ForgotPasswordHandler struct {
	userRepo     user.Repository
	cache        cache.Cache
	emailService email.Provider
	baseURL      string
}

func NewForgotPasswordHandler(userRepo user.Repository, cache cache.Cache, emailService email.Provider, baseURL string) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{
		userRepo:     userRepo,
		cache:        cache,
		emailService: emailService,
		baseURL:      baseURL,
	}
}

func (h *ForgotPasswordHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// Find user by email (don't expose whether user exists)
	u, err := h.userRepo.FindByEmail(r.Context(), req.Email)
	if err == nil && u != nil {
		// Generate reset token
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err == nil {
			token := base64.URLEncoding.EncodeToString(tokenBytes)

			// Store token in cache (expires in 1 hour)
			cacheKey := "password_reset:" + token
			if err := h.cache.Set(r.Context(), cacheKey, []byte(u.ID.String()), time.Hour); err == nil {
				// Send reset email
				resetURL := h.baseURL + "/reset-password?token=" + token
				msg := &email.Message{
					To:      []string{u.Email},
					Subject: "Reset Your Password",
					HTMLBody: `<p>Hello ` + u.FirstName + `,</p>
<p>You requested to reset your password. Click the link below to reset it:</p>
<p><a href="` + resetURL + `">Reset Password</a></p>
<p>This link will expire in 1 hour.</p>
<p>If you didn't request this, you can ignore this email.</p>`,
					TextBody: "Hello " + u.FirstName + ",\n\nYou requested to reset your password. Visit the following link to reset it:\n\n" + resetURL + "\n\nThis link will expire in 1 hour.\n\nIf you didn't request this, you can ignore this email.",
				}
				_ = h.emailService.Send(r.Context(), msg)
			}
		}
	}

	// Always return success to prevent email enumeration
	common.Success(w, map[string]string{
		"message": "If an account exists with this email, a password reset link has been sent",
	})
}
