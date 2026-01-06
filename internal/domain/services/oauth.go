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
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/rs/zerolog/log"
)

// OAuthProvider configuration for different OAuth providers
type OAuthProvider struct {
	Name         string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	ClientID     string
	ClientSecret string
	Scopes       []string
	RedirectURL  string
}

// Common OAuth providers configuration (credentials populated from config)
var OAuthProviders = map[string]OAuthProvider{
	"google": {
		Name:        "Google",
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:    "https://oauth2.googleapis.com/token",
		UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/gmail.modify",
		},
	},
	"slack": {
		Name:        "Slack",
		AuthURL:     "https://slack.com/oauth/v2/authorize",
		TokenURL:    "https://slack.com/api/oauth.v2.access",
		UserInfoURL: "https://slack.com/api/users.identity",
		Scopes: []string{
			"chat:write",
			"channels:read",
			"users:read",
			"files:write",
		},
	},
	"github": {
		Name:        "GitHub",
		AuthURL:     "https://github.com/login/oauth/authorize",
		TokenURL:    "https://github.com/login/oauth/access_token",
		UserInfoURL: "https://api.github.com/user",
		Scopes:      []string{"repo", "read:user", "read:org"},
	},
	"notion": {
		Name:        "Notion",
		AuthURL:     "https://api.notion.com/v1/oauth/authorize",
		TokenURL:    "https://api.notion.com/v1/oauth/token",
		UserInfoURL: "",
		Scopes:      []string{},
	},
	"hubspot": {
		Name:        "HubSpot",
		AuthURL:     "https://app.hubspot.com/oauth/authorize",
		TokenURL:    "https://api.hubapi.com/oauth/v1/token",
		UserInfoURL: "",
		Scopes:      []string{"crm.objects.contacts.read", "crm.objects.contacts.write"},
	},
	"salesforce": {
		Name:        "Salesforce",
		AuthURL:     "https://login.salesforce.com/services/oauth2/authorize",
		TokenURL:    "https://login.salesforce.com/services/oauth2/token",
		UserInfoURL: "",
		Scopes:      []string{"api", "refresh_token"},
	},
	"stripe": {
		Name:        "Stripe",
		AuthURL:     "https://connect.stripe.com/oauth/authorize",
		TokenURL:    "https://connect.stripe.com/oauth/token",
		UserInfoURL: "",
		Scopes:      []string{"read_write"},
	},
	"airtable": {
		Name:        "Airtable",
		AuthURL:     "https://airtable.com/oauth2/v1/authorize",
		TokenURL:    "https://airtable.com/oauth2/v1/token",
		UserInfoURL: "https://api.airtable.com/v0/meta/whoami",
		Scopes:      []string{"data.records:read", "data.records:write", "schema.bases:read"},
	},
}

type OAuthService struct {
	stateRepo      *repositories.OAuthStateRepository
	credentialRepo *repositories.CredentialRepository
	encryptor      *crypto.Encryptor
	baseURL        string
	frontendURL    string
	providers      map[string]OAuthProvider
}

func NewOAuthService(
	stateRepo *repositories.OAuthStateRepository,
	credentialRepo *repositories.CredentialRepository,
	encryptor *crypto.Encryptor,
	baseURL string,
	frontendURL string,
) *OAuthService {
	// Make a copy of providers to avoid modifying the global map
	providers := make(map[string]OAuthProvider)
	for k, v := range OAuthProviders {
		providers[k] = v
	}

	return &OAuthService{
		stateRepo:      stateRepo,
		credentialRepo: credentialRepo,
		encryptor:      encryptor,
		baseURL:        strings.TrimSuffix(baseURL, "/"),
		frontendURL:    strings.TrimSuffix(frontendURL, "/"),
		providers:      providers,
	}
}

// ConfigureProvider sets up OAuth credentials for a provider
func (s *OAuthService) ConfigureProvider(provider, clientID, clientSecret string) {
	if p, exists := s.providers[provider]; exists {
		p.ClientID = clientID
		p.ClientSecret = clientSecret
		p.RedirectURL = fmt.Sprintf("%s/api/v1/oauth/callback/%s", s.baseURL, provider)
		s.providers[provider] = p
		log.Debug().Str("provider", provider).Msg("OAuth provider configured")
	}
}

// GetProvider returns a provider by name
func (s *OAuthService) GetProvider(name string) (*OAuthProvider, bool) {
	p, ok := s.providers[name]
	if !ok {
		return nil, false
	}
	return &p, true
}

