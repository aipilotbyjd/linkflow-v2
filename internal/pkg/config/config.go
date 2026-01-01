package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	instance *Config
	once     sync.Once
	mu       sync.RWMutex
)

// Get returns the singleton config instance (thread-safe)
func Get() *Config {
	once.Do(func() {
		var err error
		instance, err = Load()
		if err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}
	})
	return instance
}

// MustLoad loads config and panics on error (for application startup)
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// Reset clears the singleton instance (for testing only)
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	instance = nil
	once = sync.Once{}
}

// SetInstance sets a custom config instance (for testing only)
func SetInstance(cfg *Config) {
	mu.Lock()
	defer mu.Unlock()
	instance = cfg
	once.Do(func() {})
}

// =============================================================================
// Configuration Structs
// =============================================================================

type Config struct {
	App            AppConfig
	Server         ServerConfig
	Database       DatabaseConfig
	Redis          RedisConfig
	JWT            JWTConfig
	OAuth          OAuthConfig
	S3             S3Config
	Stripe         StripeConfig
	SMTP           SMTPConfig
	Queue          QueueConfig
	Encryption     EncryptionConfig
	Tracing        TracingConfig
	CircuitBreaker CircuitBreakerConfig
	Retry          RetryConfig
	RateLimit      RateLimitConfig
	CORS           CORSConfig
	Logging        LoggingConfig
	Metrics        MetricsConfig
	Health         HealthConfig
	Features       FeaturesConfig
}

type AppConfig struct {
	Name                   string
	Environment            string
	Debug                  bool
	URL                    string
	FrontendURL            string
	ExecutionRetentionDays int
	GracefulShutdownDelay  time.Duration
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int
	TrustedProxies  []string
}

type DatabaseConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	ConnMaxIdleTime   time.Duration
	SlowQueryThreshold time.Duration
	LogQueries        bool
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *DatabaseConfig) DSNWithoutPassword() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Name, c.SSLMode,
	)
}

type RedisConfig struct {
	Host            string
	Port            int
	Password        string
	DB              int
	TLS             bool
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	MaxConnAge      time.Duration
	ClusterMode     bool
	ClusterAddrs    []string
	SentinelMaster  string
	SentinelAddrs   []string
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type JWTConfig struct {
	Secret                 string
	AccessExpiry           time.Duration
	RefreshExpiry          time.Duration
	Issuer                 string
	Audience               []string
	SigningMethod          string
	RefreshTokenRotation   bool
	RefreshTokenReuseLimit int
}

type OAuthConfig struct {
	Google    OAuthProviderConfig
	GitHub    OAuthProviderConfig
	Microsoft OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type S3Config struct {
	Endpoint         string
	Region           string
	Bucket           string
	AccessKeyID      string
	SecretAccessKey  string
	UseSSL           bool
	PathStyle        bool
	PresignedExpiry  time.Duration
	UploadPartSize   int64
	MaxUploadParts   int
}

type StripeConfig struct {
	SecretKey        string
	WebhookSecret    string
	PublishableKey   string
	WebhookTolerance time.Duration
}

type SMTPConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	From            string
	FromName        string
	UseTLS          bool
	InsecureSkipVerify bool
	ConnectionTimeout time.Duration
}

type QueueConfig struct {
	Concurrency     int
	StrictPriority  bool
	ShutdownTimeout time.Duration
	HealthCheckInterval time.Duration
	RetryLimit      int
	RetryDelay      time.Duration
	Retention       time.Duration
	Queues          QueuePriorities
}

type QueuePriorities struct {
	Critical int
	Default  int
	Low      int
}

type EncryptionConfig struct {
	Key        string
	Algorithm  string
	KeyRotation bool
	OldKeys    []string
}

type TracingConfig struct {
	Enabled     bool
	Provider    string
	Endpoint    string
	ServiceName string
	SampleRate  float64
	Insecure    bool
	Headers     map[string]string
}

type CircuitBreakerConfig struct {
	Enabled             bool
	Threshold           int
	Timeout             time.Duration
	MaxRequests         int
	Interval            time.Duration
	OnStateChange       bool
}

type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomizeFactor float64
}

type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond int
	Burst             int
	ByIP              bool
	ByUser            bool
	ExcludePaths      []string
	CustomLimits      map[string]int
}

type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type LoggingConfig struct {
	Level           string
	Format          string
	Output          string
	File            string
	MaxSize         int
	MaxBackups      int
	MaxAge          int
	Compress        bool
	IncludeCaller   bool
	RedactSecrets   bool
}

type MetricsConfig struct {
	Enabled        bool
	Path           string
	Namespace      string
	Subsystem      string
	EnableRuntimeMetrics bool
	EnableDBMetrics      bool
	EnableHTTPMetrics    bool
}

type HealthConfig struct {
	Path             string
	ReadyPath        string
	LivePath         string
	DetailedResponse bool
	IncludeDBCheck   bool
	IncludeRedisCheck bool
	Timeout          time.Duration
}

