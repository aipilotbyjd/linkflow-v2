package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

type CreateAPIKeyHandler struct {
	apiKeyRepo user.APIKeyRepository
}

func NewCreateAPIKeyHandler(apiKeyRepo user.APIKeyRepository) *CreateAPIKeyHandler {
	return &CreateAPIKeyHandler{apiKeyRepo: apiKeyRepo}
}

type CreateAPIKeyRequest struct {
	Name      string   `json:"name"`
	ExpiresAt *string  `json:"expires_at,omitempty"`
	Scopes    []string `json:"scopes,omitempty"`
}

type CreateAPIKeyResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Key       string   `json:"key"`
	KeyPrefix string   `json:"key_prefix"`
	Scopes    []string `json:"scopes"`
	ExpiresAt *string  `json:"expires_at,omitempty"`
	CreatedAt string   `json:"created_at"`
}

func (h *CreateAPIKeyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		common.BadRequest(w, "name is required")
		return
	}

	// Generate API key
	rawKey := generateAPIKey()
	keyPrefix := rawKey[:8]
	keyHash := hashKey(rawKey)

	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{"*"}
	}

	apiKey := user.NewAPIKey(userClaims.UserID, req.Name, keyPrefix, keyHash, scopes)

	if req.ExpiresAt != nil {
		expiresAt, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			common.BadRequest(w, "invalid expires_at format, use RFC3339")
			return
		}
		apiKey = apiKey.WithExpiration(expiresAt)
	}

	if err := h.apiKeyRepo.Create(r.Context(), apiKey); err != nil {
		common.HandleError(w, err)
		return
	}

	response := CreateAPIKeyResponse{
		ID:        apiKey.ID.String(),
		Name:      apiKey.Name,
		Key:       "lf_" + rawKey,
		KeyPrefix: keyPrefix,
		Scopes:    apiKey.Scopes,
		CreatedAt: apiKey.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if apiKey.ExpiresAt != nil {
		s := apiKey.ExpiresAt.Format("2006-01-02T15:04:05Z")
		response.ExpiresAt = &s
	}

	common.Created(w, response)
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
