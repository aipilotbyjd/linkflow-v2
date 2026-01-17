package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
)

type contextKey string

const (
	UserContextKey      contextKey = "user"
	WorkspaceContextKey contextKey = "workspace"
	TokenContextKey     contextKey = "token"
	ClaimsContextKey    contextKey = "claims"
)

// UserClaims represents authenticated user information
type UserClaims struct {
	UserID      uuid.UUID
	Email       string
	WorkspaceID *uuid.UUID
}

// Auth creates an authentication middleware
func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				common.Unauthorized(w, "missing authorization header")
				return
			}

			// Extract token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				common.Unauthorized(w, "invalid authorization header format")
				return
			}
			token := parts[1]

			// Validate token
			claims, err := jwtManager.ValidateAccessToken(token)
			if err != nil {
				switch err {
				case jwt.ErrExpiredToken:
					common.Unauthorized(w, "token expired")
				case jwt.ErrInvalidToken, jwt.ErrInvalidType:
					common.Unauthorized(w, "invalid token")
				default:
					common.Unauthorized(w, "authentication failed")
				}
				return
			}

			// Set user and token in context
			userClaims := &UserClaims{
				UserID:      claims.UserID,
				Email:       claims.Email,
				WorkspaceID: claims.WorkspaceID,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
			ctx = context.WithValue(ctx, TokenContextKey, token)
			ctx = context.WithValue(ctx, ClaimsContextKey, claims)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth creates an optional authentication middleware
func OptionalAuth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := jwtManager.ValidateAccessToken(parts[1])
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			userClaims := &UserClaims{
				UserID:      claims.UserID,
				Email:       claims.Email,
				WorkspaceID: claims.WorkspaceID,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves user claims from context
func GetUserFromContext(ctx context.Context) *UserClaims {
	if claims, ok := ctx.Value(UserContextKey).(*UserClaims); ok {
		return claims
	}
	return nil
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	claims := GetUserFromContext(ctx)
	if claims != nil {
		return claims.UserID
	}
	return uuid.Nil
}

// GetUserID is an alias for GetUserIDFromContext
func GetUserID(ctx context.Context) uuid.UUID {
	return GetUserIDFromContext(ctx)
}

// RequireAuth ensures user is authenticated
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserFromContext(r.Context()) == nil {
			common.Unauthorized(w, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetTokenFromContext retrieves the raw token from context
func GetTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(TokenContextKey).(string); ok {
		return token
	}
	return ""
}

// GetClaimsFromContext retrieves raw JWT claims from context
func GetClaimsFromContext(ctx context.Context) *jwt.Claims {
	if claims, ok := ctx.Value(ClaimsContextKey).(*jwt.Claims); ok {
		return claims
	}
	return nil
}