type FeaturesConfig struct {
	WebhookStream WebhookStreamConfig
	Execution     ExecutionConfig
}

type WebhookStreamConfig struct {
	Enabled       bool
	MaxLen        int64
	DLQMaxLen     int64
	BatchSize     int
	MaxRetries    int
	StaleTimeout  int
	ConsumerCount int
	ProcessingTimeout time.Duration
}

type ExecutionConfig struct {
	MaxConcurrent    int
	DefaultTimeout   time.Duration
	MaxTimeout       time.Duration
	RetentionDays    int
	CleanupInterval  time.Duration
	MaxPayloadSize   int64
}

// =============================================================================
// Loading Functions
// =============================================================================

func Load() (*Config, error) {
	loadEnvFiles()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/linkflow")
	viper.AddConfigPath("$HOME/.linkflow")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	bindEnvVars()
	setDefaults()
	parseConnectionURLs()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		log.Warn().Msg("No config file found, using environment variables and defaults")
	} else {
		log.Info().Str("file", viper.ConfigFileUsed()).Msg("Loaded config file")
	}

	loadEnvOverlay()

	// Ensure env vars override config file values
	applyEnvOverrides()

	cfg := buildConfig()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	logConfigSummary(cfg)

	return cfg, nil
}

func loadEnvFiles() {
	envFiles := []string{".env", ".env.local"}
	
	if env := os.Getenv("APP_ENVIRONMENT"); env != "" {
		envFiles = append(envFiles, fmt.Sprintf(".env.%s", env))
		envFiles = append(envFiles, fmt.Sprintf(".env.%s.local", env))
	}

	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			log.Debug().Str("file", file).Msg("Loaded env file")
		}
	}
}

func loadEnvOverlay() {
	env := viper.GetString("app.environment")
	if env != "" && env != "development" {
		viper.SetConfigName(fmt.Sprintf("config.%s", env))
		if err := viper.MergeInConfig(); err == nil {
			log.Info().Str("environment", env).Str("file", viper.ConfigFileUsed()).Msg("Merged environment-specific config")
		}
	}
}

// applyEnvOverrides ensures environment variables always override config file values
func applyEnvOverrides() {
	overrides := map[string]string{
		"database.sslmode":  "DATABASE_SSLMODE",
		"database.host":     "DATABASE_HOST",
		"database.port":     "DATABASE_PORT",
		"database.user":     "DATABASE_USER",
		"database.password": "DATABASE_PASSWORD",
		"database.name":     "DATABASE_NAME",
		"redis.host":        "REDIS_HOST",
		"redis.port":        "REDIS_PORT",
		"redis.password":    "REDIS_PASSWORD",
		"redis.tls":         "REDIS_TLS",
		"encryption.key":    "ENCRYPTION_KEY",
		"jwt.secret":        "JWT_SECRET",
		"app.debug":         "APP_DEBUG",
	}

	for key, envVar := range overrides {
		if val := os.Getenv(envVar); val != "" {
			viper.Set(key, val)
		}
	}
}

