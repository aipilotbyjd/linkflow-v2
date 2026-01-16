package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

// GetUserQuery represents the query to get a user by ID
type GetUserQuery struct {
	UserID uuid.UUID
}

// GetUserByEmailQuery represents the query to get a user by email
type GetUserByEmailQuery struct {
	Email string
}

// GetUserHandler handles getting users
type GetUserHandler struct {
	userRepo user.Repository
}

// NewGetUserHandler creates a new handler
func NewGetUserHandler(userRepo user.Repository) *GetUserHandler {
	return &GetUserHandler{userRepo: userRepo}
}

// Handle executes the get user query
func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*user.User, error) {
	return h.userRepo.FindByID(ctx, q.UserID)
}

// HandleByEmail gets a user by email
func (h *GetUserHandler) HandleByEmail(ctx context.Context, q GetUserByEmailQuery) (*user.User, error) {
	return h.userRepo.FindByEmail(ctx, q.Email)
}
