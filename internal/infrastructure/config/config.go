package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Email      EmailConfig      `mapstructure:"email"`
	OAuth      OAuthConfig      `mapstructure:"oauth"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Worker     WorkerConfig     `mapstructure:"worker"`
	Scheduler  SchedulerConfig  `mapstructure:"scheduler"`
	WebSocket  WebSocketConfig  `mapstructure:"websocket"`
	Crypto     CryptoConfig     `mapstructure:"crypto"`
	AI         AIConfig         `mapstructure:"ai"`
	Features   FeaturesConfig   `mapstructure:"features"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	Resilience ResilienceConfig `mapstructure:"resilience"`
	Sentry     SentryConfig     `mapstructure:"sentry"`
}

// SentryConfig holds Sentry error tracking settings
type SentryConfig struct {
	Enabled          bool    `mapstructure:"enabled"`
	DSN              string  `mapstructure:"dsn"`
	Environment      string  `mapstructure:"environment"`
	TracesSampleRate float64 `mapstructure:"traces_sample_rate"`
	Debug            bool    `mapstructure:"debug"`
}

// AppConfig holds application settings
type AppConfig struct {
	Name        string   `mapstructure:"name"`
	Environment string   `mapstructure:"environment"`
	Debug       bool     `mapstructure:"debug"`
	URL         string   `mapstructure:"url"`
	FrontendURL string   `mapstructure:"frontend_url"`
	SecretKey   string   `mapstructure:"secret_key"`
	CorsOrigins []string `mapstructure:"cors_origins"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	Name         string        `mapstructure:"name"`
	SSLMode      string        `mapstructure:"sslmode"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

// RedisConfig holds Redis settings
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig holds JWT settings
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
	Issuer        string        `mapstructure:"issuer"`
}

// EmailConfig holds email settings
type EmailConfig struct {
	Provider    string `mapstructure:"provider"` // smtp, sendgrid
	From        string `mapstructure:"from"`
	FromName    string `mapstructure:"from_name"`
	SMTPHost    string `mapstructure:"smtp_host"`
	SMTPPort    int    `mapstructure:"smtp_port"`
	SMTPUser    string `mapstructure:"smtp_user"`
	SMTPPass    string `mapstructure:"smtp_pass"`
	SendGridKey string `mapstructure:"sendgrid_key"`
}

// OAuthConfig holds OAuth provider settings
type OAuthConfig struct {
	Google    OAuthProviderConfig `mapstructure:"google"`
	GitHub    OAuthProviderConfig `mapstructure:"github"`
	Microsoft OAuthProviderConfig `mapstructure:"microsoft"`
}

// OAuthProviderConfig holds OAuth provider credentials
type OAuthProviderConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// FeaturesConfig holds feature flags and limits
type FeaturesConfig struct {
	WebhookStream WebhookStreamConfig `mapstructure:"webhook_stream"`
	RateLimit     RateLimitConfig     `mapstructure:"rate_limit"`
	Execution     ExecutionConfig     `mapstructure:"execution"`
}

// WebhookStreamConfig holds webhook streaming settings
type WebhookStreamConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	MaxLen        int  `mapstructure:"max_len"`
	BatchSize     int  `mapstructure:"batch_size"`
	MaxRetries    int  `mapstructure:"max_retries"`
	ConsumerCount int  `mapstructure:"consumer_count"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	BurstSize         int  `mapstructure:"burst_size"`
}

// ExecutionConfig holds execution settings
type ExecutionConfig struct {
	DefaultTimeout int `mapstructure:"default_timeout"`
	MaxRetries     int `mapstructure:"max_retries"`
	WorkerCount    int `mapstructure:"worker_count"`
	QueueSize      int `mapstructure:"queue_size"`
}

// MetricsConfig holds metrics settings
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// StorageConfig holds file storage settings
type StorageConfig struct {
	Provider  string   `mapstructure:"provider"` // local, s3
	LocalPath string   `mapstructure:"local_path"`
	S3        S3Config `mapstructure:"s3"`
}

// S3Config holds S3 storage settings
type S3Config struct {
	Bucket          string `mapstructure:"bucket"`
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Endpoint        string `mapstructure:"endpoint"`
	UsePathStyle    bool   `mapstructure:"use_path_style"`
}

// WorkerConfig holds worker settings
type WorkerConfig struct {
	Concurrency         int           `mapstructure:"concurrency"`
	Queues              []string      `mapstructure:"queues"`
	StrictPriority      bool          `mapstructure:"strict_priority"`
	ShutdownTimeout     time.Duration `mapstructure:"shutdown_timeout"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	RetryLimit          int           `mapstructure:"retry_limit"`
}

