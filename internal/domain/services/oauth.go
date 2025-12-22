package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

// OAuthProvider configuration for different OAuth providers
type OAuthProvider struct {
	Name         string
	AuthURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scopes       []string
	RedirectURL  string
}

// Common OAuth providers configuration
var OAuthProviders = map[string]OAuthProvider{
	"google": {
		Name:     "Google",
		AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/gmail.modify",
		},
	},
	"slack": {
		Name:     "Slack",
		AuthURL:  "https://slack.com/oauth/v2/authorize",
		TokenURL: "https://slack.com/api/oauth.v2.access",
		Scopes: []string{
			"chat:write",
			"channels:read",
			"users:read",
			"files:write",
		},
	},
	"github": {
		Name:     "GitHub",
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
		Scopes:   []string{"repo", "read:user", "read:org"},
	},
	"notion": {
		Name:     "Notion",
		AuthURL:  "https://api.notion.com/v1/oauth/authorize",
		TokenURL: "https://api.notion.com/v1/oauth/token",
		Scopes:   []string{},
	},
	"hubspot": {
		Name:     "HubSpot",
		AuthURL:  "https://app.hubspot.com/oauth/authorize",
		TokenURL: "https://api.hubapi.com/oauth/v1/token",
		Scopes:   []string{"crm.objects.contacts.read", "crm.objects.contacts.write"},
	},
	"salesforce": {
		Name:     "Salesforce",
		AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
		TokenURL: "https://login.salesforce.com/services/oauth2/token",
		Scopes:   []string{"api", "refresh_token"},
	},
	"stripe": {
		Name:     "Stripe",
		AuthURL:  "https://connect.stripe.com/oauth/authorize",
		TokenURL: "https://connect.stripe.com/oauth/token",
		Scopes:   []string{"read_write"},
	},
	"airtable": {
		Name:     "Airtable",
		AuthURL:  "https://airtable.com/oauth2/v1/authorize",
		TokenURL: "https://airtable.com/oauth2/v1/token",
		Scopes:   []string{"data.records:read", "data.records:write", "schema.bases:read"},
	},
}

type OAuthService struct {
	stateRepo      *repositories.OAuthStateRepository
	credentialRepo *repositories.CredentialRepository
	baseURL        string
	providers      map[string]OAuthProvider
}

func NewOAuthService(
	stateRepo *repositories.OAuthStateRepository,
	credentialRepo *repositories.CredentialRepository,
	baseURL string,
) *OAuthService {
	return &OAuthService{
		stateRepo:      stateRepo,
		credentialRepo: credentialRepo,
		baseURL:        strings.TrimSuffix(baseURL, "/"),
		providers:      OAuthProviders,
	}
}

// ConfigureProvider sets up OAuth credentials for a provider
func (s *OAuthService) ConfigureProvider(provider, clientID, clientSecret string) {
	if p, exists := s.providers[provider]; exists {
		p.ClientID = clientID
		p.ClientSecret = clientSecret
		p.RedirectURL = fmt.Sprintf("%s/api/v1/oauth/callback/%s", s.baseURL, provider)
		s.providers[provider] = p
	}
}

// GetAuthorizationURL generates an OAuth authorization URL
type AuthURLInput struct {
	Provider    string
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Scopes      []string // Optional override scopes
	RedirectURL string   // Optional redirect after completion
}

type AuthURLResult struct {
	URL   string `json:"url"`
	State string `json:"state"`
}

func (s *OAuthService) GetAuthorizationURL(ctx context.Context, input AuthURLInput) (*AuthURLResult, error) {
	provider, exists := s.providers[input.Provider]
	if !exists {
		return nil, fmt.Errorf("unknown OAuth provider: %s", input.Provider)
	}

	if provider.ClientID == "" {
		return nil, fmt.Errorf("OAuth provider %s not configured", input.Provider)
	}

	// Generate state token
	state, err := generateState()
	if err != nil {
		return nil, err
	}

	// Determine scopes
	scopes := provider.Scopes
	if len(input.Scopes) > 0 {
		scopes = input.Scopes
	}

	// Store state
	oauthState := &models.OAuthState{
		State:       state,
		UserID:      input.UserID,
		WorkspaceID: input.WorkspaceID,
		Provider:    input.Provider,
		RedirectURL: input.RedirectURL,
		Scopes:      scopes,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	if err := s.stateRepo.Create(ctx, oauthState); err != nil {
		return nil, fmt.Errorf("failed to save OAuth state: %w", err)
	}

	// Build authorization URL
	params := url.Values{
		"client_id":     {provider.ClientID},
		"redirect_uri":  {provider.RedirectURL},
		"response_type": {"code"},
		"state":         {state},
	}

	if len(scopes) > 0 {
		params.Set("scope", strings.Join(scopes, " "))
	}

	// Provider-specific parameters
	switch input.Provider {
	case "google":
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
	case "notion":
		params.Set("owner", "user")
	case "salesforce":
		params.Set("prompt", "consent")
	}

	authURL := fmt.Sprintf("%s?%s", provider.AuthURL, params.Encode())

	return &AuthURLResult{
		URL:   authURL,
		State: state,
	}, nil
}

// HandleCallback processes the OAuth callback
type CallbackInput struct {
	Provider string
	Code     string
	State    string
}

type TokenResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
}