func bindEnvVars() {
	bindings := map[string][]string{
		// App
		"app.name":            {"APP_NAME"},
		"app.environment":     {"APP_ENVIRONMENT", "GO_ENV", "ENVIRONMENT", "ENV"},
		"app.debug":           {"APP_DEBUG", "DEBUG"},
		"app.url":             {"APP_URL"},
		"app.frontend_url":    {"APP_FRONTEND_URL", "FRONTEND_URL"},

		// Server
		"server.port": {"PORT", "SERVER_PORT"},
		"server.host": {"HOST", "SERVER_HOST"},

		// Database
		"database.host":     {"DATABASE_HOST", "DB_HOST", "PGHOST"},
		"database.port":     {"DATABASE_PORT", "DB_PORT", "PGPORT"},
		"database.user":     {"DATABASE_USER", "DB_USER", "PGUSER"},
		"database.password": {"DATABASE_PASSWORD", "DB_PASSWORD", "PGPASSWORD"},
		"database.name":     {"DATABASE_NAME", "DB_NAME", "PGDATABASE"},
		"database.sslmode":  {"DATABASE_SSLMODE", "DB_SSLMODE", "PGSSLMODE"},

		// Redis
		"redis.host":     {"REDIS_HOST"},
		"redis.port":     {"REDIS_PORT"},
		"redis.password": {"REDIS_PASSWORD"},
		"redis.db":       {"REDIS_DB"},
		"redis.tls":      {"REDIS_TLS"},

		// JWT
		"jwt.secret": {"JWT_SECRET"},

		// OAuth
		"oauth.google.client_id":        {"OAUTH_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID"},
		"oauth.google.client_secret":    {"OAUTH_GOOGLE_CLIENT_SECRET", "GOOGLE_CLIENT_SECRET"},
		"oauth.github.client_id":        {"OAUTH_GITHUB_CLIENT_ID", "GITHUB_CLIENT_ID"},
		"oauth.github.client_secret":    {"OAUTH_GITHUB_CLIENT_SECRET", "GITHUB_CLIENT_SECRET"},
		"oauth.microsoft.client_id":     {"OAUTH_MICROSOFT_CLIENT_ID", "MICROSOFT_CLIENT_ID"},
		"oauth.microsoft.client_secret": {"OAUTH_MICROSOFT_CLIENT_SECRET", "MICROSOFT_CLIENT_SECRET"},

		// S3
		"s3.endpoint":          {"S3_ENDPOINT", "AWS_S3_ENDPOINT"},
		"s3.region":            {"S3_REGION", "AWS_REGION", "AWS_DEFAULT_REGION"},
		"s3.bucket":            {"S3_BUCKET", "AWS_S3_BUCKET"},
		"s3.access_key_id":     {"S3_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID"},
		"s3.secret_access_key": {"S3_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY"},

		// Stripe
		"stripe.secret_key":     {"STRIPE_SECRET_KEY"},
		"stripe.webhook_secret": {"STRIPE_WEBHOOK_SECRET"},
		"stripe.publishable_key": {"STRIPE_PUBLISHABLE_KEY"},

		// SMTP
		"smtp.host":     {"SMTP_HOST"},
		"smtp.port":     {"SMTP_PORT"},
		"smtp.username": {"SMTP_USERNAME"},
		"smtp.password": {"SMTP_PASSWORD"},
		"smtp.from":     {"SMTP_FROM"},
		"smtp.from_name": {"SMTP_FROM_NAME"},

		// Encryption
		"encryption.key": {"ENCRYPTION_KEY", "CREDENTIAL_ENCRYPTION_KEY"},

		// Tracing
		"tracing.endpoint": {"TRACING_ENDPOINT", "OTEL_EXPORTER_OTLP_ENDPOINT"},

		// Logging
		"logging.level":  {"LOG_LEVEL"},
		"logging.format": {"LOG_FORMAT"},
	}

	for key, envVars := range bindings {
		_ = viper.BindEnv(append([]string{key}, envVars...)...)
	}
}

func parseConnectionURLs() {
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		if err := parseDatabaseURL(dbURL); err != nil {
			log.Warn().Err(err).Msg("Failed to parse DATABASE_URL")
		} else {
			log.Debug().Msg("Parsed DATABASE_URL")
		}
	}

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		if err := parseRedisURL(redisURL); err != nil {
			log.Warn().Err(err).Msg("Failed to parse REDIS_URL")
		} else {
			log.Debug().Msg("Parsed REDIS_URL")
		}
	}
}

func parseDatabaseURL(dbURL string) error {
	u, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	viper.Set("database.host", u.Hostname())
	if port := u.Port(); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			viper.Set("database.port", p)
		}
	}
	if u.User != nil {
		viper.Set("database.user", u.User.Username())
		if pass, ok := u.User.Password(); ok {
			viper.Set("database.password", pass)
		}
	}
	if len(u.Path) > 1 {
		viper.Set("database.name", u.Path[1:])
	}
	if sslmode := u.Query().Get("sslmode"); sslmode != "" {
		viper.Set("database.sslmode", sslmode)
	}

	return nil
}

func parseRedisURL(redisURL string) error {
	u, err := url.Parse(redisURL)
	if err != nil {
		return fmt.Errorf("invalid REDIS_URL: %w", err)
	}

	viper.Set("redis.host", u.Hostname())
	if port := u.Port(); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			viper.Set("redis.port", p)
		}
	}
	if u.User != nil {
		if pass, ok := u.User.Password(); ok {
			viper.Set("redis.password", pass)
		}
	}
	if len(u.Path) > 1 {
		if db, err := strconv.Atoi(u.Path[1:]); err == nil {
			viper.Set("redis.db", db)
		}
	}
	viper.Set("redis.tls", u.Scheme == "rediss")

	return nil
}

