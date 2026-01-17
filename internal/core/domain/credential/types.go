package credential

// Type represents credential type
type Type string

const (
	TypeAPIKey      Type = "api_key"
	TypeOAuth2      Type = "oauth2"
	TypeBasic       Type = "basic"
	TypeBearer      Type = "bearer"
	TypeCustom      Type = "custom"
	TypeDatabase    Type = "database"
	TypeSSH         Type = "ssh"
	TypeCertificate Type = "certificate"
)

func (t Type) String() string {
	return string(t)
}

func (t Type) IsValid() bool {
	switch t {
	case TypeAPIKey, TypeOAuth2, TypeBasic, TypeBearer, TypeCustom, TypeDatabase, TypeSSH, TypeCertificate:
		return true
	default:
		return false
	}
}

func ParseType(s string) (Type, bool) {
	t := Type(s)
	return t, t.IsValid()
}

// SharingScope defines who can access a credential
type SharingScope string

const (
	ScopePrivate  SharingScope = "private"
	ScopWorkspace SharingScope = "workspace"
	ScopeSpecific SharingScope = "specific"
)

func (s SharingScope) String() string {
	return string(s)
}

func (s SharingScope) IsValid() bool {
	switch s {
	case ScopePrivate, ScopWorkspace, ScopeSpecific:
		return true
	default:
		return false
	}
}

// Data represents the decrypted credential data structure
type Data struct {
	// Provider info (for OAuth)
	Provider string `json:"provider,omitempty"`

	// API Key
	APIKey string `json:"api_key,omitempty"`

	// OAuth2
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`

	// Basic Auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	// Bearer Token
	Token string `json:"token,omitempty"`

	// Database credentials
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database,omitempty"`

	// Connection string (for MongoDB etc.)
	ConnectionString string `json:"connectionString,omitempty"`

	// SSH credentials
	PrivateKey string `json:"private_key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`

	// Custom fields
	Custom map[string]string `json:"custom,omitempty"`

	// Generic data map for flexible access
	Fields map[string]interface{} `json:"data,omitempty"`
}

// GetField retrieves a custom field value
func (d *Data) GetField(key string) interface{} {
	if d.Fields == nil {
		return nil
	}
	return d.Fields[key]
}

// SetField sets a custom field value
func (d *Data) SetField(key string, value interface{}) {
	if d.Fields == nil {
		d.Fields = make(map[string]interface{})
	}
	d.Fields[key] = value
}
