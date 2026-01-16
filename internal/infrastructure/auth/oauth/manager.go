package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
)

// Manager manages OAuth providers
type Manager struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewManager creates a new OAuth manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers an OAuth provider
func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return provider, nil
}

// GetAuthURL returns the authorization URL for a provider
func (m *Manager) GetAuthURL(providerName string) (string, string, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return "", "", err
	}

	state, err := generateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	url := provider.GetAuthURL(state)
	return url, state, nil
}

// ExchangeCode exchanges an authorization code for tokens and user info
func (m *Manager) ExchangeCode(ctx context.Context, providerName, code string) (*UserInfo, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	token, err := provider.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	userInfo, err := provider.GetUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo, nil
}

// ListProviders returns a list of registered provider names
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}
	return providers
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