func buildConfig() *Config {
	cfg := &Config{}

	// App
	cfg.App.Name = viper.GetString("app.name")
	cfg.App.Environment = viper.GetString("app.environment")
	cfg.App.Debug = viper.GetBool("app.debug")
	cfg.App.URL = viper.GetString("app.url")
	cfg.App.FrontendURL = viper.GetString("app.frontend_url")
	cfg.App.ExecutionRetentionDays = viper.GetInt("app.execution_retention_days")
	cfg.App.GracefulShutdownDelay = viper.GetDuration("app.graceful_shutdown_delay")

	// Server
	cfg.Server.Host = viper.GetString("server.host")
	cfg.Server.Port = viper.GetInt("server.port")
	cfg.Server.ReadTimeout = viper.GetDuration("server.read_timeout")
	cfg.Server.WriteTimeout = viper.GetDuration("server.write_timeout")
	cfg.Server.IdleTimeout = viper.GetDuration("server.idle_timeout")
	cfg.Server.ShutdownTimeout = viper.GetDuration("server.shutdown_timeout")
	cfg.Server.MaxHeaderBytes = viper.GetInt("server.max_header_bytes")
	cfg.Server.TrustedProxies = viper.GetStringSlice("server.trusted_proxies")

	// Database
	cfg.Database.Host = viper.GetString("database.host")
	cfg.Database.Port = viper.GetInt("database.port")
	cfg.Database.User = viper.GetString("database.user")
	cfg.Database.Password = viper.GetString("database.password")
	cfg.Database.Name = viper.GetString("database.name")
	cfg.Database.SSLMode = viper.GetString("database.sslmode")
	cfg.Database.MaxOpenConns = viper.GetInt("database.max_open_conns")
	cfg.Database.MaxIdleConns = viper.GetInt("database.max_idle_conns")
	cfg.Database.ConnMaxLifetime = viper.GetDuration("database.conn_max_lifetime")
	cfg.Database.ConnMaxIdleTime = viper.GetDuration("database.conn_max_idle_time")
	cfg.Database.SlowQueryThreshold = viper.GetDuration("database.slow_query_threshold")
	cfg.Database.LogQueries = viper.GetBool("database.log_queries")

	// Redis
	cfg.Redis.Host = viper.GetString("redis.host")
	cfg.Redis.Port = viper.GetInt("redis.port")
	cfg.Redis.Password = viper.GetString("redis.password")
	cfg.Redis.DB = viper.GetInt("redis.db")
	cfg.Redis.TLS = viper.GetBool("redis.tls")
	cfg.Redis.PoolSize = viper.GetInt("redis.pool_size")
	cfg.Redis.MinIdleConns = viper.GetInt("redis.min_idle_conns")
	cfg.Redis.MaxRetries = viper.GetInt("redis.max_retries")
	cfg.Redis.DialTimeout = viper.GetDuration("redis.dial_timeout")
	cfg.Redis.ReadTimeout = viper.GetDuration("redis.read_timeout")
	cfg.Redis.WriteTimeout = viper.GetDuration("redis.write_timeout")
	cfg.Redis.PoolTimeout = viper.GetDuration("redis.pool_timeout")
	cfg.Redis.MaxConnAge = viper.GetDuration("redis.max_conn_age")
	cfg.Redis.ClusterMode = viper.GetBool("redis.cluster_mode")
	cfg.Redis.ClusterAddrs = viper.GetStringSlice("redis.cluster_addrs")
	cfg.Redis.SentinelMaster = viper.GetString("redis.sentinel_master")
	cfg.Redis.SentinelAddrs = viper.GetStringSlice("redis.sentinel_addrs")

	// JWT
	cfg.JWT.Secret = viper.GetString("jwt.secret")
	cfg.JWT.AccessExpiry = viper.GetDuration("jwt.access_expiry")
	cfg.JWT.RefreshExpiry = viper.GetDuration("jwt.refresh_expiry")
	cfg.JWT.Issuer = viper.GetString("jwt.issuer")
	cfg.JWT.Audience = viper.GetStringSlice("jwt.audience")
	cfg.JWT.SigningMethod = viper.GetString("jwt.signing_method")
	cfg.JWT.RefreshTokenRotation = viper.GetBool("jwt.refresh_token_rotation")
	cfg.JWT.RefreshTokenReuseLimit = viper.GetInt("jwt.refresh_token_reuse_limit")

	// OAuth
	cfg.OAuth.Google.ClientID = viper.GetString("oauth.google.client_id")
	cfg.OAuth.Google.ClientSecret = viper.GetString("oauth.google.client_secret")
	cfg.OAuth.Google.RedirectURL = viper.GetString("oauth.google.redirect_url")
	cfg.OAuth.Google.Scopes = viper.GetStringSlice("oauth.google.scopes")
	cfg.OAuth.GitHub.ClientID = viper.GetString("oauth.github.client_id")
	cfg.OAuth.GitHub.ClientSecret = viper.GetString("oauth.github.client_secret")
	cfg.OAuth.GitHub.RedirectURL = viper.GetString("oauth.github.redirect_url")
	cfg.OAuth.GitHub.Scopes = viper.GetStringSlice("oauth.github.scopes")
	cfg.OAuth.Microsoft.ClientID = viper.GetString("oauth.microsoft.client_id")
	cfg.OAuth.Microsoft.ClientSecret = viper.GetString("oauth.microsoft.client_secret")
	cfg.OAuth.Microsoft.RedirectURL = viper.GetString("oauth.microsoft.redirect_url")
	cfg.OAuth.Microsoft.Scopes = viper.GetStringSlice("oauth.microsoft.scopes")

	// S3
	cfg.S3.Endpoint = viper.GetString("s3.endpoint")
	cfg.S3.Region = viper.GetString("s3.region")
	cfg.S3.Bucket = viper.GetString("s3.bucket")
	cfg.S3.AccessKeyID = viper.GetString("s3.access_key_id")
	cfg.S3.SecretAccessKey = viper.GetString("s3.secret_access_key")
	cfg.S3.UseSSL = viper.GetBool("s3.use_ssl")
	cfg.S3.PathStyle = viper.GetBool("s3.path_style")
	cfg.S3.PresignedExpiry = viper.GetDuration("s3.presigned_expiry")
	cfg.S3.UploadPartSize = viper.GetInt64("s3.upload_part_size")
	cfg.S3.MaxUploadParts = viper.GetInt("s3.max_upload_parts")

	// Stripe
	cfg.Stripe.SecretKey = viper.GetString("stripe.secret_key")
	cfg.Stripe.WebhookSecret = viper.GetString("stripe.webhook_secret")
	cfg.Stripe.PublishableKey = viper.GetString("stripe.publishable_key")
	cfg.Stripe.WebhookTolerance = viper.GetDuration("stripe.webhook_tolerance")

	// SMTP
	cfg.SMTP.Host = viper.GetString("smtp.host")
	cfg.SMTP.Port = viper.GetInt("smtp.port")
	cfg.SMTP.Username = viper.GetString("smtp.username")
	cfg.SMTP.Password = viper.GetString("smtp.password")
	cfg.SMTP.From = viper.GetString("smtp.from")
	cfg.SMTP.FromName = viper.GetString("smtp.from_name")
	cfg.SMTP.UseTLS = viper.GetBool("smtp.use_tls")
	cfg.SMTP.InsecureSkipVerify = viper.GetBool("smtp.insecure_skip_verify")
	cfg.SMTP.ConnectionTimeout = viper.GetDuration("smtp.connection_timeout")

	// Queue
	cfg.Queue.Concurrency = viper.GetInt("queue.concurrency")
	cfg.Queue.StrictPriority = viper.GetBool("queue.strict_priority")
	cfg.Queue.ShutdownTimeout = viper.GetDuration("queue.shutdown_timeout")
	cfg.Queue.HealthCheckInterval = viper.GetDuration("queue.health_check_interval")
	cfg.Queue.RetryLimit = viper.GetInt("queue.retry_limit")
	cfg.Queue.RetryDelay = viper.GetDuration("queue.retry_delay")
	cfg.Queue.Retention = viper.GetDuration("queue.retention")
	cfg.Queue.Queues.Critical = viper.GetInt("queue.queues.critical")
	cfg.Queue.Queues.Default = viper.GetInt("queue.queues.default")
	cfg.Queue.Queues.Low = viper.GetInt("queue.queues.low")

	// Encryption
	cfg.Encryption.Key = viper.GetString("encryption.key")
	cfg.Encryption.Algorithm = viper.GetString("encryption.algorithm")
	cfg.Encryption.KeyRotation = viper.GetBool("encryption.key_rotation")
	cfg.Encryption.OldKeys = viper.GetStringSlice("encryption.old_keys")

	// Tracing
	cfg.Tracing.Enabled = viper.GetBool("tracing.enabled")
	cfg.Tracing.Provider = viper.GetString("tracing.provider")
	cfg.Tracing.Endpoint = viper.GetString("tracing.endpoint")
	cfg.Tracing.ServiceName = viper.GetString("tracing.service_name")
	cfg.Tracing.SampleRate = viper.GetFloat64("tracing.sample_rate")
	cfg.Tracing.Insecure = viper.GetBool("tracing.insecure")
	cfg.Tracing.Headers = viper.GetStringMapString("tracing.headers")

	// Circuit Breaker
	cfg.CircuitBreaker.Enabled = viper.GetBool("circuit_breaker.enabled")
	cfg.CircuitBreaker.Threshold = viper.GetInt("circuit_breaker.threshold")
	cfg.CircuitBreaker.Timeout = viper.GetDuration("circuit_breaker.timeout")
	cfg.CircuitBreaker.MaxRequests = viper.GetInt("circuit_breaker.max_requests")
	cfg.CircuitBreaker.Interval = viper.GetDuration("circuit_breaker.interval")
	cfg.CircuitBreaker.OnStateChange = viper.GetBool("circuit_breaker.on_state_change")

	// Retry
	cfg.Retry.MaxAttempts = viper.GetInt("retry.max_attempts")
	cfg.Retry.InitialInterval = viper.GetDuration("retry.initial_interval")
	cfg.Retry.MaxInterval = viper.GetDuration("retry.max_interval")
	cfg.Retry.Multiplier = viper.GetFloat64("retry.multiplier")
	cfg.Retry.RandomizeFactor = viper.GetFloat64("retry.randomize_factor")

	// Rate Limit
	cfg.RateLimit.Enabled = viper.GetBool("rate_limit.enabled")
	cfg.RateLimit.RequestsPerSecond = viper.GetInt("rate_limit.requests_per_second")
	cfg.RateLimit.Burst = viper.GetInt("rate_limit.burst")
	cfg.RateLimit.ByIP = viper.GetBool("rate_limit.by_ip")
	cfg.RateLimit.ByUser = viper.GetBool("rate_limit.by_user")
	cfg.RateLimit.ExcludePaths = viper.GetStringSlice("rate_limit.exclude_paths")
	cfg.RateLimit.CustomLimits = getStringMapInt("rate_limit.custom_limits")

	// CORS
	cfg.CORS.Enabled = viper.GetBool("cors.enabled")
	cfg.CORS.AllowedOrigins = viper.GetStringSlice("cors.allowed_origins")
	cfg.CORS.AllowedMethods = viper.GetStringSlice("cors.allowed_methods")
	cfg.CORS.AllowedHeaders = viper.GetStringSlice("cors.allowed_headers")
	cfg.CORS.ExposedHeaders = viper.GetStringSlice("cors.exposed_headers")
	cfg.CORS.AllowCredentials = viper.GetBool("cors.allow_credentials")
	cfg.CORS.MaxAge = viper.GetInt("cors.max_age")

	// Logging
	cfg.Logging.Level = viper.GetString("logging.level")
	cfg.Logging.Format = viper.GetString("logging.format")
	cfg.Logging.Output = viper.GetString("logging.output")
	cfg.Logging.File = viper.GetString("logging.file")
	cfg.Logging.MaxSize = viper.GetInt("logging.max_size")
	cfg.Logging.MaxBackups = viper.GetInt("logging.max_backups")
	cfg.Logging.MaxAge = viper.GetInt("logging.max_age")
	cfg.Logging.Compress = viper.GetBool("logging.compress")
	cfg.Logging.IncludeCaller = viper.GetBool("logging.include_caller")
	cfg.Logging.RedactSecrets = viper.GetBool("logging.redact_secrets")

	// Metrics
	cfg.Metrics.Enabled = viper.GetBool("metrics.enabled")
	cfg.Metrics.Path = viper.GetString("metrics.path")
	cfg.Metrics.Namespace = viper.GetString("metrics.namespace")
	cfg.Metrics.Subsystem = viper.GetString("metrics.subsystem")
	cfg.Metrics.EnableRuntimeMetrics = viper.GetBool("metrics.enable_runtime_metrics")
	cfg.Metrics.EnableDBMetrics = viper.GetBool("metrics.enable_db_metrics")
	cfg.Metrics.EnableHTTPMetrics = viper.GetBool("metrics.enable_http_metrics")

	// Health
	cfg.Health.Path = viper.GetString("health.path")
	cfg.Health.ReadyPath = viper.GetString("health.ready_path")
	cfg.Health.LivePath = viper.GetString("health.live_path")
	cfg.Health.DetailedResponse = viper.GetBool("health.detailed_response")
	cfg.Health.IncludeDBCheck = viper.GetBool("health.include_db_check")
	cfg.Health.IncludeRedisCheck = viper.GetBool("health.include_redis_check")
	cfg.Health.Timeout = viper.GetDuration("health.timeout")

	// Features - Webhook Stream
	cfg.Features.WebhookStream.Enabled = viper.GetBool("features.webhook_stream.enabled")
	cfg.Features.WebhookStream.MaxLen = viper.GetInt64("features.webhook_stream.max_len")
	cfg.Features.WebhookStream.DLQMaxLen = viper.GetInt64("features.webhook_stream.dlq_max_len")
	cfg.Features.WebhookStream.BatchSize = viper.GetInt("features.webhook_stream.batch_size")
	cfg.Features.WebhookStream.MaxRetries = viper.GetInt("features.webhook_stream.max_retries")
	cfg.Features.WebhookStream.StaleTimeout = viper.GetInt("features.webhook_stream.stale_timeout")
	cfg.Features.WebhookStream.ConsumerCount = viper.GetInt("features.webhook_stream.consumer_count")
	cfg.Features.WebhookStream.ProcessingTimeout = viper.GetDuration("features.webhook_stream.processing_timeout")

	// Features - Execution
	cfg.Features.Execution.MaxConcurrent = viper.GetInt("features.execution.max_concurrent")
	cfg.Features.Execution.DefaultTimeout = viper.GetDuration("features.execution.default_timeout")
	cfg.Features.Execution.MaxTimeout = viper.GetDuration("features.execution.max_timeout")
	cfg.Features.Execution.RetentionDays = viper.GetInt("features.execution.retention_days")
	cfg.Features.Execution.CleanupInterval = viper.GetDuration("features.execution.cleanup_interval")
	cfg.Features.Execution.MaxPayloadSize = viper.GetInt64("features.execution.max_payload_size")

	return cfg
}