// GetAuthorizationURL generates an OAuth authorization URL
type AuthURLInput struct {
	Provider       string
	UserID         uuid.UUID
	WorkspaceID    uuid.UUID
	CredentialName string   // Name for the credential to be created
	Scopes         []string // Optional override scopes
	RedirectURL    string   // Optional redirect after completion (frontend URL)
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
		return nil, fmt.Errorf("OAuth provider %s not configured (missing client credentials)", input.Provider)
	}

	// Generate state token
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Determine scopes
	scopes := provider.Scopes
	if len(input.Scopes) > 0 {
		scopes = input.Scopes
	}

	// Default redirect URL
	redirectURL := input.RedirectURL
	if redirectURL == "" {
		redirectURL = fmt.Sprintf("%s/credentials?oauth=success", s.frontendURL)
	}

	// Store state with metadata
	oauthState := &models.OAuthState{
		State:       state,
		UserID:      input.UserID,
		WorkspaceID: input.WorkspaceID,
		Provider:    input.Provider,
		RedirectURL: redirectURL,
		Scopes:      scopes,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	// Store credential name in redirect URL as query param
	if input.CredentialName != "" {
		parsedURL, _ := url.Parse(oauthState.RedirectURL)
		if parsedURL != nil {
			q := parsedURL.Query()
			q.Set("credential_name", input.CredentialName)
			parsedURL.RawQuery = q.Encode()
			oauthState.RedirectURL = parsedURL.String()
		}
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
	case "airtable":
		params.Set("code_challenge_method", "S256")
	}

	authURL := fmt.Sprintf("%s?%s", provider.AuthURL, params.Encode())

	log.Info().
		Str("provider", input.Provider).
		Str("workspace_id", input.WorkspaceID.String()).
		Str("state", state[:8]+"...").
		Msg("OAuth authorization URL generated")

	return &AuthURLResult{
		URL:   authURL,
		State: state,
	}, nil
}

// CallbackInput represents the OAuth callback parameters
type CallbackInput struct {
	Provider string
	Code     string
	State    string
}

// CallbackResult contains the result of the OAuth callback
type CallbackResult struct {
	Credential  *models.Credential
	RedirectURL string
	IsNew       bool
}

// TokenResult represents the tokens received from the OAuth provider
type TokenResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
}

