package oauth

// OAuthProvider defines the OAuth provider interface
type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(code string) (Token, error)
	GetUserInfo(token Token) (UserInfo, error)
}

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

// ProviderInfo represents OAuth provider info
type ProviderInfo struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon,omitempty"`
}