func setDefaults() {
	// App
	viper.SetDefault("app.name", "linkflow")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", false)
	viper.SetDefault("app.url", "http://localhost:8080")
	viper.SetDefault("app.frontend_url", "http://localhost:3000")
	viper.SetDefault("app.execution_retention_days", 30)
	viper.SetDefault("app.graceful_shutdown_delay", "5s")

	// Server
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.shutdown_timeout", "30s")
	viper.SetDefault("server.max_header_bytes", 1<<20) // 1MB

	// Database
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "linkflow")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", "30m")
	viper.SetDefault("database.conn_max_idle_time", "5m")
	viper.SetDefault("database.slow_query_threshold", "200ms")
	viper.SetDefault("database.log_queries", false)

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.tls", false)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")
	viper.SetDefault("redis.pool_timeout", "4s")
	viper.SetDefault("redis.max_conn_age", "0s")

	// JWT
	viper.SetDefault("jwt.access_expiry", "15m")
	viper.SetDefault("jwt.refresh_expiry", "168h")
	viper.SetDefault("jwt.issuer", "linkflow")
	viper.SetDefault("jwt.signing_method", "HS256")
	viper.SetDefault("jwt.refresh_token_rotation", true)
	viper.SetDefault("jwt.refresh_token_reuse_limit", 0)

	// S3
	viper.SetDefault("s3.region", "us-east-1")
	viper.SetDefault("s3.use_ssl", true)
	viper.SetDefault("s3.path_style", false)
	viper.SetDefault("s3.presigned_expiry", "15m")
	viper.SetDefault("s3.upload_part_size", 5*1024*1024) // 5MB
	viper.SetDefault("s3.max_upload_parts", 10000)

	// Stripe
	viper.SetDefault("stripe.webhook_tolerance", "300s")

	// SMTP
	viper.SetDefault("smtp.port", 587)
	viper.SetDefault("smtp.from_name", "LinkFlow")
	viper.SetDefault("smtp.use_tls", true)
	viper.SetDefault("smtp.insecure_skip_verify", false)
	viper.SetDefault("smtp.connection_timeout", "30s")

	// Queue
	viper.SetDefault("queue.concurrency", 10)
	viper.SetDefault("queue.strict_priority", false)
	viper.SetDefault("queue.shutdown_timeout", "30s")
	viper.SetDefault("queue.health_check_interval", "15s")
	viper.SetDefault("queue.retry_limit", 3)
	viper.SetDefault("queue.retry_delay", "10s")
	viper.SetDefault("queue.retention", "24h")
	viper.SetDefault("queue.queues.critical", 6)
	viper.SetDefault("queue.queues.default", 3)
	viper.SetDefault("queue.queues.low", 1)

	// Encryption
	viper.SetDefault("encryption.algorithm", "aes-256-gcm")
	viper.SetDefault("encryption.key_rotation", false)

	// Tracing
	viper.SetDefault("tracing.enabled", false)
	viper.SetDefault("tracing.provider", "otlp")
	viper.SetDefault("tracing.service_name", "linkflow")
	viper.SetDefault("tracing.sample_rate", 0.1)
	viper.SetDefault("tracing.insecure", false)

	// Circuit Breaker
	viper.SetDefault("circuit_breaker.enabled", true)
	viper.SetDefault("circuit_breaker.threshold", 5)
	viper.SetDefault("circuit_breaker.timeout", "30s")
	viper.SetDefault("circuit_breaker.max_requests", 100)
	viper.SetDefault("circuit_breaker.interval", "60s")
	viper.SetDefault("circuit_breaker.on_state_change", true)

	// Retry
	viper.SetDefault("retry.max_attempts", 3)
	viper.SetDefault("retry.initial_interval", "100ms")
	viper.SetDefault("retry.max_interval", "10s")
	viper.SetDefault("retry.multiplier", 2.0)
	viper.SetDefault("retry.randomize_factor", 0.5)

	// Rate Limit
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_second", 100)
	viper.SetDefault("rate_limit.burst", 200)
	viper.SetDefault("rate_limit.by_ip", true)
	viper.SetDefault("rate_limit.by_user", true)
	viper.SetDefault("rate_limit.exclude_paths", []string{"/health", "/ready", "/live", "/metrics"})

	// CORS
	viper.SetDefault("cors.enabled", true)
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Authorization", "Content-Type", "X-Request-ID", "X-Workspace-ID"})
	viper.SetDefault("cors.exposed_headers", []string{"X-Request-ID"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400)

	// Logging
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
	viper.SetDefault("logging.compress", true)
	viper.SetDefault("logging.include_caller", false)
	viper.SetDefault("logging.redact_secrets", true)

	// Metrics
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.namespace", "linkflow")
	viper.SetDefault("metrics.subsystem", "api")
	viper.SetDefault("metrics.enable_runtime_metrics", true)
	viper.SetDefault("metrics.enable_db_metrics", true)
	viper.SetDefault("metrics.enable_http_metrics", true)

	// Health
	viper.SetDefault("health.path", "/health")
	viper.SetDefault("health.ready_path", "/ready")
	viper.SetDefault("health.live_path", "/live")
	viper.SetDefault("health.detailed_response", false)
	viper.SetDefault("health.include_db_check", true)
	viper.SetDefault("health.include_redis_check", true)
	viper.SetDefault("health.timeout", "5s")

	// Features - Webhook Stream
	viper.SetDefault("features.webhook_stream.enabled", true)
	viper.SetDefault("features.webhook_stream.max_len", 100000)
	viper.SetDefault("features.webhook_stream.dlq_max_len", 10000)
	viper.SetDefault("features.webhook_stream.batch_size", 10)
	viper.SetDefault("features.webhook_stream.max_retries", 3)
	viper.SetDefault("features.webhook_stream.stale_timeout", 300)
	viper.SetDefault("features.webhook_stream.consumer_count", 2)
	viper.SetDefault("features.webhook_stream.processing_timeout", "30s")

	// Features - Execution
	viper.SetDefault("features.execution.max_concurrent", 100)
	viper.SetDefault("features.execution.default_timeout", "300s")
	viper.SetDefault("features.execution.max_timeout", "3600s")
	viper.SetDefault("features.execution.retention_days", 30)
	viper.SetDefault("features.execution.cleanup_interval", "1h")
	viper.SetDefault("features.execution.max_payload_size", 10*1024*1024) // 10MB
}

