package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidType  = errors.New("invalid token type")
)

// Config holds JWT configuration
type Config struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// Manager handles JWT token operations
type Manager struct {
	config Config
}

// Claims represents JWT claims
type Claims struct {
	UserID      uuid.UUID  `json:"user_id"`
	Email       string     `json:"email"`
	WorkspaceID *uuid.UUID `json:"workspace_id,omitempty"`
	Type        TokenType  `json:"type"`
	jwt.RegisteredClaims
}

// TokenType represents the type of JWT token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// NewManager creates a new JWT manager
func NewManager(config Config) *Manager {
	return &Manager{config: config}
}

// GenerateTokenPair generates a new access and refresh token pair
func (m *Manager) GenerateTokenPair(userID uuid.UUID, email string, workspaceID *uuid.UUID) (*TokenPair, error) {
	accessToken, accessExp, err := m.generateToken(userID, email, workspaceID, TokenTypeAccess, m.config.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := m.generateToken(userID, email, workspaceID, TokenTypeRefresh, m.config.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
		TokenType:    "Bearer",
	}, nil
}

// GenerateAccessToken generates only an access token
func (m *Manager) GenerateAccessToken(userID uuid.UUID, email string, workspaceID *uuid.UUID) (string, time.Time, error) {
	return m.generateToken(userID, email, workspaceID, TokenTypeAccess, m.config.AccessExpiry)
}

// GenerateRefreshToken generates only a refresh token
func (m *Manager) GenerateRefreshToken(userID uuid.UUID, email string, workspaceID *uuid.UUID) (string, time.Time, error) {
	return m.generateToken(userID, email, workspaceID, TokenTypeRefresh, m.config.RefreshExpiry)
}

func (m *Manager) generateToken(userID uuid.UUID, email string, workspaceID *uuid.UUID, tokenType TokenType, expiry time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := Claims{
		UserID:      userID,
		Email:       email,
		WorkspaceID: workspaceID,
		Type:        tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns its claims
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeAccess {
		return nil, ErrInvalidType
	}
	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeRefresh {
		return nil, ErrInvalidType
	}
	return claims, nil
}

// RefreshTokens refreshes the token pair using a refresh token
func (m *Manager) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	return m.GenerateTokenPair(claims.UserID, claims.Email, claims.WorkspaceID)
}

// GetTokenExpiry returns the access token expiry duration
func (m *Manager) GetTokenExpiry() time.Duration {
	return m.config.AccessExpiry
}

// GetRefreshExpiry returns the refresh token expiry duration
func (m *Manager) GetRefreshExpiry() time.Duration {
	return m.config.RefreshExpiry
}
