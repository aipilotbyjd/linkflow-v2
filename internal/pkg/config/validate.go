package config

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Validate performs comprehensive configuration validation
func (c *Config) Validate() error {
	var errs []error

	// Core validations
	errs = append(errs, c.validateApp()...)
	errs = append(errs, c.validateServer()...)
	errs = append(errs, c.validateDatabase()...)
	errs = append(errs, c.validateRedis()...)
	errs = append(errs, c.validateJWT()...)
	errs = append(errs, c.validateOAuth()...)
	errs = append(errs, c.validateSMTP()...)
	errs = append(errs, c.validateS3()...)
	errs = append(errs, c.validateQueue()...)
	errs = append(errs, c.validateEncryption()...)
	errs = append(errs, c.validateTracing()...)
	errs = append(errs, c.validateRateLimit()...)
	errs = append(errs, c.validateCORS()...)
	errs = append(errs, c.validateLogging()...)
	errs = append(errs, c.validateFeatures()...)

	// Production-specific validations
	if c.IsProduction() {
		errs = append(errs, c.validateProduction()...)
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// ValidateRequired validates only critical configuration needed to start
func (c *Config) ValidateRequired() error {
	var errs []error

	if c.JWT.Secret == "" {
		errs = append(errs, errors.New("JWT_SECRET is required"))
	}
	if c.Database.Host == "" {
		errs = append(errs, errors.New("DATABASE_HOST is required"))
	}
	if c.Redis.Host == "" {
		errs = append(errs, errors.New("REDIS_HOST is required"))
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

func (c *Config) validateApp() []error {
	var errs []error

	if c.App.Name == "" {
		errs = append(errs, errors.New("app.name is required"))
	}

	validEnvs := []string{"development", "staging", "production", "test"}
	if !contains(validEnvs, c.App.Environment) {
		errs = append(errs, fmt.Errorf("app.environment must be one of: %s", strings.Join(validEnvs, ", ")))
	}

	if c.App.URL != "" {
		if _, err := url.Parse(c.App.URL); err != nil {
			errs = append(errs, fmt.Errorf("app.url is invalid: %w", err))
		}
	}

	if c.App.FrontendURL != "" {
		for _, u := range strings.Split(c.App.FrontendURL, ",") {
			if _, err := url.Parse(strings.TrimSpace(u)); err != nil {
				errs = append(errs, fmt.Errorf("app.frontend_url contains invalid URL '%s': %w", u, err))
			}
		}
	}

	if c.App.ExecutionRetentionDays < 1 || c.App.ExecutionRetentionDays > 365 {
		errs = append(errs, errors.New("app.execution_retention_days must be between 1 and 365"))
	}

	return errs
}

func (c *Config) validateServer() []error {
	var errs []error

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errs = append(errs, errors.New("server.port must be between 1 and 65535"))
	}

	if c.Server.ReadTimeout <= 0 {
		errs = append(errs, errors.New("server.read_timeout must be positive"))
	}

	if c.Server.WriteTimeout <= 0 {
		errs = append(errs, errors.New("server.write_timeout must be positive"))
	}

	if c.Server.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("server.shutdown_timeout must be positive"))
	}

	if c.Server.MaxHeaderBytes < 0 {
		errs = append(errs, errors.New("server.max_header_bytes cannot be negative"))
	}

	return errs
}

func (c *Config) validateDatabase() []error {
	var errs []error

	if c.Database.Host == "" {
		errs = append(errs, errors.New("database.host is required"))
	}

	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		errs = append(errs, errors.New("database.port must be between 1 and 65535"))
	}

	if c.Database.User == "" {
		errs = append(errs, errors.New("database.user is required"))
	}

	if c.Database.Name == "" {
		errs = append(errs, errors.New("database.name is required"))
	}

	validSSLModes := []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}
	if !contains(validSSLModes, c.Database.SSLMode) {
		errs = append(errs, fmt.Errorf("database.sslmode must be one of: %s", strings.Join(validSSLModes, ", ")))
	}

	if c.Database.MaxOpenConns < 1 {
		errs = append(errs, errors.New("database.max_open_conns must be at least 1"))
	}

	if c.Database.MaxIdleConns < 0 {
		errs = append(errs, errors.New("database.max_idle_conns cannot be negative"))
	}

	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		errs = append(errs, errors.New("database.max_idle_conns cannot exceed max_open_conns"))
	}

	if c.Database.ConnMaxLifetime < 0 {
		errs = append(errs, errors.New("database.conn_max_lifetime cannot be negative"))
	}

	return errs
}

