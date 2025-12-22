package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"gorm.io/gorm"
)

type Manager struct {
	db        *gorm.DB
	providers map[string]Provider
	crypto    *crypto.Encryptor
	states    sync.Map
	mu        sync.RWMutex
}

type StateData struct {
	Provider    string
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	RedirectURI string
	Scopes      []string
	Purpose     string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

func NewManager(db *gorm.DB, encryptor *crypto.Encryptor) *Manager {
	return &Manager{
		db:        db,
		providers: make(map[string]Provider),
		crypto:    encryptor,
	}
}

func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

func (m *Manager) GetProvider(name string) (Provider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.providers[name]
	return p, ok
}

func (m *Manager) GenerateState(data *StateData) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	data.CreatedAt = time.Now()
	data.ExpiresAt = time.Now().Add(10 * time.Minute)

	m.states.Store(state, data)

	go func() {
		time.Sleep(10 * time.Minute)
		m.states.Delete(state)
	}()

	return state, nil
}

func (m *Manager) ValidateState(state string) (*StateData, error) {
	v, ok := m.states.Load(state)
	if !ok {
		return nil, fmt.Errorf("invalid or expired state")
	}

	data := v.(*StateData)
	if time.Now().After(data.ExpiresAt) {
		m.states.Delete(state)
		return nil, fmt.Errorf("state expired")
	}

	m.states.Delete(state)
	return data, nil
}

func (m *Manager) GetAuthURL(providerName string, stateData *StateData, redirectURI string, scopes []string) (string, error) {
	provider, ok := m.GetProvider(providerName)
	if !ok {
		return "", fmt.Errorf("provider %s not found", providerName)
	}

	stateData.Provider = providerName
	stateData.RedirectURI = redirectURI
	stateData.Scopes = scopes

	state, err := m.GenerateState(stateData)
	if err != nil {
		return "", err
	}

	return provider.AuthURL(state, redirectURI, scopes), nil
}

