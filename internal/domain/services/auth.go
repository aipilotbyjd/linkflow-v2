package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Auth errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserLocked         = errors.New("account is locked")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidMFACode     = errors.New("invalid MFA code")
	ErrMFARequired        = errors.New("MFA verification required")
	ErrMFANotSetup        = errors.New("MFA not set up for this user")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrPasswordRequired   = errors.New("password is required")
	ErrEmailRequired      = errors.New("email is required")
)

// Auth configuration constants
const (
	MaxFailedLoginAttempts = 5
	AccountLockDuration    = 15 * time.Minute
	PasswordResetTokenTTL  = 1 * time.Hour
)

type AuthService struct {
	userRepo    *repositories.UserRepository
	sessionRepo *repositories.SessionRepository
	jwtManager  *crypto.JWTManager
	otpManager  *crypto.OTPManager
	encryptor   *crypto.Encryptor
}

// NewAuthService creates a new AuthService with required dependencies.
func NewAuthService(
	userRepo *repositories.UserRepository,
	sessionRepo *repositories.SessionRepository,
	jwtManager *crypto.JWTManager,
	otpManager *crypto.OTPManager,
	encryptor *crypto.Encryptor,
) *AuthService {
	if userRepo == nil || sessionRepo == nil || jwtManager == nil {
		panic("auth service: userRepo, sessionRepo, and jwtManager are required")
	}
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
		otpManager:  otpManager,
		encryptor:   encryptor,
	}
}

type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type LoginInput struct {
	Email    string
	Password string
	MFACode  string
	IP       string
	UserAgent string
}

type AuthResult struct {
	User        *models.User
	TokenPair   *crypto.TokenPair
	RequiresMFA bool
}

// Register creates a new user account and returns auth tokens.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	// Validate input
	if input.Email == "" {
		return nil, ErrEmailRequired
	}
	if input.Password == "" {
		return nil, ErrPasswordRequired
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	passwordHash, err := crypto.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:        input.Email,
		PasswordHash: passwordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Status:       models.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	log.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("User registered")

	return &AuthResult{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

// Login authenticates a user and returns auth tokens.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user.Status != models.UserStatusActive {
		return nil, ErrUserLocked
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, ErrUserLocked
	}

	if !crypto.CheckPassword(input.Password, user.PasswordHash) {
		// Track failed login attempt
		if err := s.userRepo.IncrementFailedLogins(ctx, user.ID); err != nil {
			log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("Failed to increment failed logins")
		}

		// Lock account if too many failed attempts
		if user.FailedLogins >= MaxFailedLoginAttempts-1 {
			lockUntil := time.Now().Add(AccountLockDuration)
			if err := s.userRepo.LockUser(ctx, user.ID, lockUntil); err != nil {
				log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to lock user account")
			} else {
				log.Warn().
					Str("user_id", user.ID.String()).
					Time("locked_until", lockUntil).
					Msg("User account locked due to failed login attempts")
			}
		}
		return nil, ErrInvalidCredentials
	}

	// Handle MFA if enabled
	if user.MFAEnabled {
		if input.MFACode == "" {
			return &AuthResult{
				User:        user,
				RequiresMFA: true,
			}, nil
		}

		if !s.otpManager.ValidateCode(*user.MFASecret, input.MFACode) {
			return nil, ErrInvalidMFACode
		}
	}

	// Update last login timestamp
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("Failed to update last login")
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &models.Session{
		UserID:    user.ID,
		TokenHash: hashToken(tokenPair.AccessToken),
		ExpiresAt: tokenPair.ExpiresAt,
		IPAddress: &input.IP,
		UserAgent: &input.UserAgent,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("Failed to create session")
	}

	log.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("User logged in")

	return &AuthResult{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

// Logout revokes a specific session.
func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessionRepo.RevokeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	log.Info().Str("session_id", sessionID.String()).Msg("Session revoked")
	return nil
}

// LogoutAll revokes all sessions for a user.
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessionRepo.RevokeAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke all sessions: %w", err)
	}
	log.Info().Str("user_id", userID.String()).Msg("All user sessions revoked")
	return nil
}