func (c *Config) validateRedis() []error {
	var errs []error

	if c.Redis.Host == "" && !c.Redis.ClusterMode && len(c.Redis.SentinelAddrs) == 0 {
		errs = append(errs, errors.New("redis.host is required (or use cluster_mode/sentinel)"))
	}

	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		if !c.Redis.ClusterMode && len(c.Redis.SentinelAddrs) == 0 {
			errs = append(errs, errors.New("redis.port must be between 1 and 65535"))
		}
	}

	if c.Redis.DB < 0 || c.Redis.DB > 15 {
		errs = append(errs, errors.New("redis.db must be between 0 and 15"))
	}

	if c.Redis.PoolSize < 1 {
		errs = append(errs, errors.New("redis.pool_size must be at least 1"))
	}

	if c.Redis.MinIdleConns < 0 {
		errs = append(errs, errors.New("redis.min_idle_conns cannot be negative"))
	}

	if c.Redis.MinIdleConns > c.Redis.PoolSize {
		errs = append(errs, errors.New("redis.min_idle_conns cannot exceed pool_size"))
	}

	if c.Redis.MaxRetries < 0 {
		errs = append(errs, errors.New("redis.max_retries cannot be negative"))
	}

	if c.Redis.ClusterMode && len(c.Redis.ClusterAddrs) == 0 {
		errs = append(errs, errors.New("redis.cluster_addrs required when cluster_mode is enabled"))
	}

	if len(c.Redis.SentinelAddrs) > 0 && c.Redis.SentinelMaster == "" {
		errs = append(errs, errors.New("redis.sentinel_master required when sentinel_addrs is set"))
	}

	return errs
}

func (c *Config) validateJWT() []error {
	var errs []error

	if c.JWT.Secret == "" {
		errs = append(errs, errors.New("jwt.secret is required (set JWT_SECRET env var)"))
	} else if len(c.JWT.Secret) < 32 {
		errs = append(errs, errors.New("jwt.secret must be at least 32 characters"))
	}

	if c.JWT.Secret == "change-me-in-production" && c.IsProduction() {
		errs = append(errs, errors.New("jwt.secret must be changed from default in production"))
	}

	if c.JWT.AccessExpiry <= 0 {
		errs = append(errs, errors.New("jwt.access_expiry must be positive"))
	}

	if c.JWT.RefreshExpiry <= 0 {
		errs = append(errs, errors.New("jwt.refresh_expiry must be positive"))
	}

	if c.JWT.RefreshExpiry <= c.JWT.AccessExpiry {
		errs = append(errs, errors.New("jwt.refresh_expiry must be greater than access_expiry"))
	}

	if c.JWT.Issuer == "" {
		errs = append(errs, errors.New("jwt.issuer is required"))
	}

	validSigningMethods := []string{"HS256", "HS384", "HS512", "RS256", "RS384", "RS512"}
	if !contains(validSigningMethods, c.JWT.SigningMethod) {
		errs = append(errs, fmt.Errorf("jwt.signing_method must be one of: %s", strings.Join(validSigningMethods, ", ")))
	}

	return errs
}

func (c *Config) validateOAuth() []error {
	var errs []error

	// Google OAuth
	if c.OAuth.Google.ClientID != "" && c.OAuth.Google.ClientSecret == "" {
		errs = append(errs, errors.New("oauth.google.client_secret required when client_id is set"))
	}

	// GitHub OAuth
	if c.OAuth.GitHub.ClientID != "" && c.OAuth.GitHub.ClientSecret == "" {
		errs = append(errs, errors.New("oauth.github.client_secret required when client_id is set"))
	}

	// Microsoft OAuth
	if c.OAuth.Microsoft.ClientID != "" && c.OAuth.Microsoft.ClientSecret == "" {
		errs = append(errs, errors.New("oauth.microsoft.client_secret required when client_id is set"))
	}

	return errs
}

