package crypto

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
)

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

type JWTManager struct {
	config JWTConfig
}

type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	WorkspaceID *uuid.UUID `json:"workspace_id,omitempty"`
	Type        string    `json:"type"` // access or refresh
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewJWTManager(config JWTConfig) *JWTManager {
	return &JWTManager{config: config}
}

func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, email string, workspaceID *uuid.UUID) (*TokenPair, error) {
	accessToken, accessExp, err := m.generateToken(userID, email, workspaceID, "access", m.config.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := m.generateToken(userID, email, workspaceID, "refresh", m.config.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
	}, nil
}

func (m *JWTManager) generateToken(userID uuid.UUID, email string, workspaceID *uuid.UUID, tokenType string, expiry time.Duration) (string, time.Time, error) {
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

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
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

func (m *JWTManager) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	return m.GenerateTokenPair(claims.UserID, claims.Email, claims.WorkspaceID)
}
