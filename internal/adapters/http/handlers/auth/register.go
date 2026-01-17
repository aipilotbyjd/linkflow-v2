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