func (c *Config) validateSMTP() []error {
	var errs []error

	// Only validate if SMTP is configured
	if c.SMTP.Host == "" {
		return errs
	}

	if c.SMTP.Port <= 0 || c.SMTP.Port > 65535 {
		errs = append(errs, errors.New("smtp.port must be between 1 and 65535"))
	}

	if c.SMTP.From != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(c.SMTP.From) {
			errs = append(errs, errors.New("smtp.from must be a valid email address"))
		}
	}

	return errs
}

func (c *Config) validateS3() []error {
	var errs []error

	// Only validate if S3 is configured
	if c.S3.Bucket == "" && c.S3.Endpoint == "" {
		return errs
	}

	if c.S3.Bucket == "" {
		errs = append(errs, errors.New("s3.bucket is required when S3 is configured"))
	}

	if c.S3.Region == "" {
		errs = append(errs, errors.New("s3.region is required"))
	}

	if c.S3.PresignedExpiry <= 0 {
		errs = append(errs, errors.New("s3.presigned_expiry must be positive"))
	}

	if c.S3.UploadPartSize < 5*1024*1024 {
		errs = append(errs, errors.New("s3.upload_part_size must be at least 5MB"))
	}

	return errs
}

func (c *Config) validateQueue() []error {
	var errs []error

	if c.Queue.Concurrency < 1 {
		errs = append(errs, errors.New("queue.concurrency must be at least 1"))
	}

	if c.Queue.Concurrency > 1000 {
		errs = append(errs, errors.New("queue.concurrency should not exceed 1000"))
	}

	if c.Queue.RetryLimit < 0 {
		errs = append(errs, errors.New("queue.retry_limit cannot be negative"))
	}

	if c.Queue.Queues.Critical < 1 || c.Queue.Queues.Default < 1 || c.Queue.Queues.Low < 1 {
		errs = append(errs, errors.New("queue priorities must be at least 1"))
	}

	if c.Queue.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("queue.shutdown_timeout must be positive"))
	}

	return errs
}

func (c *Config) validateEncryption() []error {
	var errs []error

	// Only validate if encryption is configured
	if c.Encryption.Key == "" {
		return errs
	}

	// AES-256 requires a 32-byte key
	if c.Encryption.Algorithm == "aes-256-gcm" && len(c.Encryption.Key) != 32 {
		errs = append(errs, errors.New("encryption.key must be exactly 32 characters for AES-256-GCM"))
	}

	validAlgorithms := []string{"aes-256-gcm", "aes-128-gcm", "chacha20-poly1305"}
	if !contains(validAlgorithms, c.Encryption.Algorithm) {
		errs = append(errs, fmt.Errorf("encryption.algorithm must be one of: %s", strings.Join(validAlgorithms, ", ")))
	}

	return errs
}

func (c *Config) validateTracing() []error {
	var errs []error

	if !c.Tracing.Enabled {
		return errs
	}

	if c.Tracing.Endpoint == "" {
		errs = append(errs, errors.New("tracing.endpoint required when tracing is enabled"))
	}

	validProviders := []string{"otlp", "jaeger", "zipkin"}
	if !contains(validProviders, c.Tracing.Provider) {
		errs = append(errs, fmt.Errorf("tracing.provider must be one of: %s", strings.Join(validProviders, ", ")))
	}

	if c.Tracing.SampleRate < 0 || c.Tracing.SampleRate > 1 {
		errs = append(errs, errors.New("tracing.sample_rate must be between 0 and 1"))
	}

	return errs
}

func (c *Config) validateRateLimit() []error {
	var errs []error

	if !c.RateLimit.Enabled {
		return errs
	}

	if c.RateLimit.RequestsPerSecond < 1 {
		errs = append(errs, errors.New("rate_limit.requests_per_second must be at least 1"))
	}

	if c.RateLimit.Burst < c.RateLimit.RequestsPerSecond {
		errs = append(errs, errors.New("rate_limit.burst should be at least requests_per_second"))
	}

	return errs
}

func (c *Config) validateCORS() []error {
	var errs []error

	if !c.CORS.Enabled {
		return errs
	}

	if len(c.CORS.AllowedOrigins) == 0 {
		errs = append(errs, errors.New("cors.allowed_origins cannot be empty when CORS is enabled"))
	}

	// Validate origins are valid URLs (or wildcards)
	for _, origin := range c.CORS.AllowedOrigins {
		if origin != "*" && !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			errs = append(errs, fmt.Errorf("cors.allowed_origins contains invalid origin: %s", origin))
		}
	}

	if c.CORS.MaxAge < 0 {
		errs = append(errs, errors.New("cors.max_age cannot be negative"))
	}

	return errs
}

