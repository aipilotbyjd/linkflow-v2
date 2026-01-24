package config

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field    string
	Message  string
	Severity string // "error", "warning", "info"
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Severity, e.Field, e.Message)
}

// ValidationResult contains the results of configuration validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
	Info     []ValidationError
}

// HasErrors returns true if there are any error-level validation issues
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any warning-level validation issues
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// AddError adds an error to the result
func (r *ValidationResult) AddError(field, message string) {
	r.Errors = append(r.Errors, ValidationError{
		Field:    field,
		Message:  message,
		Severity: "error",
	})
}

// AddWarning adds a warning to the result
func (r *ValidationResult) AddWarning(field, message string) {
	r.Warnings = append(r.Warnings, ValidationError{
		Field:    field,
		Message:  message,
		Severity: "warning",
	})
}

// AddInfo adds an info message to the result
func (r *ValidationResult) AddInfo(field, message string) {
	r.Info = append(r.Info, ValidationError{
		Field:    field,
		Message:  message,
		Severity: "info",
	})
}

// ValidateConfiguration performs comprehensive validation of the configuration
func ValidateConfiguration(cfg *Config) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate individual sections
	validateAppConfig(cfg, result)
	validateDatabaseConfig(cfg, result)
	validateRedisConfig(cfg, result)
	validateJWTConfig(cfg, result)
	validateEmailConfig(cfg, result)
	validateOAuthConfig(cfg, result)
	validateAIConfig(cfg, result)
	validateStorageConfig(cfg, result)
	validateCryptoConfig(cfg, result)
	validateFeaturesConfig(cfg, result)

	// Validate cross-dependencies
	validateCrossDependencies(cfg, result)

	// Validate external dependencies (if not in dry-run mode)
	if os.Getenv("CONFIG_VALIDATION_SKIP_EXTERNAL") != "true" {
		validateExternalDependencies(cfg, result)
	}

	// Set overall validity
	result.Valid = !result.HasErrors()

	return result
}

// validateAppConfig validates application configuration
func validateAppConfig(cfg *Config, result *ValidationResult) {
	if cfg.App.Name == "" {
		result.AddError("app.name", "application name is required")
	}

	if cfg.App.Environment == "" {
		result.AddError("app.environment", "environment is required")
	}

	validEnvs := []string{"development", "staging", "production", "local", "test"}
	if !contains(validEnvs, cfg.App.Environment) {
		result.AddError("app.environment", "environment must be one of: "+strings.Join(validEnvs, ", "))
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		result.AddError("server.port", "port must be between 1 and 65535")
	}

	if len(cfg.App.CorsOrigins) == 0 {
		result.AddWarning("app.cors_origins", "no CORS origins specified, API may be inaccessible from browsers")
	}
}

// validateDatabaseConfig validates database configuration
func validateDatabaseConfig(cfg *Config, result *ValidationResult) {
	if cfg.Database.Host == "" {
		result.AddError("database.host", "database host is required")
	}

	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		result.AddError("database.port", "database port must be between 1 and 65535")
	}

	if cfg.Database.User == "" {
		result.AddError("database.user", "database user is required")
	}

	if cfg.Database.Name == "" {
		result.AddError("database.name", "database name is required")
	}

	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	if !contains(validSSLModes, cfg.Database.SSLMode) {
		result.AddError("database.ssl_mode", "SSL mode must be one of: "+strings.Join(validSSLModes, ", "))
	}

	if cfg.Database.MaxOpenConns <= 0 {
		result.AddWarning("database.max_open_conns", "max open connections should be positive")
	}

	if cfg.Database.MaxIdleConns <= 0 {
		result.AddWarning("database.max_idle_conns", "max idle connections should be positive")
	}

	if cfg.Database.MaxLifetime <= 0 {
		result.AddWarning("database.max_lifetime", "max lifetime should be positive")
	}
}

// validateRedisConfig validates Redis configuration
func validateRedisConfig(cfg *Config, result *ValidationResult) {
	if cfg.Redis.Host == "" {
		result.AddError("redis.host", "Redis host is required")
	}

	if cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535 {
		result.AddError("redis.port", "Redis port must be between 1 and 65535")
	}

	if cfg.Redis.DB < 0 || cfg.Redis.DB > 15 {
		result.AddWarning("redis.db", "Redis database should be between 0 and 15")
	}

	if cfg.Redis.PoolSize <= 0 {
		result.AddWarning("redis.pool_size", "Redis pool size should be positive")
	}
}