func (m *Manager) HandleCallback(ctx context.Context, providerName, code, state string) (*OAuthResult, error) {
	stateData, err := m.ValidateState(state)
	if err != nil {
		return nil, err
	}

	if stateData.Provider != providerName {
		return nil, fmt.Errorf("provider mismatch")
	}

	provider, ok := m.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	token, err := provider.ExchangeCode(ctx, code, stateData.RedirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	userInfo, err := provider.GetUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	result := &OAuthResult{
		Provider:  providerName,
		Token:     token,
		UserInfo:  userInfo,
		StateData: stateData,
	}

	switch stateData.Purpose {
	case "login", "signup":
		if err := m.handleLoginSignup(ctx, result); err != nil {
			return nil, err
		}
	case "connect":
		if err := m.handleConnect(ctx, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

type OAuthResult struct {
	Provider     string
	Token        *Token
	UserInfo     *UserInfo
	StateData    *StateData
	User         *models.User
	Connection   *models.OAuthConnection
	IsNewUser    bool
}

func (m *Manager) handleLoginSignup(ctx context.Context, result *OAuthResult) error {
	var user models.User
	err := m.db.Where("email = ?", result.UserInfo.Email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		avatarURL := result.UserInfo.Picture
		user = models.User{
			Email:         result.UserInfo.Email,
			FirstName:     result.UserInfo.FirstName,
			LastName:      result.UserInfo.LastName,
			EmailVerified: result.UserInfo.EmailVerified,
			AvatarURL:     &avatarURL,
			Status:        "active",
		}
		if err := m.db.Create(&user).Error; err != nil {
			return err
		}
		result.IsNewUser = true
	} else if err != nil {
		return err
	}

	result.User = &user

	var conn models.OAuthConnection
	err = m.db.Where("user_id = ? AND provider = ?", user.ID, result.Provider).First(&conn).Error

	tokenData, _ := json.Marshal(result.Token)
	encryptedToken, _ := m.crypto.Encrypt(string(tokenData))

	refreshToken := result.Token.RefreshToken
	if err == gorm.ErrRecordNotFound {
		conn = models.OAuthConnection{
			UserID:       user.ID,
			Provider:     result.Provider,
			ProviderID:   result.UserInfo.ID,
			AccessToken:  &encryptedToken,
			RefreshToken: &refreshToken,
			ExpiresAt:    &result.Token.ExpiresAt,
		}
		if err := m.db.Create(&conn).Error; err != nil {
			return err
		}
	} else if err == nil {
		conn.AccessToken = &encryptedToken
		conn.RefreshToken = &refreshToken
		conn.ExpiresAt = &result.Token.ExpiresAt
		if err := m.db.Save(&conn).Error; err != nil {
			return err
		}
	} else {
		return err
	}

	result.Connection = &conn
	return nil
}

func (m *Manager) handleConnect(ctx context.Context, result *OAuthResult) error {
	if result.StateData.UserID == uuid.Nil {
		return fmt.Errorf("user ID required for connect")
	}

	var conn models.OAuthConnection
	err := m.db.Where("user_id = ? AND provider = ?", result.StateData.UserID, result.Provider).First(&conn).Error

	tokenData, _ := json.Marshal(result.Token)
	encryptedToken, _ := m.crypto.Encrypt(string(tokenData))
	refreshToken := result.Token.RefreshToken

	if err == gorm.ErrRecordNotFound {
		conn = models.OAuthConnection{
			UserID:       result.StateData.UserID,
			Provider:     result.Provider,
			ProviderID:   result.UserInfo.ID,
			AccessToken:  &encryptedToken,
			RefreshToken: &refreshToken,
			ExpiresAt:    &result.Token.ExpiresAt,
		}
		if err := m.db.Create(&conn).Error; err != nil {
			return err
		}
	} else if err == nil {
		conn.AccessToken = &encryptedToken
		conn.RefreshToken = &refreshToken
		conn.ExpiresAt = &result.Token.ExpiresAt
		if err := m.db.Save(&conn).Error; err != nil {
			return err
		}
	} else {
		return err
	}

	result.Connection = &conn
	return nil
}

func (m *Manager) RefreshUserToken(ctx context.Context, userID uuid.UUID, providerName string) (*Token, error) {
	var conn models.OAuthConnection
	if err := m.db.Where("user_id = ? AND provider = ?", userID, providerName).First(&conn).Error; err != nil {
		return nil, err
	}

	if conn.RefreshToken == nil || *conn.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	provider, ok := m.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	token, err := provider.RefreshToken(ctx, *conn.RefreshToken)
	if err != nil {
		return nil, err
	}

	tokenData, _ := json.Marshal(token)
	encryptedToken, _ := m.crypto.Encrypt(string(tokenData))

	conn.AccessToken = &encryptedToken
	if token.RefreshToken != "" {
		conn.RefreshToken = &token.RefreshToken
	}
	conn.ExpiresAt = &token.ExpiresAt

	if err := m.db.Save(&conn).Error; err != nil {
		return nil, err
	}

	return token, nil
}

func (m *Manager) GetUserToken(ctx context.Context, userID uuid.UUID, providerName string) (*Token, error) {
	var conn models.OAuthConnection
	if err := m.db.Where("user_id = ? AND provider = ?", userID, providerName).First(&conn).Error; err != nil {
		return nil, err
	}

	if conn.AccessToken == nil {
		return nil, fmt.Errorf("no access token available")
	}
	decrypted, err := m.crypto.Decrypt(*conn.AccessToken)
	if err != nil {
		return nil, err
	}

	var token Token
	if err := json.Unmarshal([]byte(decrypted), &token); err != nil {
		return nil, err
	}

	if conn.ExpiresAt != nil && time.Now().After(*conn.ExpiresAt) {
		return m.RefreshUserToken(ctx, userID, providerName)
	}

	return &token, nil
}

func (m *Manager) DisconnectProvider(ctx context.Context, userID uuid.UUID, providerName string) error {
	var conn models.OAuthConnection
	if err := m.db.Where("user_id = ? AND provider = ?", userID, providerName).First(&conn).Error; err != nil {
		return err
	}

	if conn.AccessToken != nil {
		decrypted, err := m.crypto.Decrypt(*conn.AccessToken)
		if err == nil {
			var token Token
			if json.Unmarshal([]byte(decrypted), &token) == nil {
				if provider, ok := m.GetProvider(providerName); ok {
					_ = provider.RevokeToken(ctx, token.AccessToken)
				}
			}
		}
	}

	return m.db.Delete(&conn).Error
}

func (m *Manager) ListConnections(ctx context.Context, userID uuid.UUID) ([]models.OAuthConnection, error) {
	var connections []models.OAuthConnection
	err := m.db.Where("user_id = ?", userID).Find(&connections).Error
	return connections, err
}
