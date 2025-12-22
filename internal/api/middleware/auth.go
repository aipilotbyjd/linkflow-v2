package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
)

type contextKey string

const (
	UserContextKey      contextKey = "user"
	WorkspaceContextKey contextKey = "workspace"
)

type AuthMiddleware struct {
	jwtManager  *crypto.JWTManager
	redisClient *pkgredis.Client
}

func NewAuthMiddleware(jwtManager *crypto.JWTManager, redisClient *pkgredis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:  jwtManager,
		redisClient: redisClient,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			dto.ErrorResponse(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		token := parts[1]
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			if err == crypto.ErrExpiredToken {
				dto.ErrorResponse(w, http.StatusUnauthorized, "token expired")
				return
			}
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid token")
			return
		}

		if claims.Type != "access" {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid token type")
			return
		}

		// Check if token is blacklisted (by JTI)
		if claims.ID != "" {
			blacklisted, err := m.redisClient.IsTokenBlacklisted(r.Context(), claims.ID)
			if err == nil && blacklisted {
				dto.ErrorResponse(w, http.StatusUnauthorized, "token has been revoked")
				return
			}
		}

		// Check if user logged out after token was issued
		logoutTime, err := m.redisClient.GetUserLogoutTime(r.Context(), claims.UserID.String())
		if err == nil && logoutTime > 0 && claims.IssuedAt != nil {
			if logoutTime > claims.IssuedAt.Unix() {
				dto.ErrorResponse(w, http.StatusUnauthorized, "token has been revoked")
				return
			}
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			next.ServeHTTP(w, r)
			return
		}

		token := parts[1]
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) *crypto.Claims {
	claims, ok := ctx.Value(UserContextKey).(*crypto.Claims)
	if !ok {
		return nil
	}
	return claims
}
