package credential

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/oauth"
)

// OAuthProviderFactory defines interface to get provider by type
type OAuthProviderFactory interface {
	GetProvider(providerType string) *oauth.Provider
}

// RefreshService implements TokenRefresher
type RefreshService struct {
	encryptor EncryptionService
	providers map[string]*oauth.Provider
}

// NewRefreshService creates a new refresh service
func NewRefreshService(encryptor EncryptionService, providers map[string]*oauth.Provider) *RefreshService {
	return &RefreshService{
		encryptor: encryptor,
		providers: providers,
	}
}

// RefreshToken refreshes the token for a credential
func (s *RefreshService) RefreshToken(ctx context.Context, cred *credential.Credential) (string, *time.Time, error) {
	// Only OAuth2 credentials support refresh
	if cred.Type != credential.TypeOAuth2 {
		return "", nil, fmt.Errorf("credential type %s does not support refresh", cred.Type)
	}

	// decrypt data
	decryptedData, err := s.encryptor.Decrypt(cred.Data)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decrypt credential data: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(decryptedData), &data); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal credential data: %w", err)
	}

	refreshToken, ok := data["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return "", nil, fmt.Errorf("no refresh token found in credential")
	}

	// Get provider
	// The credential data usually contains "provider" field or we infer from credential name/type?
	// The domain credential doesn't explicitly store "provider name" separately from type "oauth2".
	// But `OAuthConnection` does. `Credential` is generic.
	// We assume `data["provider"]` exists or we need a way to map.
	// Let's check if `data` has provider.
	providerName, ok := data["provider"].(string)
	if !ok {
		return "", nil, fmt.Errorf("provider name not found in credential data")
	}

	provider, ok := s.providers[providerName]
	if !ok {
		return "", nil, fmt.Errorf("provider %s not configured", providerName)
	}

	// Refresh
	token, err := provider.RefreshToken(refreshToken)
	if err != nil {
		return "", nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update data
	data["access_token"] = token.AccessToken
	if token.RefreshToken != "" {
		data["refresh_token"] = token.RefreshToken
	}
	// Update expiry if provided
	var expiresAt *time.Time
	if token.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		expiresAt = &exp
		data["expires_at"] = exp
	}

	// Encrypt new data
	newDataBytes, err := json.Marshal(data)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal new data: %w", err)
	}

	newEncryptedData, err := s.encryptor.Encrypt(string(newDataBytes))
	if err != nil {
		return "", nil, fmt.Errorf("failed to encrypt new data: %w", err)
	}

	return newEncryptedData, expiresAt, nil
}