// validateJWTConfig validates JWT configuration
func validateJWTConfig(cfg *Config, result *ValidationResult) {
	if cfg.JWT.Secret == "" {
		result.AddError("jwt.secret", "JWT secret is required")
	}

	if len(cfg.JWT.Secret) < 32 {
		result.AddWarning("jwt.secret", "JWT secret should be at least 32 characters for security")
	}

	if cfg.JWT.AccessExpiry <= 0 {
		result.AddError("jwt.access_expiry", "JWT access expiry must be positive")
	}

	if cfg.JWT.RefreshExpiry <= 0 {
		result.AddError("jwt.refresh_expiry", "JWT refresh expiry must be positive")
	}

	if cfg.JWT.AccessExpiry >= cfg.JWT.RefreshExpiry {
		result.AddError("jwt.access_expiry", "JWT access expiry must be less than refresh expiry")
	}

	if cfg.JWT.Issuer == "" {
		result.AddWarning("jwt.issuer", "JWT issuer is not set")
	}
}

// validateEmailConfig validates email configuration
func validateEmailConfig(cfg *Config, result *ValidationResult) {
	if cfg.Email.Provider == "" {
		result.AddError("email.provider", "email provider is required")
	}

	validProviders := []string{"smtp", "sendgrid", "mailgun", "ses", "noop"}
	if !contains(validProviders, cfg.Email.Provider) {
		result.AddError("email.provider", "email provider must be one of: "+strings.Join(validProviders, ", "))
	}

	if cfg.Email.Provider == "smtp" {
		if cfg.Email.SMTPHost == "" {
			result.AddError("email.smtp_host", "SMTP host is required for SMTP provider")
		}
		if cfg.Email.SMTPPort <= 0 || cfg.Email.SMTPPort > 65535 {
			result.AddError("email.smtp_port", "SMTP port must be between 1 and 65535")
		}
	}

	if cfg.Email.From == "" {
		result.AddError("email.from", "from email is required")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(cfg.Email.From) {
		result.AddError("email.from", "from email must be a valid email address")
	}
}

// validateOAuthConfig validates OAuth configuration
func validateOAuthConfig(cfg *Config, result *ValidationResult) {
	// Google OAuth
	if cfg.OAuth.Google.ClientID != "" {
		if cfg.OAuth.Google.ClientSecret == "" {
			result.AddError("oauth.google.client_secret", "Google client secret is required when Google client ID is provided")
		}
		if cfg.OAuth.Google.RedirectURL == "" {
			result.AddError("oauth.google.redirect_url", "Google redirect URL is required when Google client ID is provided")
		} else {
			if err := validateURL(cfg.OAuth.Google.RedirectURL); err != nil {
				result.AddError("oauth.google.redirect_url", "Google redirect URL is invalid: "+err.Error())
			}
		}
		result.AddInfo("oauth.google", "Google OAuth is configured")
	}

	// GitHub OAuth
	if cfg.OAuth.GitHub.ClientID != "" {
		if cfg.OAuth.GitHub.ClientSecret == "" {
			result.AddError("oauth.github.client_secret", "GitHub client secret is required when GitHub client ID is provided")
		}
		if cfg.OAuth.GitHub.RedirectURL == "" {
			result.AddError("oauth.github.redirect_url", "GitHub redirect URL is required when GitHub client ID is provided")
		} else {
			if err := validateURL(cfg.OAuth.GitHub.RedirectURL); err != nil {
				result.AddError("oauth.github.redirect_url", "GitHub redirect URL is invalid: "+err.Error())
			}
		}
		result.AddInfo("oauth.github", "GitHub OAuth is configured")
	}
}

// validateAIConfig validates AI configuration
func validateAIConfig(cfg *Config, result *ValidationResult) {
	if cfg.AI.OpenAI.APIKey != "" {
		if cfg.AI.OpenAI.Model == "" {
			result.AddWarning("ai.openai.model", "OpenAI model not specified, using default")
		}
		result.AddInfo("ai.openai", "OpenAI is configured")
	}

	if cfg.AI.Anthropic.APIKey != "" {
		if cfg.AI.Anthropic.Model == "" {
			result.AddWarning("ai.anthropic.model", "Anthropic model not specified, using default")
		}
		result.AddInfo("ai.anthropic", "Anthropic is configured")
	}
}

// validateStorageConfig validates storage configuration
func validateStorageConfig(cfg *Config, result *ValidationResult) {
	validProviders := []string{"local", "s3"}
	if !contains(validProviders, cfg.Storage.Provider) {
		result.AddError("storage.provider", "storage provider must be one of: "+strings.Join(validProviders, ", "))
	}

	if cfg.Storage.Provider == "local" {
		if cfg.Storage.LocalPath == "" {
			result.AddWarning("storage.local_path", "local storage path not specified, using default")
		}
	}

	if cfg.Storage.Provider == "s3" {
		if cfg.Storage.S3.Bucket == "" {
			result.AddError("storage.s3.bucket", "S3 bucket is required for S3 storage")
		}
		if cfg.Storage.S3.Region == "" {
			result.AddError("storage.s3.region", "S3 region is required for S3 storage")
		}
		if cfg.Storage.S3.AccessKeyID == "" {
			result.AddError("storage.s3.access_key_id", "S3 access key ID is required for S3 storage")
		}
		if cfg.Storage.S3.SecretAccessKey == "" {
			result.AddError("storage.s3.secret_access_key", "S3 secret access key is required for S3 storage")
		}
	}
}

// validateCryptoConfig validates crypto configuration
func validateCryptoConfig(cfg *Config, result *ValidationResult) {
	if cfg.Crypto.EncryptionKey == "" {
		result.AddError("crypto.encryption_key", "encryption key is required")
	}

	if len(cfg.Crypto.EncryptionKey) < 32 {
		result.AddError("crypto.encryption_key", "encryption key must be at least 32 characters (256-bit)")
	}

	if cfg.Crypto.SigningKey == "" {
		result.AddWarning("crypto.signing_key", "signing key not specified")
	}
}

// validateFeaturesConfig validates features configuration
func validateFeaturesConfig(cfg *Config, result *ValidationResult) {
	// No feature flags to validate
}

// validateCrossDependencies validates interdependencies between configurations
func validateCrossDependencies(cfg *Config, result *ValidationResult) {
	// Webhook streaming requires Redis
	if cfg.Features.WebhookStream.Enabled && cfg.Redis.Host == "" {
		result.AddError("features.webhook_stream", "webhook streaming requires Redis configuration")
	}

	// Rate limiting requires Redis for distributed rate limiting
	if cfg.Features.RateLimit.Enabled && cfg.Redis.Host == "" {
		result.AddWarning("features.rate_limit", "rate limiting works better with Redis for distributed environments")
	}
}

// validateExternalDependencies validates external service connectivity
func validateExternalDependencies(cfg *Config, result *ValidationResult) {
	// Test database connectivity
	if err := testDatabaseConnection(cfg); err != nil {
		result.AddError("database", "database connection failed: "+err.Error())
	} else {
		result.AddInfo("database", "database connection successful")
	}

	// Test Redis connectivity
	if err := testRedisConnection(cfg); err != nil {
		result.AddError("redis", "Redis connection failed: "+err.Error())
	} else {
		result.AddInfo("redis", "Redis connection successful")
	}
}

// testDatabaseConnection tests database connectivity
func testDatabaseConnection(cfg *Config) error {
	// Simple connection test without using the full postgres client
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		strconv.Itoa(cfg.Database.Port),
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// testRedisConnection tests Redis connectivity
func testRedisConnection(cfg *Config) error {
	// Simple TCP connection test
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	conn.Close()

	return nil
}

// validateURL validates a URL format
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid URL format")
	}
	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateAndPrint validates configuration and prints results
func ValidateAndPrint(cfg *Config, log logger.Logger) bool {
	result := ValidateConfiguration(cfg)

	if result.HasErrors() {
		log.Error().Msg("Configuration validation failed:")
		for _, err := range result.Errors {
			log.Error().Str("field", err.Field).Msg(err.Message)
		}
		return false
	}

	if result.HasWarnings() {
		log.Warn().Msg("Configuration validation warnings:")
		for _, warning := range result.Warnings {
			log.Warn().Str("field", warning.Field).Msg(warning.Message)
		}
	}

	if len(result.Info) > 0 {
		log.Info().Msg("Configuration validation info:")
		for _, info := range result.Info {
			log.Info().Str("field", info.Field).Msg(info.Message)
		}
	}

	log.Info().Msg("Configuration validation successful")
	return true
}