// HandleCallback processes the OAuth callback
func (s *OAuthService) HandleCallback(ctx context.Context, input CallbackInput) (*CallbackResult, error) {
	// Validate state
	oauthState, err := s.stateRepo.FindByState(ctx, input.State)
	if err != nil {
		log.Warn().Str("state", input.State[:8]+"...").Msg("Invalid or expired OAuth state")
		return nil, fmt.Errorf("invalid or expired OAuth state")
	}

	provider, exists := s.providers[input.Provider]
	if !exists || provider.ClientID == "" {
		return nil, fmt.Errorf("OAuth provider not configured: %s", input.Provider)
	}

	// Exchange code for tokens
	tokens, err := s.exchangeCode(provider, input.Code)
	if err != nil {
		log.Error().Err(err).Str("provider", input.Provider).Msg("Failed to exchange OAuth code")
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Build credential data
	credData := models.CredentialData{
		Provider:     input.Provider,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		Scope:        tokens.Scope,
	}
	if !tokens.ExpiresAt.IsZero() {
		credData.ExpiresAt = tokens.ExpiresAt.Format(time.RFC3339)
	}

	// Serialize and encrypt credential data
	credDataJSON, err := json.Marshal(credData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize credential: %w", err)
	}

	encryptedData, err := s.encryptor.Encrypt(string(credDataJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// Extract credential name from redirect URL or use default
	credentialName := fmt.Sprintf("%s Connection", provider.Name)
	if parsedURL, _ := url.Parse(oauthState.RedirectURL); parsedURL != nil {
		if name := parsedURL.Query().Get("credential_name"); name != "" {
			credentialName = name
		}
	}

	// Create credential
	providerName := input.Provider
	var tokenExpiresAt *time.Time
	if !tokens.ExpiresAt.IsZero() {
		tokenExpiresAt = &tokens.ExpiresAt
	}

	credential := &models.Credential{
		WorkspaceID:    oauthState.WorkspaceID,
		CreatedBy:      oauthState.UserID,
		Name:           credentialName,
		Type:           models.CredentialTypeOAuth2,
		Data:           encryptedData,
		Provider:       &providerName,
		TokenExpiresAt: tokenExpiresAt,
	}

	if err := s.credentialRepo.Create(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to save credential: %w", err)
	}

	// Cleanup state
	_ = s.stateRepo.DeleteByState(ctx, input.State)

	// Build redirect URL with credential ID
	redirectURL := oauthState.RedirectURL
	if parsedURL, _ := url.Parse(redirectURL); parsedURL != nil {
		q := parsedURL.Query()
		q.Set("credential_id", credential.ID.String())
		q.Del("credential_name") // Remove the name param, no longer needed
		parsedURL.RawQuery = q.Encode()
		redirectURL = parsedURL.String()
	}

	log.Info().
		Str("credential_id", credential.ID.String()).
		Str("provider", input.Provider).
		Str("workspace_id", oauthState.WorkspaceID.String()).
		Msg("OAuth credential created successfully")

	return &CallbackResult{
		Credential:  credential,
		RedirectURL: redirectURL,
		IsNew:       true,
	}, nil
}

// RefreshToken refreshes an OAuth token for a credential
func (s *OAuthService) RefreshToken(ctx context.Context, credentialID uuid.UUID) (*models.Credential, error) {
	credential, err := s.credentialRepo.FindByID(ctx, credentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	if credential.Type != models.CredentialTypeOAuth2 {
		return nil, fmt.Errorf("credential is not OAuth2 type")
	}

	// Decrypt existing credential data
	decrypted, err := s.encryptor.Decrypt(credential.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential data: %w", err)
	}

	var credData models.CredentialData
	if err := json.Unmarshal([]byte(decrypted), &credData); err != nil {
		return nil, fmt.Errorf("failed to parse credential data: %w", err)
	}

	if credData.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Get provider
	providerName := credData.Provider
	if credential.Provider != nil {
		providerName = *credential.Provider
	}
	if providerName == "" {
		return nil, fmt.Errorf("unknown provider for credential")
	}

	provider, exists := s.providers[providerName]
	if !exists || provider.ClientID == "" {
		return nil, fmt.Errorf("OAuth provider not configured: %s", providerName)
	}

	// Refresh the token
	tokens, err := s.refreshTokenRequest(provider, credData.RefreshToken)
	if err != nil {
		log.Error().Err(err).
			Str("credential_id", credentialID.String()).
			Str("provider", providerName).
			Msg("Failed to refresh OAuth token")
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update credential data
	credData.AccessToken = tokens.AccessToken
	if tokens.RefreshToken != "" {
		credData.RefreshToken = tokens.RefreshToken
	}
	if !tokens.ExpiresAt.IsZero() {
		credData.ExpiresAt = tokens.ExpiresAt.Format(time.RFC3339)
	}

	// Re-encrypt
	credDataJSON, err := json.Marshal(credData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize credential: %w", err)
	}

	encryptedData, err := s.encryptor.Encrypt(string(credDataJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	credential.Data = encryptedData
	if !tokens.ExpiresAt.IsZero() {
		credential.TokenExpiresAt = &tokens.ExpiresAt
	}

	if err := s.credentialRepo.Update(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	log.Info().
		Str("credential_id", credentialID.String()).
		Str("provider", providerName).
		Msg("OAuth token refreshed successfully")

	return credential, nil
}

// RefreshExpiringTokens refreshes all tokens expiring within the given duration
func (s *OAuthService) RefreshExpiringTokens(ctx context.Context, within time.Duration) (int, int, error) {
	credentials, err := s.credentialRepo.FindExpiringTokens(ctx, within)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to find expiring tokens: %w", err)
	}

	var refreshed, failed int
	for _, cred := range credentials {
		_, err := s.RefreshToken(ctx, cred.ID)
		if err != nil {
			log.Warn().Err(err).
				Str("credential_id", cred.ID.String()).
				Msg("Failed to refresh expiring token")
			failed++
		} else {
			refreshed++
		}
	}

	if refreshed > 0 || failed > 0 {
		log.Info().
			Int("refreshed", refreshed).
			Int("failed", failed).
			Dur("within", within).
			Msg("Completed token refresh cycle")
	}

	return refreshed, failed, nil
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

// GetFrontendURL returns the configured frontend URL
func (s *OAuthService) GetFrontendURL() string {
	return s.frontendURL
}

func (s *OAuthService) exchangeCode(provider OAuthProvider, code string) (*TokenResult, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {provider.RedirectURL},
		"client_id":     {provider.ClientID},
		"client_secret": {provider.ClientSecret},
	}

	return s.tokenRequest(provider.TokenURL, data, provider.Name)
}

func (s *OAuthService) refreshTokenRequest(provider OAuthProvider, refreshToken string) (*TokenResult, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {provider.ClientID},
		"client_secret": {provider.ClientSecret},
	}

	return s.tokenRequest(provider.TokenURL, data, provider.Name)
}

func (s *OAuthService) tokenRequest(tokenURL string, data url.Values, providerName string) (*TokenResult, error) {
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Provider-specific headers
	switch providerName {
	case "Notion":
		// Notion requires Basic auth
		auth := base64.StdEncoding.EncodeToString([]byte(data.Get("client_id") + ":" + data.Get("client_secret")))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("OAuth error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
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
