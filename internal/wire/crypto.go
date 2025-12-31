package wire

import (
	"fmt"

	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/oauth"
	"gorm.io/gorm"
)

// ProvideJWTManager creates the JWT manager
func ProvideJWTManager(cfg *config.Config) *crypto.JWTManager {
	return crypto.NewJWTManager(crypto.JWTConfig{
		Secret:        cfg.JWT.Secret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
		Issuer:        cfg.JWT.Issuer,
	})
}

// ProvideEncryptor creates the AES encryptor for credential storage
func ProvideEncryptor(cfg *config.Config) (*crypto.Encryptor, error) {
	key := cfg.Encryption.Key
	if key == "" {
		// Fallback to JWT secret for backward compatibility
		if len(cfg.JWT.Secret) >= 32 {
			key = cfg.JWT.Secret[:32]
		} else {
			return nil, fmt.Errorf("encryption key is required (set ENCRYPTION_KEY or ensure JWT_SECRET is at least 32 chars)")
		}
	}
	return crypto.NewEncryptor(key)
}

// ProvideOTPManager creates the OTP manager
func ProvideOTPManager(cfg *config.Config) *crypto.OTPManager {
	return crypto.NewOTPManager(cfg.App.Name)
}

// ProvideOAuthManager creates the OAuth manager with configured providers
func ProvideOAuthManager(cfg *config.Config, db *gorm.DB, encryptor *crypto.Encryptor) *oauth.Manager {
	manager := oauth.NewManager(db, encryptor)

	// Register configured providers
	if cfg.HasOAuth("google") {
		manager.RegisterProvider(oauth.NewGoogleProvider(
			cfg.OAuth.Google.ClientID,
			cfg.OAuth.Google.ClientSecret,
		))
	}

	if cfg.HasOAuth("github") {
		manager.RegisterProvider(oauth.NewGitHubProvider(
			cfg.OAuth.GitHub.ClientID,
			cfg.OAuth.GitHub.ClientSecret,
		))
	}

	if cfg.HasOAuth("microsoft") {
		manager.RegisterProvider(oauth.NewMicrosoftProvider(
			cfg.OAuth.Microsoft.ClientID,
			cfg.OAuth.Microsoft.ClientSecret,
			"", // tenantID - uses "common" by default
		))
	}

	return manager
}

// CryptoSet provides crypto dependencies
var CryptoSet = wire.NewSet(
	ProvideJWTManager,
	ProvideEncryptor,
	ProvideOTPManager,
	ProvideOAuthManager,
)
