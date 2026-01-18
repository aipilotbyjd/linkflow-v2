package auth

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	userCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/user"
)

// RegisterHandler handles user registration
type RegisterHandler struct {
	handler *userCmd.RegisterUserHandler
}

// NewRegisterHandler creates a new register handler
func NewRegisterHandler(handler *userCmd.RegisterUserHandler) *RegisterHandler {
	return &RegisterHandler{handler: handler}
}

// Handle handles the registration request
func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" {
		common.Error(w, 400, "EMAIL_REQUIRED", "Email is required")
		return
	}
	if req.Password == "" {
		common.Error(w, 400, "PASSWORD_REQUIRED", "Password is required")
		return
	}
	if len(req.Password) < 8 {
		common.Error(w, 400, "PASSWORD_TOO_SHORT", "Password must be at least 8 characters")
		return
	}
	if req.FirstName == "" {
		common.Error(w, 400, "FIRST_NAME_REQUIRED", "First name is required")
		return
	}
	if req.LastName == "" {
		common.Error(w, 400, "LAST_NAME_REQUIRED", "Last name is required")
		return
	}

	result, err := h.handler.Handle(r.Context(), userCmd.RegisterUserCommand{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, RegisterResponse{
		User:         ToUserResponse(result.User),
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
		ExpiresAt:    result.TokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		TokenType:    "Bearer",
	})
}
