package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserLocked         = errors.New("account is locked")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidMFACode     = errors.New("invalid MFA code")
	ErrMFARequired        = errors.New("MFA verification required")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

type AuthService struct {
	userRepo    *repositories.UserRepository
	sessionRepo *repositories.SessionRepository
	jwtManager  *crypto.JWTManager
	otpManager  *crypto.OTPManager
	encryptor   *crypto.Encryptor
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	sessionRepo *repositories.SessionRepository,
	jwtManager *crypto.JWTManager,
	otpManager *crypto.OTPManager,
	encryptor *crypto.Encryptor,
) *AuthService {
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

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	passwordHash, err := crypto.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        input.Email,
		PasswordHash: passwordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Status:       models.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, nil)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if user.Status != models.UserStatusActive {
		return nil, ErrUserLocked
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, ErrUserLocked
	}

	if !crypto.CheckPassword(input.Password, user.PasswordHash) {
		s.userRepo.IncrementFailedLogins(ctx, user.ID)
		if user.FailedLogins >= 4 {
			lockUntil := time.Now().Add(15 * time.Minute)
			s.userRepo.LockUser(ctx, user.ID, lockUntil)
		}
		return nil, ErrInvalidCredentials
	}

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

	s.userRepo.UpdateLastLogin(ctx, user.ID)

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, nil)
	if err != nil {
		return nil, err
	}

	session := &models.Session{
		UserID:    user.ID,
		TokenHash: hashToken(tokenPair.AccessToken),
		ExpiresAt: tokenPair.ExpiresAt,
		IPAddress: &input.IP,
		UserAgent: &input.UserAgent,
	}
	s.sessionRepo.Create(ctx, session)

	return &AuthResult{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeSession(ctx, sessionID)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.RevokeAllUserSessions(ctx, userID)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*crypto.TokenPair, error) {
	return s.jwtManager.RefreshTokens(refreshToken)
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*crypto.Claims, error) {
	return s.jwtManager.ValidateToken(token)
}

func (s *AuthService) SetupMFA(ctx context.Context, userID uuid.UUID) (string, string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	secret, url, err := s.otpManager.GenerateSecret(user.Email)
	if err != nil {
		return "", "", err
	}

	encryptedSecret, err := s.encryptor.Encrypt(secret)
	if err != nil {
		return "", "", err
	}

	if err := s.userRepo.EnableMFA(ctx, userID, encryptedSecret); err != nil {
		return "", "", err
	}

	return secret, url, nil
}

func (s *AuthService) VerifyMFA(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if user.MFASecret == nil {
		return false, errors.New("MFA not set up")
	}

	secret, err := s.encryptor.Decrypt(*user.MFASecret)
	if err != nil {
		return false, err
	}

	return s.otpManager.ValidateCode(secret, code), nil
}

func (s *AuthService) DisableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	valid, err := s.VerifyMFA(ctx, userID, code)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidMFACode
	}

	return s.userRepo.DisableMFA(ctx, userID)
}

func (s *AuthService) ResetPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return err
	}

	return s.sessionRepo.RevokeAllUserSessions(ctx, userID)
}

func hashToken(token string) string {
	return token[:32]
}
