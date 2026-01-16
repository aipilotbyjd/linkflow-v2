package user

import (
	"context"
	"fmt"
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

const (
	MaxFailedLoginAttempts = 5
	AccountLockDuration    = 15 * time.Minute
)

// LoginUserCommand represents the command to login a user
type LoginUserCommand struct {
	Email     string
	Password  string
	MFACode   string
	IPAddress string
	UserAgent string
}

// LoginUserResult contains the result of user login
type LoginUserResult struct {
	User        *user.User
	TokenPair   *jwt.TokenPair
	RequiresMFA bool
}

// LoginUserHandler handles user login
type LoginUserHandler struct {
	userRepo    user.Repository
	sessionRepo user.SessionRepository
	jwtManager  *jwt.Manager
	eventBus    events.Bus
}

// NewLoginUserHandler creates a new handler
func NewLoginUserHandler(
	userRepo user.Repository,
	sessionRepo user.SessionRepository,
	jwtManager *jwt.Manager,
	eventBus events.Bus,
) *LoginUserHandler {
	return &LoginUserHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
		eventBus:    eventBus,
	}
}

// Handle executes the login user command
func (h *LoginUserHandler) Handle(ctx context.Context, cmd LoginUserCommand) (*LoginUserResult, error) {
	// Find user
	u, err := h.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, user.ErrInvalidCredentials
	}

	// Check if user can login
	if !u.CanLogin() {
		if u.IsLocked() {
			return nil, user.ErrAccountLocked
		}
		return nil, user.ErrAccountSuspended
	}

	// Verify password
	if !crypto.CheckPassword(cmd.Password, u.PasswordHash) {
		// Track failed login
		_ = h.userRepo.IncrementFailedLogins(ctx, u.ID)

		// Lock account if too many failures
		if u.FailedLogins >= MaxFailedLoginAttempts-1 {
			lockUntil := time.Now().Add(AccountLockDuration)
			_ = h.userRepo.LockUser(ctx, u.ID, lockUntil)
		}

		return nil, user.ErrInvalidCredentials
	}

	// Check MFA
	if u.MFAEnabled {
		if cmd.MFACode == "" {
			return &LoginUserResult{
				User:        u,
				RequiresMFA: true,
			}, nil
		}
		// MFA validation would go here
		// For now, we'll skip the actual validation
	}

	// Update last login
	_ = h.userRepo.UpdateLastLogin(ctx, u.ID)

	// Generate tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(u.ID, u.Email, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := user.NewSession(u.ID, crypto.HashSHA256(tokenPair.AccessToken), tokenPair.ExpiresAt)
	session.WithDeviceInfo(cmd.IPAddress, cmd.UserAgent, nil)
	
	if err := h.sessionRepo.Create(ctx, session); err != nil {
		// Non-fatal, continue
	}

	// Publish event
	if h.eventBus != nil {
		event := events.UserLoggedIn{
			BaseEvent: events.NewBaseEvent("user.logged_in", u.ID, "user"),
			UserID:    u.ID,
			IPAddress: cmd.IPAddress,
			UserAgent: cmd.UserAgent,
		}
		_ = h.eventBus.Publish(ctx, event)
	}

	return &LoginUserResult{
		User:      u,
		TokenPair: tokenPair,
	}, nil
}