func (c *Config) validateLogging() []error {
	var errs []error

	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLevels, c.Logging.Level) {
		errs = append(errs, fmt.Errorf("logging.level must be one of: %s", strings.Join(validLevels, ", ")))
	}

	validFormats := []string{"json", "console", "text"}
	if !contains(validFormats, c.Logging.Format) {
		errs = append(errs, fmt.Errorf("logging.format must be one of: %s", strings.Join(validFormats, ", ")))
	}

	validOutputs := []string{"stdout", "stderr", "file"}
	if !contains(validOutputs, c.Logging.Output) {
		errs = append(errs, fmt.Errorf("logging.output must be one of: %s", strings.Join(validOutputs, ", ")))
	}

	if c.Logging.Output == "file" && c.Logging.File == "" {
		errs = append(errs, errors.New("logging.file required when output is 'file'"))
	}

	return errs
}

func (c *Config) validateFeatures() []error {
	var errs []error

	// Webhook Stream
	if c.Features.WebhookStream.Enabled {
		if c.Features.WebhookStream.MaxLen < 1000 {
			errs = append(errs, errors.New("features.webhook_stream.max_len must be at least 1000"))
		}
		if c.Features.WebhookStream.BatchSize < 1 {
			errs = append(errs, errors.New("features.webhook_stream.batch_size must be at least 1"))
		}
		if c.Features.WebhookStream.ConsumerCount < 1 {
			errs = append(errs, errors.New("features.webhook_stream.consumer_count must be at least 1"))
		}
	}

	// Execution
	if c.Features.Execution.MaxConcurrent < 1 {
		errs = append(errs, errors.New("features.execution.max_concurrent must be at least 1"))
	}
	if c.Features.Execution.DefaultTimeout <= 0 {
		errs = append(errs, errors.New("features.execution.default_timeout must be positive"))
	}
	if c.Features.Execution.MaxTimeout < c.Features.Execution.DefaultTimeout {
		errs = append(errs, errors.New("features.execution.max_timeout must be >= default_timeout"))
	}
	if c.Features.Execution.MaxPayloadSize < 1024 {
		errs = append(errs, errors.New("features.execution.max_payload_size must be at least 1KB"))
	}

	return errs
}

func (c *Config) validateProduction() []error {
	var errs []error

	// Security checks
	if c.App.Debug {
		errs = append(errs, errors.New("app.debug must be false in production"))
	}

	if c.Database.SSLMode == "disable" {
		errs = append(errs, errors.New("database.sslmode must not be 'disable' in production"))
	}

	if !c.Redis.TLS {
		// Warning only - some internal Redis deployments don't use TLS
		// errs = append(errs, errors.New("redis.tls should be enabled in production"))
	}

	if len(c.JWT.Secret) < 64 {
		errs = append(errs, errors.New("jwt.secret should be at least 64 characters in production"))
	}

	if !c.RateLimit.Enabled {
		errs = append(errs, errors.New("rate_limit should be enabled in production"))
	}

	// Check for wildcard CORS
	for _, origin := range c.CORS.AllowedOrigins {
		if origin == "*" {
			errs = append(errs, errors.New("cors.allowed_origins should not contain '*' in production"))
		}
	}

	// Logging checks
	if c.Logging.Level == "debug" || c.Logging.Level == "trace" {
		errs = append(errs, errors.New("logging.level should be 'info' or higher in production"))
	}

	if !c.Logging.RedactSecrets {
		errs = append(errs, errors.New("logging.redact_secrets should be true in production"))
	}

	// Encryption check
	if c.Encryption.Key == "" {
		errs = append(errs, errors.New("encryption.key should be set for credential storage in production"))
	}

	return errs
}

// =============================================================================
// Error Types
// =============================================================================

// ValidationError contains multiple validation errors
type ValidationError struct {
	Errors []error
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 1 {
		return fmt.Sprintf("config validation error: %s", e.Errors[0].Error())
	}
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("config validation errors (%d):\n", len(e.Errors)))
	for i, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return sb.String()
}

func (e *ValidationError) Unwrap() []error {
	return e.Errors
}

// =============================================================================
// Helper Functions
// =============================================================================

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