func (s *OAuthService) HandleCallback(ctx context.Context, input CallbackInput) (*models.Credential, error) {
	// Validate state
	oauthState, err := s.stateRepo.FindByState(ctx, input.State)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired OAuth state")
	}

	provider, exists := s.providers[input.Provider]
	if !exists || provider.ClientID == "" {
		return nil, fmt.Errorf("OAuth provider not configured: %s", input.Provider)
	}

	// Exchange code for tokens
	tokens, err := s.exchangeCode(provider, input.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Create credential data
	credData := map[string]interface{}{
		"provider":     input.Provider,
		"access_token": tokens.AccessToken,
		"token_type":   tokens.TokenType,
		"scope":        tokens.Scope,
	}
	if tokens.RefreshToken != "" {
		credData["refresh_token"] = tokens.RefreshToken
	}
	if !tokens.ExpiresAt.IsZero() {
		credData["expires_at"] = tokens.ExpiresAt.Format(time.RFC3339)
	}

	// Serialize credential data to JSON string
	credDataJSON, err := json.Marshal(credData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize credential: %w", err)
	}

	credential := &models.Credential{
		WorkspaceID: oauthState.WorkspaceID,
		CreatedBy:   oauthState.UserID,
		Name:        fmt.Sprintf("%s OAuth", provider.Name),
		Type:        models.CredentialTypeOAuth2,
		Data:        string(credDataJSON),
	}

	if err := s.credentialRepo.Create(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to save credential: %w", err)
	}

	// Cleanup state
	_ = s.stateRepo.DeleteByState(ctx, input.State)

	return credential, nil
}

// RefreshToken refreshes an OAuth token
func (s *OAuthService) RefreshToken(ctx context.Context, credentialID uuid.UUID) (*models.Credential, error) {
	credential, err := s.credentialRepo.FindByID(ctx, credentialID)
	if err != nil {
		return nil, err
	}

	if credential.Type != models.CredentialTypeOAuth2 {
		return nil, fmt.Errorf("credential is not OAuth2 type")
	}

	// Parse existing credential data
	var credData map[string]interface{}
	if err := json.Unmarshal([]byte(credential.Data), &credData); err != nil {
		return nil, fmt.Errorf("failed to parse credential data: %w", err)
	}

	refreshToken, ok := credData["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Determine provider from credential data
	providerName := ""
	if name, ok := credData["provider"].(string); ok {
		providerName = name
	}

	provider, exists := s.providers[providerName]
	if !exists || provider.ClientID == "" {
		return nil, fmt.Errorf("OAuth provider not configured")
	}

	// Refresh the token
	tokens, err := s.refreshTokenRequest(provider, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update credential data
	credData["access_token"] = tokens.AccessToken
	if tokens.RefreshToken != "" {
		credData["refresh_token"] = tokens.RefreshToken
	}
	if !tokens.ExpiresAt.IsZero() {
		credData["expires_at"] = tokens.ExpiresAt.Format(time.RFC3339)
	}

	// Serialize back to JSON
	credDataJSON, err := json.Marshal(credData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize credential: %w", err)
	}
	credential.Data = string(credDataJSON)

	if err := s.credentialRepo.Update(ctx, credential); err != nil {
		return nil, err
	}

	return credential, nil
}

// GetSupportedProviders returns list of supported OAuth providers
func (s *OAuthService) GetSupportedProviders() []map[string]interface{} {
	providers := make([]map[string]interface{}, 0)
	for key, p := range s.providers {
		providers = append(providers, map[string]interface{}{
			"id":         key,
			"name":       p.Name,
			"configured": p.ClientID != "",
			"scopes":     p.Scopes,
		})
	}
	return providers
}

func (s *OAuthService) exchangeCode(provider OAuthProvider, code string) (*TokenResult, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {provider.RedirectURL},
		"client_id":     {provider.ClientID},
		"client_secret": {provider.ClientSecret},
	}

	return s.tokenRequest(provider.TokenURL, data)
}

func (s *OAuthService) refreshTokenRequest(provider OAuthProvider, refreshToken string) (*TokenResult, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {provider.ClientID},
		"client_secret": {provider.ClientSecret},
	}

	return s.tokenRequest(provider.TokenURL, data)
}

func (s *OAuthService) tokenRequest(tokenURL string, data url.Values) (*TokenResult, error) {
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	result := &TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
	}

	if tokenResp.ExpiresIn > 0 {
		result.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	return result, nil
}

func generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
