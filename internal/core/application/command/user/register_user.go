package user

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

// RegisterUserCommand represents the command to register a new user
type RegisterUserCommand struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// RegisterUserResult contains the result of user registration
type RegisterUserResult struct {
	User      *user.User
	TokenPair *jwt.TokenPair
}

// RegisterUserHandler handles user registration
type RegisterUserHandler struct {
	userRepo   user.Repository
	jwtManager *jwt.Manager
	eventBus   events.Bus
}

// NewRegisterUserHandler creates a new handler
func NewRegisterUserHandler(
	userRepo user.Repository,
	jwtManager *jwt.Manager,
	eventBus events.Bus,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		eventBus:   eventBus,
	}
}

// Handle executes the register user command
func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (*RegisterUserResult, error) {
	// Validate input
	if cmd.Email == "" {
		return nil, user.ErrUserNotFound // Should be validation error
	}
	if cmd.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Check if email exists
	exists, err := h.userRepo.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, user.ErrEmailAlreadyExists
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	newUser := user.NewUser(cmd.Email, passwordHash, cmd.FirstName, cmd.LastName)

	if err := h.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(newUser.ID, newUser.Email, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Publish event
	if h.eventBus != nil {
		event := events.UserRegistered{
			BaseEvent: events.NewBaseEvent("user.registered", newUser.ID, "user"),
			UserID:    newUser.ID,
			Email:     newUser.Email,
			FirstName: newUser.FirstName,
			LastName:  newUser.LastName,
		}
		_ = h.eventBus.Publish(ctx, event)
	}

	return &RegisterUserResult{
		User:      newUser,
		TokenPair: tokenPair,
	}, nil
}
