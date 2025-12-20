package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Provider interface {
	Name() string
	AuthURL(state, redirectURI string, scopes []string) string
	ExchangeCode(ctx context.Context, code, redirectURI string) (*Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)
	RevokeToken(ctx context.Context, token string) error
}

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
	IDToken      string    `json:"id_token,omitempty"`
}

type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
}

type GoogleProvider struct {
	config ProviderConfig
}

func NewGoogleProvider(clientID, clientSecret string) *GoogleProvider {
	return &GoogleProvider{
		config: ProviderConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"openid", "email", "profile"},
		},
	}
}

func (p *GoogleProvider) Name() string {
	return "google"
}

func (p *GoogleProvider) AuthURL(state, redirectURI string, scopes []string) string {
	if len(scopes) == 0 {
		scopes = p.config.Scopes
	}

	params := url.Values{
		"client_id":     {p.config.ClientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {strings.Join(scopes, " ")},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}

	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (*Token, error) {
	data := url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}

func (p *GoogleProvider) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	data := url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	token.RefreshToken = refreshToken
	return &token, nil
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var data struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:            data.ID,
		Email:         data.Email,
		EmailVerified: data.VerifiedEmail,
		Name:          data.Name,
		FirstName:     data.GivenName,
		LastName:      data.FamilyName,
		Picture:       data.Picture,
		Locale:        data.Locale,
	}, nil
}

func (p *GoogleProvider) RevokeToken(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/revoke?token="+token, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke failed: %s", string(body))
	}

	return nil
}

type GitHubProvider struct {
	config ProviderConfig
}

func NewGitHubProvider(clientID, clientSecret string) *GitHubProvider {
	return &GitHubProvider{
		config: ProviderConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"user:email", "read:user"},
		},
	}
}

func (p *GitHubProvider) Name() string {
	return "github"
}

func (p *GitHubProvider) AuthURL(state, redirectURI string, scopes []string) string {
	if len(scopes) == 0 {
		scopes = p.config.Scopes
	}

	params := url.Values{
		"client_id":    {p.config.ClientID},
		"redirect_uri": {redirectURI},
		"scope":        {strings.Join(scopes, " ")},
		"state":        {state},
	}

	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (p *GitHubProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (*Token, error) {
	data := url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
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

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("github oauth error: %s", result.Error)
	}

	return &Token{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		Scope:       result.Scope,
	}, nil
}

func (p *GitHubProvider) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	return nil, fmt.Errorf("github does not support token refresh")
}

func (p *GitHubProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if data.Email == "" {
		email, _ := p.getPrimaryEmail(ctx, accessToken)
		data.Email = email
	}

	return &UserInfo{
		ID:            fmt.Sprintf("%d", data.ID),
		Email:         data.Email,
		EmailVerified: true,
		Name:          data.Name,
		Picture:       data.AvatarURL,
	}, nil
}

func (p *GitHubProvider) getPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", nil
}

func (p *GitHubProvider) RevokeToken(ctx context.Context, token string) error {
	return nil
}

type MicrosoftProvider struct {
	config   ProviderConfig
	tenantID string
}

func NewMicrosoftProvider(clientID, clientSecret, tenantID string) *MicrosoftProvider {
	if tenantID == "" {
		tenantID = "common"
	}
	return &MicrosoftProvider{
		config: ProviderConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"openid", "email", "profile", "User.Read"},
		},
		tenantID: tenantID,
	}
}

func (p *MicrosoftProvider) Name() string {
	return "microsoft"
}

func (p *MicrosoftProvider) AuthURL(state, redirectURI string, scopes []string) string {
	if len(scopes) == 0 {
		scopes = p.config.Scopes
	}

	params := url.Values{
		"client_id":     {p.config.ClientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {strings.Join(scopes, " ")},
		"state":         {state},
		"response_mode": {"query"},
	}

	return fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?%s", p.tenantID, params.Encode())
}

func (p *MicrosoftProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (*Token, error) {
	data := url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
		"scope":         {strings.Join(p.config.Scopes, " ")},
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", p.tenantID)
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}

func (p *MicrosoftProvider) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	data := url.Values{
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
		"scope":         {strings.Join(p.config.Scopes, " ")},
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", p.tenantID)
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}

func (p *MicrosoftProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var data struct {
		ID                string `json:"id"`
		DisplayName       string `json:"displayName"`
		GivenName         string `json:"givenName"`
		Surname           string `json:"surname"`
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	email := data.Mail
	if email == "" {
		email = data.UserPrincipalName
	}

	return &UserInfo{
		ID:            data.ID,
		Email:         email,
		EmailVerified: true,
		Name:          data.DisplayName,
		FirstName:     data.GivenName,
		LastName:      data.Surname,
	}, nil
}

func (p *MicrosoftProvider) RevokeToken(ctx context.Context, token string) error {
	return nil
}
