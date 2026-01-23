package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Token represents an OAuth token
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// UserInfo represents user info from OAuth provider
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ProviderConfig holds OAuth provider configuration
type ProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	RedirectURL  string
}

// Provider implements OAuthProvider interface
type Provider struct {
	config ProviderConfig
}

// NewProvider creates a new OAuth provider
func NewProvider(config ProviderConfig) *Provider {
	return &Provider{config: config}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return p.config.Name
}

// GetAuthURL returns the OAuth authorization URL
func (p *Provider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(p.config.Scopes, " "))
	params.Set("state", state)

	return p.config.AuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges authorization code for tokens
func (p *Provider) ExchangeCode(code string) (Token, error) {
	return p.makeTokenRequest(map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	})
}

// RefreshToken refreshes the access token using a refresh token
func (p *Provider) RefreshToken(refreshToken string) (Token, error) {
	return p.makeTokenRequest(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	})
}

func (p *Provider) makeTokenRequest(params map[string]string) (Token, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("redirect_uri", p.config.RedirectURL) // Some providers need this even for refresh

	for k, v := range params {
		data.Set(k, v)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", p.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return Token{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Token{}, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return Token{}, fmt.Errorf("failed to decode token: %w", err)
	}

	return token, nil
}

// GetUserInfo gets user info using the access token
func (p *Provider) GetUserInfo(token Token) (UserInfo, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", p.config.UserInfoURL, nil)
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return UserInfo{}, fmt.Errorf("get user info failed: %s", string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return UserInfo{}, fmt.Errorf("failed to decode user info: %w", err)
	}

	return userInfo, nil
}

// NewGoogleProvider creates a Google OAuth provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string) *Provider {
	return NewProvider(ProviderConfig{
		Name:         "google",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"email", "profile"},
		RedirectURL:  redirectURL,
	})
}

// NewGitHubProvider creates a GitHub OAuth provider
func NewGitHubProvider(clientID, clientSecret, redirectURL string) *Provider {
	return NewProvider(ProviderConfig{
		Name:         "github",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Scopes:       []string{"user:email"},
		RedirectURL:  redirectURL,
	})
}