func logConfigSummary(cfg *Config) {
	log.Info().
		Str("app", cfg.App.Name).
		Str("environment", cfg.App.Environment).
		Bool("debug", cfg.App.Debug).
		Int("server_port", cfg.Server.Port).
		Str("database", cfg.Database.DSNWithoutPassword()).
		Str("redis", cfg.Redis.Addr()).
		Bool("redis_tls", cfg.Redis.TLS).
		Bool("tracing_enabled", cfg.Tracing.Enabled).
		Bool("metrics_enabled", cfg.Metrics.Enabled).
		Int("queue_concurrency", cfg.Queue.Concurrency).
		Msg("Configuration loaded")
}

// =============================================================================
// Helper Methods
// =============================================================================

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development" || c.App.Environment == ""
}

func (c *Config) IsStaging() bool {
	return c.App.Environment == "staging"
}

func (c *Config) IsTest() bool {
	return c.App.Environment == "test"
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) HasEncryption() bool {
	return c.Encryption.Key != ""
}

func (c *Config) HasSMTP() bool {
	return c.SMTP.Host != "" && c.SMTP.Port > 0
}

func (c *Config) HasS3() bool {
	return c.S3.Bucket != "" && (c.S3.AccessKeyID != "" || c.S3.Endpoint != "")
}