// RefreshToken generates new tokens using a refresh token.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*crypto.TokenPair, error) {
	tokenPair, err := s.jwtManager.RefreshTokens(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}
	return tokenPair, nil
}

// ValidateToken validates an access token and returns its claims.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*crypto.Claims, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}

// SetupMFA generates and stores MFA secret for a user.
func (s *AuthService) SetupMFA(ctx context.Context, userID uuid.UUID) (string, string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("%w: %s", ErrUserNotFound, userID)
	}

	secret, url, err := s.otpManager.GenerateSecret(user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate MFA secret: %w", err)
	}

	encryptedSecret, err := s.encryptor.Encrypt(secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt MFA secret: %w", err)
	}

	if err := s.userRepo.EnableMFA(ctx, userID, encryptedSecret); err != nil {
		return "", "", fmt.Errorf("failed to enable MFA: %w", err)
	}

	log.Info().Str("user_id", userID.String()).Msg("MFA setup initiated")

	return secret, url, nil
}

// VerifyMFA verifies a MFA code for a user.
func (s *AuthService) VerifyMFA(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrUserNotFound, userID)
	}

	if user.MFASecret == nil {
		return false, ErrMFANotSetup
	}

	secret, err := s.encryptor.Decrypt(*user.MFASecret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt MFA secret: %w", err)
	}

	return s.otpManager.ValidateCode(secret, code), nil
}

// DisableMFA disables MFA for a user after verifying the code.
func (s *AuthService) DisableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	valid, err := s.VerifyMFA(ctx, userID, code)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidMFACode
	}

	if err := s.userRepo.DisableMFA(ctx, userID); err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	log.Info().Str("user_id", userID.String()).Msg("MFA disabled")

	return nil
}

// ResetPasswordForUser resets a user's password (admin action).
func (s *AuthService) ResetPasswordForUser(ctx context.Context, userID uuid.UUID, newPassword string) error {
	if newPassword == "" {
		return ErrPasswordRequired
	}

	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := s.sessionRepo.RevokeAllUserSessions(ctx, userID); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("Failed to revoke sessions after password reset")
	}

	log.Info().Str("user_id", userID.String()).Msg("Password reset for user")

	return nil
}

// InitiatePasswordReset creates a password reset token and would send an email.
func (s *AuthService) InitiatePasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	// Generate reset token
	token := crypto.GenerateRandomToken(32)
	expiresAt := time.Now().Add(PasswordResetTokenTTL)

	// Store reset token
	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := s.userRepo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	// TODO: In production, send email with reset link here
	log.Info().
		Str("user_id", user.ID.String()).
		Str("email", email).
		Msg("Password reset initiated")

	return nil
}

// ResetPassword resets a user's password using a reset token.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if newPassword == "" {
		return ErrPasswordRequired
	}

	// Find and validate token
	resetToken, err := s.userRepo.FindPasswordResetToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	if resetToken.ExpiresAt.Before(time.Now()) {
		if err := s.userRepo.DeletePasswordResetToken(ctx, token); err != nil {
			log.Warn().Err(err).Msg("Failed to delete expired password reset token")
		}
		return ErrTokenExpired
	}

	if resetToken.UsedAt != nil {
		return ErrInvalidToken
	}

	// Hash new password
	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, passwordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.userRepo.MarkPasswordResetTokenUsed(ctx, token); err != nil {
		log.Warn().Err(err).Msg("Failed to mark password reset token as used")
	}

	// Revoke all sessions
	if err := s.sessionRepo.RevokeAllUserSessions(ctx, resetToken.UserID); err != nil {
		log.Warn().Err(err).Str("user_id", resetToken.UserID.String()).Msg("Failed to revoke sessions after password reset")
	}

	log.Info().Str("user_id", resetToken.UserID.String()).Msg("Password reset completed")

	return nil
}

func hashToken(token string) string {
	// SECURITY: Use proper cryptographic hashing instead of truncation
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
