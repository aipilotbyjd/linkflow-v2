package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type APIKeyService struct {
	apiKeyRepo *repositories.APIKeyRepository
}

func NewAPIKeyService(apiKeyRepo *repositories.APIKeyRepository) *APIKeyService {
	return &APIKeyService{apiKeyRepo: apiKeyRepo}
}

type CreateAPIKeyInput struct {
	UserID      uuid.UUID
	WorkspaceID *uuid.UUID
	Name        string
	Scopes      []string
	ExpiresAt   *time.Time
}

type APIKeyInfo struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	KeyPrefix   string     `json:"key_prefix"`
	Scopes      []string   `json:"scopes"`
	WorkspaceID *uuid.UUID `json:"workspace_id,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateAPIKeyResult struct {
	APIKeyInfo
	Key string `json:"key"`
}

func (s *APIKeyService) Create(ctx context.Context, input CreateAPIKeyInput) (*CreateAPIKeyResult, error) {
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	keyHash := hashAPIKey(rawKey)
	keyPrefix := rawKey[:8]

	apiKey := &models.APIKey{
		UserID:      input.UserID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		KeyPrefix:   keyPrefix,
		KeyHash:     keyHash,
		Scopes:      input.Scopes,
		ExpiresAt:   input.ExpiresAt,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return &CreateAPIKeyResult{
		APIKeyInfo: APIKeyInfo{
			ID:          apiKey.ID,
			Name:        apiKey.Name,
			KeyPrefix:   apiKey.KeyPrefix,
			Scopes:      apiKey.Scopes,
			WorkspaceID: apiKey.WorkspaceID,
			ExpiresAt:   apiKey.ExpiresAt,
			CreatedAt:   apiKey.CreatedAt,
		},
		Key: rawKey,
	}, nil
}

func (s *APIKeyService) List(ctx context.Context, userID uuid.UUID) ([]APIKeyInfo, error) {
	keys, err := s.apiKeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]APIKeyInfo, len(keys))
	for i, k := range keys {
		result[i] = APIKeyInfo{
			ID:          k.ID,
			Name:        k.Name,
			KeyPrefix:   k.KeyPrefix,
			Scopes:      k.Scopes,
			WorkspaceID: k.WorkspaceID,
			LastUsedAt:  k.LastUsedAt,
			ExpiresAt:   k.ExpiresAt,
			CreatedAt:   k.CreatedAt,
		}
	}
	return result, nil
}

func (s *APIKeyService) Revoke(ctx context.Context, userID, keyID uuid.UUID) error {
	key, err := s.apiKeyRepo.FindByID(ctx, keyID)
	if err != nil {
		return ErrNotFound
	}

	if key.UserID != userID {
		return ErrForbidden
	}

	if key.RevokedAt != nil {
		return fmt.Errorf("API key already revoked")
	}

	return s.apiKeyRepo.Revoke(ctx, keyID)
}

func (s *APIKeyService) Validate(ctx context.Context, rawKey string) (*models.APIKey, error) {
	if !strings.HasPrefix(rawKey, "lf_") {
		return nil, fmt.Errorf("invalid API key format")
	}

	keyHash := hashAPIKey(rawKey)

	apiKey, err := s.apiKeyRepo.FindByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if apiKey.RevokedAt != nil {
		return nil, fmt.Errorf("API key has been revoked")
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	go func() {
		_ = s.apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)
	}()

	return apiKey, nil
}

func (s *APIKeyService) HasScope(apiKey *models.APIKey, requiredScope string) bool {
	if apiKey.Scopes == nil || len(apiKey.Scopes) == 0 {
		return true
	}

	for _, scope := range apiKey.Scopes {
		if scope == "*" || scope == requiredScope {
			return true
		}
		if strings.HasSuffix(scope, ":*") {
			prefix := strings.TrimSuffix(scope, "*")
			if strings.HasPrefix(requiredScope, prefix) {
				return true
			}
		}
	}
	return false
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "lf_" + hex.EncodeToString(bytes), nil
}

func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
