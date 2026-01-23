package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
)

// CallbackHandler handles OAuth callback
type CallbackHandler struct {
	providers  map[string]OAuthProvider
	cache      cache.Cache
	userRepo   user.Repository
	oauthRepo  user.OAuthRepository
	jwtManager *jwt.Manager
}

// NewCallbackHandler creates a new handler
func NewCallbackHandler(
	providers map[string]OAuthProvider,
	cache cache.Cache,
	userRepo user.Repository,
	oauthRepo user.OAuthRepository,
	jwtManager *jwt.Manager,
) *CallbackHandler {
	return &CallbackHandler{
		providers:  providers,
		cache:      cache,
		userRepo:   userRepo,
		oauthRepo:  oauthRepo,
		jwtManager: jwtManager,
	}
}

// Handle handles the OAuth callback request
func (h *CallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providerID := chi.URLParam(r, "provider")

	provider, ok := h.providers[providerID]
	if !ok {
		common.NotFound(w, "OAuth provider")
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		errorMsg := r.URL.Query().Get("error")
		if errorMsg != "" {
			common.BadRequest(w, "OAuth error: "+errorMsg)
			return
		}
		common.BadRequest(w, "Authorization code not provided")
		return
	}

	// Validate state parameter
	cachedProvider, err := h.cache.Get(ctx, "oauth_state:"+state)
	if err != nil {
		if err == cache.ErrNotFound {
			common.BadRequest(w, "Invalid or expired state parameter")
			return
		}
		common.HandleError(w, err)
		return
	}

	if string(cachedProvider) != providerID {
		common.BadRequest(w, "State parameter mismatch")
		return
	}

	// Delete state from cache to prevent reuse
	_ = h.cache.Delete(ctx, "oauth_state:"+state)

	// Exchange code for token
	token, err := provider.ExchangeCode(code)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Find or create user
	var targetUser *user.User

	// 1. Check if OAuth connection exists
	conn, err := h.oauthRepo.FindByProviderAndProviderID(ctx, providerID, userInfo.ID)
	if err == nil && conn != nil {
		// Found connection, get user
		targetUser, err = h.userRepo.FindByID(ctx, conn.UserID)
		if err != nil {
			common.HandleError(w, err)
			return
		}
	} else if userInfo.Email != "" {
		// 2. Check if user exists by email
		targetUser, err = h.userRepo.FindByEmail(ctx, userInfo.Email)
		if err == nil && targetUser != nil {
			// User exists, create connection
			newConn := &user.OAuthConnection{
				ID:           uuid.New(),
				UserID:       targetUser.ID,
				Provider:     providerID,
				ProviderID:   userInfo.ID,
				Email:        &userInfo.Email,
				AccessToken:  &token.AccessToken,
				RefreshToken: &token.RefreshToken,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			if token.ExpiresIn > 0 {
				exp := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
				newConn.ExpiresAt = &exp
			}
			if err := h.oauthRepo.Create(ctx, newConn); err != nil {
				common.HandleError(w, err)
				return
			}
		}
	}

	// 3. Create new user if not found
	if targetUser == nil {
		if userInfo.Email == "" {
			common.BadRequest(w, "Email not provided by OAuth provider")
			return
		}

		// Generate random password
		randomPwdBytes := make([]byte, 16)
		_, _ = rand.Read(randomPwdBytes)
		randomPwd := hex.EncodeToString(randomPwdBytes)

		pwdHash, err := crypto.HashPassword(randomPwd)
		if err != nil {
			common.HandleError(w, err)
			return
		}

		// Parse name
		firstName := userInfo.Name
		lastName := ""
		// Simple name split, can be improved
		// ...

		targetUser, err = user.NewUser(userInfo.Email, pwdHash, firstName, lastName)
		if err != nil {
			common.HandleError(w, err)
			return
		}
		targetUser.EmailVerified = true // Trusted provider

		if err := h.userRepo.Create(ctx, targetUser); err != nil {
			common.HandleError(w, err)
			return
		}

		// Create connection
		newConn := &user.OAuthConnection{
			ID:           uuid.New(),
			UserID:       targetUser.ID,
			Provider:     providerID,
			ProviderID:   userInfo.ID,
			Email:        &userInfo.Email,
			AccessToken:  &token.AccessToken,
			RefreshToken: &token.RefreshToken,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if token.ExpiresIn > 0 {
			exp := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
			newConn.ExpiresAt = &exp
		}
		if err := h.oauthRepo.Create(ctx, newConn); err != nil {
			common.HandleError(w, err)
			return
		}
	}

	// Generate JWT tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(targetUser.ID, targetUser.Email, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    targetUser.ID,
			"email": targetUser.Email,
			"name":  targetUser.FullName(),
		},
		"tokens": tokenPair,
	})
}
