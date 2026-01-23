package oauth

// OAuthProvider defines the OAuth provider interface
type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(code string) (Token, error)
	RefreshToken(refreshToken string) (Token, error)
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

// ProviderInfo represents OAuth provider info for listing
type ProviderInfo struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon,omitempty"`
}

// ProviderDisplayNames maps provider IDs to display names
var ProviderDisplayNames = map[string]string{
	"google": "Google",
	"github": "GitHub",
}

// ProviderIcons maps provider IDs to icons
var ProviderIcons = map[string]string{
	"google": "https://www.google.com/favicon.ico",
	"github": "https://github.com/favicon.ico",
}