func (c *Config) HasStripe() bool {
	return c.Stripe.SecretKey != ""
}

func (c *Config) HasOAuth(provider string) bool {
	switch strings.ToLower(provider) {
	case "google":
		return c.OAuth.Google.ClientID != "" && c.OAuth.Google.ClientSecret != ""
	case "github":
		return c.OAuth.GitHub.ClientID != "" && c.OAuth.GitHub.ClientSecret != ""
	case "microsoft":
		return c.OAuth.Microsoft.ClientID != "" && c.OAuth.Microsoft.ClientSecret != ""
	default:
		return false
	}
}

func (c *Config) EnabledOAuthProviders() []string {
	var providers []string
	if c.HasOAuth("google") {
		providers = append(providers, "google")
	}
	if c.HasOAuth("github") {
		providers = append(providers, "github")
	}
	if c.HasOAuth("microsoft") {
		providers = append(providers, "microsoft")
	}
	return providers
}

// Redact returns config summary with secrets masked
func (c *Config) Redact() map[string]interface{} {
	return map[string]interface{}{
		"app": map[string]interface{}{
			"name":        c.App.Name,
			"environment": c.App.Environment,
			"debug":       c.App.Debug,
			"url":         c.App.URL,
		},
		"server": map[string]interface{}{
			"host": c.Server.Host,
			"port": c.Server.Port,
		},
		"database": map[string]interface{}{
			"host":     c.Database.Host,
			"port":     c.Database.Port,
			"name":     c.Database.Name,
			"user":     c.Database.User,
			"password": maskSecret(c.Database.Password),
			"sslmode":  c.Database.SSLMode,
		},
		"redis": map[string]interface{}{
			"host":     c.Redis.Host,
			"port":     c.Redis.Port,
			"password": maskSecret(c.Redis.Password),
			"tls":      c.Redis.TLS,
		},
		"jwt": map[string]interface{}{
			"secret":        maskSecret(c.JWT.Secret),
			"access_expiry": c.JWT.AccessExpiry.String(),
			"issuer":        c.JWT.Issuer,
		},
		"queue": map[string]interface{}{
			"concurrency": c.Queue.Concurrency,
		},
		"features": map[string]interface{}{
			"webhook_stream_enabled": c.Features.WebhookStream.Enabled,
			"execution_max_concurrent": c.Features.Execution.MaxConcurrent,
		},
	}
}

func maskSecret(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}

func getStringMapInt(key string) map[string]int {
	result := make(map[string]int)
	raw := viper.GetStringMap(key)
	for k, v := range raw {
		switch val := v.(type) {
		case int:
			result[k] = val
		case int64:
			result[k] = int(val)
		case float64:
			result[k] = int(val)
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				result[k] = i
			}
		}
	}
	return result
}