// SchedulerConfig holds scheduler settings
type SchedulerConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	PollInterval   time.Duration `mapstructure:"poll_interval"`
	LockTTL        time.Duration `mapstructure:"lock_ttl"`
	BatchSize      int           `mapstructure:"batch_size"`
	LeaderElection bool          `mapstructure:"leader_election"`
	HealthPort     int           `mapstructure:"health_port"`
}

// WebSocketConfig holds WebSocket settings
type WebSocketConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	PingInterval    time.Duration `mapstructure:"ping_interval"`
	PongTimeout     time.Duration `mapstructure:"pong_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ReadBufferSize  int           `mapstructure:"read_buffer_size"`
	WriteBufferSize int           `mapstructure:"write_buffer_size"`
	MaxMessageSize  int64         `mapstructure:"max_message_size"`
}

// CryptoConfig holds encryption settings
type CryptoConfig struct {
	EncryptionKey string `mapstructure:"encryption_key"` // 32 bytes for AES-256
	SigningKey    string `mapstructure:"signing_key"`
}

// AIConfig holds AI provider settings
type AIConfig struct {
	OpenAI    AIProviderConfig `mapstructure:"openai"`
	Anthropic AIProviderConfig `mapstructure:"anthropic"`
}

// AIProviderConfig holds AI provider configuration
type AIProviderConfig struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

// ResilienceConfig holds resilience settings
type ResilienceConfig struct {
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
	Retry          RetryConfig          `mapstructure:"retry"`
	RateLimit      RateLimiterConfig    `mapstructure:"rate_limit"`
}

// CircuitBreakerConfig holds circuit breaker settings
type CircuitBreakerConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	MaxRequests      uint32        `mapstructure:"max_requests"`
	Interval         time.Duration `mapstructure:"interval"`
	Timeout          time.Duration `mapstructure:"timeout"`
	FailureThreshold uint32        `mapstructure:"failure_threshold"`
}

// RetryConfig holds retry settings
type RetryConfig struct {
	MaxAttempts     int           `mapstructure:"max_attempts"`
	InitialInterval time.Duration `mapstructure:"initial_interval"`
	MaxInterval     time.Duration `mapstructure:"max_interval"`
	Multiplier      float64       `mapstructure:"multiplier"`
}

// RateLimiterConfig holds rate limiter settings
type RateLimiterConfig struct {
	Enabled bool    `mapstructure:"enabled"`
	Rate    float64 `mapstructure:"rate"`
	Burst   int     `mapstructure:"burst"`
}

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind OAuth environment variables explicitly
	_ = v.BindEnv("oauth.google.client_id", "OAUTH_GOOGLE_CLIENT_ID")
	_ = v.BindEnv("oauth.google.client_secret", "OAUTH_GOOGLE_CLIENT_SECRET")
	_ = v.BindEnv("oauth.google.redirect_url", "OAUTH_GOOGLE_REDIRECT_URL")
	_ = v.BindEnv("oauth.github.client_id", "OAUTH_GITHUB_CLIENT_ID")
	_ = v.BindEnv("oauth.github.client_secret", "OAUTH_GITHUB_CLIENT_SECRET")
	_ = v.BindEnv("oauth.github.redirect_url", "OAUTH_GITHUB_REDIRECT_URL")
	_ = v.BindEnv("oauth.microsoft.client_id", "OAUTH_MICROSOFT_CLIENT_ID")
	_ = v.BindEnv("oauth.microsoft.client_secret", "OAUTH_MICROSOFT_CLIENT_SECRET")
	_ = v.BindEnv("oauth.microsoft.redirect_url", "OAUTH_MICROSOFT_REDIRECT_URL")

	// Bind AI environment variables
	_ = v.BindEnv("ai.openai.api_key", "AI_OPENAI_API_KEY")
	_ = v.BindEnv("ai.anthropic.api_key", "AI_ANTHROPIC_API_KEY")

	// Bind JWT environment variables
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")

	// Bind Crypto environment variables
	_ = v.BindEnv("crypto.encryption_key", "CRYPTO_ENCRYPTION_KEY")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if result := ValidateConfiguration(&cfg); !result.Valid {
		// Collect all error messages
		var errorMsgs []string
		for _, err := range result.Errors {
			errorMsgs = append(errorMsgs, err.Message)
		}
		return nil, fmt.Errorf("configuration validation failed: %s", strings.Join(errorMsgs, "; "))
	}

	return &cfg, nil
}

// LoadWithValidation loads configuration and returns validation result
func LoadWithValidation(configPath string) (*Config, *ValidationResult, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind OAuth environment variables explicitly
	_ = v.BindEnv("oauth.google.client_id", "OAUTH_GOOGLE_CLIENT_ID")
	_ = v.BindEnv("oauth.google.client_secret", "OAUTH_GOOGLE_CLIENT_SECRET")
	_ = v.BindEnv("oauth.google.redirect_url", "OAUTH_GOOGLE_REDIRECT_URL")
	_ = v.BindEnv("oauth.github.client_id", "OAUTH_GITHUB_CLIENT_ID")
	_ = v.BindEnv("oauth.github.client_secret", "OAUTH_GITHUB_CLIENT_SECRET")
	_ = v.BindEnv("oauth.github.redirect_url", "OAUTH_GITHUB_REDIRECT_URL")
	_ = v.BindEnv("oauth.microsoft.client_id", "OAUTH_MICROSOFT_CLIENT_ID")
	_ = v.BindEnv("oauth.microsoft.client_secret", "OAUTH_MICROSOFT_CLIENT_SECRET")
	_ = v.BindEnv("oauth.microsoft.redirect_url", "OAUTH_MICROSOFT_REDIRECT_URL")

	// Bind AI environment variables
	_ = v.BindEnv("ai.openai.api_key", "AI_OPENAI_API_KEY")
	_ = v.BindEnv("ai.anthropic.api_key", "AI_ANTHROPIC_API_KEY")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	result := ValidateConfiguration(&cfg)

	return &cfg, result, nil
}

func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "linkflow")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.url", "http://localhost:8090")
	v.SetDefault("app.frontend_url", "http://localhost:3000")
	v.SetDefault("app.cors_origins", []string{"http://localhost:3000"})

	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8090)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "linkflow")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.max_lifetime", time.Hour)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)

	// JWT defaults
	v.SetDefault("jwt.access_expiry", 15*time.Minute)
	v.SetDefault("jwt.refresh_expiry", 168*time.Hour)
	v.SetDefault("jwt.issuer", "linkflow")

	// Features defaults
	v.SetDefault("features.webhook_stream.enabled", true)
	v.SetDefault("features.webhook_stream.max_len", 100000)
	v.SetDefault("features.webhook_stream.batch_size", 10)
	v.SetDefault("features.webhook_stream.max_retries", 3)
	v.SetDefault("features.webhook_stream.consumer_count", 2)

	v.SetDefault("features.rate_limit.enabled", true)
	v.SetDefault("features.rate_limit.requests_per_minute", 100)
	v.SetDefault("features.rate_limit.burst_size", 20)

	v.SetDefault("features.execution.default_timeout", 3600)
	v.SetDefault("features.execution.max_retries", 3)
	v.SetDefault("features.execution.worker_count", 10)
	v.SetDefault("features.execution.queue_size", 1000)

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")

	// Storage defaults
	v.SetDefault("storage.provider", "local")
	v.SetDefault("storage.local_path", "./storage")
	v.SetDefault("storage.s3.use_path_style", false)

	// Worker defaults
	v.SetDefault("worker.concurrency", 10)
	v.SetDefault("worker.queues", []string{"critical", "default", "low"})
	v.SetDefault("worker.strict_priority", true)
	v.SetDefault("worker.shutdown_timeout", 30*time.Second)
	v.SetDefault("worker.health_check_interval", 15*time.Second)
	v.SetDefault("worker.retry_limit", 3)

	// Scheduler defaults
	v.SetDefault("scheduler.enabled", true)
	v.SetDefault("scheduler.poll_interval", 10*time.Second)
	v.SetDefault("scheduler.lock_ttl", 30*time.Second)
	v.SetDefault("scheduler.batch_size", 100)
	v.SetDefault("scheduler.leader_election", true)

	// WebSocket defaults
	v.SetDefault("websocket.enabled", true)
	v.SetDefault("websocket.ping_interval", 30*time.Second)
	v.SetDefault("websocket.pong_timeout", 10*time.Second)
	v.SetDefault("websocket.write_timeout", 10*time.Second)
	v.SetDefault("websocket.read_buffer_size", 1024)
	v.SetDefault("websocket.write_buffer_size", 1024)
	v.SetDefault("websocket.max_message_size", 512*1024)

	// Resilience defaults
	v.SetDefault("resilience.circuit_breaker.enabled", true)
	v.SetDefault("resilience.circuit_breaker.max_requests", 5)
	v.SetDefault("resilience.circuit_breaker.interval", 10*time.Second)
	v.SetDefault("resilience.circuit_breaker.timeout", 60*time.Second)
	v.SetDefault("resilience.circuit_breaker.failure_threshold", 5)

	v.SetDefault("resilience.retry.max_attempts", 3)
	v.SetDefault("resilience.retry.initial_interval", 100*time.Millisecond)
	v.SetDefault("resilience.retry.max_interval", 10*time.Second)
	v.SetDefault("resilience.retry.multiplier", 2.0)

	v.SetDefault("resilience.rate_limit.enabled", true)
	v.SetDefault("resilience.rate_limit.rate", 100.0)
	v.SetDefault("resilience.rate_limit.burst", 20)
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetAddress returns the Redis address
func (c *RedisConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddress returns the server address
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment checks if running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction checks if running in production mode
func (c *AppConfig) IsProduction() bool {
	return c.Environment == "production"
}
