package oauth

import (
	"context"
)

type Provider interface {
	Name() string
	GetAuthURL(state string) string
	Exchange(ctx context.Context, code string) (*Token, error)
	GetUser(ctx context.Context, token *Token) (*UserInfo, error)
}

type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
}

type UserInfo struct {
	ID        string
	Email     string
	Name      string
	FirstName string
	LastName  string
	AvatarURL string
	Provider  string
}

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}
