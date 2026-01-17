package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/oauth"
)

type GitHubProvider struct {
	config oauth.Config
	client *http.Client
}

func NewGitHubProvider(config oauth.Config) *GitHubProvider {
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"user:email", "read:user"}
	}
	return &GitHubProvider{
		config: config,
		client: &http.Client{},
	}
}

func (p *GitHubProvider) Name() string {
	return "github"
}

func (p *GitHubProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("scope", strings.Join(p.config.Scopes, " "))
	params.Set("state", state)

	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (p *GitHubProvider) Exchange(ctx context.Context, code string) (*oauth.Token, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &oauth.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
	}, nil
}

func (p *GitHubProvider) GetUser(ctx context.Context, token *oauth.Token) (*oauth.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var userResp struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, err
	}

	email := userResp.Email
	if email == "" {
		email, _ = p.getPrimaryEmail(ctx, token)
	}

	firstName, lastName := splitName(userResp.Name)

	return &oauth.UserInfo{
		ID:        fmt.Sprintf("%d", userResp.ID),
		Email:     email,
		Name:      userResp.Name,
		FirstName: firstName,
		LastName:  lastName,
		AvatarURL: userResp.AvatarURL,
		Provider:  "github",
	}, nil
}

func (p *GitHubProvider) getPrimaryEmail(ctx context.Context, token *oauth.Token) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := p.client.Do(req)
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

func splitName(name string) (string, string) {
	parts := strings.SplitN(name, " ", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return name, ""
}
