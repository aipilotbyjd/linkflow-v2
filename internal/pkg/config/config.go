package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	S3       S3Config
	Stripe   StripeConfig
	SMTP     SMTPConfig
	Features FeaturesConfig
}

type FeaturesConfig struct {
	WebhookStream WebhookStreamConfig
}

type WebhookStreamConfig struct {
	Enabled       bool
	MaxLen        int64 // Max messages in stream (default: 100000)
	DLQMaxLen     int64 // Max messages in dead letter queue (default: 10000)
	BatchSize     int   // Consumer batch size (default: 10)
	MaxRetries    int   // Max retries before DLQ (default: 3)
	StaleTimeout  int   // Seconds before message is considered stale (default: 300)
	ConsumerCount int   // Number of consumer goroutines (default: 2)
}

type AppConfig struct {
	Name                   string
	Environment            string
	Debug                  bool
	URL                    string
	FrontendURL            string
	ExecutionRetentionDays int
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	TLS      bool
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type JWTConfig struct {
	Secret           string
	AccessExpiry     time.Duration
	RefreshExpiry    time.Duration
	Issuer           string
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
}

type S3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

type StripeConfig struct {
	SecretKey      string
	WebhookSecret  string
	PublishableKey string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Enable environment variable override
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables for Docker compatibility
	_ = viper.BindEnv("database.host", "DATABASE_HOST")
	_ = viper.BindEnv("database.port", "DATABASE_PORT")
	_ = viper.BindEnv("database.user", "DATABASE_USER")
	_ = viper.BindEnv("database.password", "DATABASE_PASSWORD")
	_ = viper.BindEnv("database.name", "DATABASE_NAME")
	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config

	// App
	cfg.App.Name = viper.GetString("app.name")
	cfg.App.Environment = viper.GetString("app.environment")
	cfg.App.Debug = viper.GetBool("app.debug")
	cfg.App.URL = viper.GetString("app.url")
	cfg.App.FrontendURL = viper.GetString("app.frontend_url")

	// Server
	cfg.Server.Host = viper.GetString("server.host")
	cfg.Server.Port = viper.GetInt("server.port")
	cfg.Server.ReadTimeout = viper.GetDuration("server.read_timeout")
	cfg.Server.WriteTimeout = viper.GetDuration("server.write_timeout")
	cfg.Server.IdleTimeout = viper.GetDuration("server.idle_timeout")

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

	// Redis
	cfg.Redis.Host = viper.GetString("redis.host")
	cfg.Redis.Port = viper.GetInt("redis.port")
	cfg.Redis.Password = viper.GetString("redis.password")
	cfg.Redis.DB = viper.GetInt("redis.db")
	cfg.Redis.TLS = viper.GetBool("redis.tls")

	// JWT
	cfg.JWT.Secret = viper.GetString("jwt.secret")
	cfg.JWT.AccessExpiry = viper.GetDuration("jwt.access_expiry")
	cfg.JWT.RefreshExpiry = viper.GetDuration("jwt.refresh_expiry")
	cfg.JWT.Issuer = viper.GetString("jwt.issuer")

	// OAuth
	cfg.OAuth.Google.ClientID = viper.GetString("oauth.google.client_id")
	cfg.OAuth.Google.ClientSecret = viper.GetString("oauth.google.client_secret")
	cfg.OAuth.Google.RedirectURL = viper.GetString("oauth.google.redirect_url")
	cfg.OAuth.GitHub.ClientID = viper.GetString("oauth.github.client_id")
	cfg.OAuth.GitHub.ClientSecret = viper.GetString("oauth.github.client_secret")
	cfg.OAuth.GitHub.RedirectURL = viper.GetString("oauth.github.redirect_url")
	cfg.OAuth.Microsoft.ClientID = viper.GetString("oauth.microsoft.client_id")
	cfg.OAuth.Microsoft.ClientSecret = viper.GetString("oauth.microsoft.client_secret")
	cfg.OAuth.Microsoft.RedirectURL = viper.GetString("oauth.microsoft.redirect_url")

	// S3
	cfg.S3.Endpoint = viper.GetString("s3.endpoint")
	cfg.S3.Region = viper.GetString("s3.region")
	cfg.S3.Bucket = viper.GetString("s3.bucket")
	cfg.S3.AccessKeyID = viper.GetString("s3.access_key_id")
	cfg.S3.SecretAccessKey = viper.GetString("s3.secret_access_key")
	cfg.S3.UseSSL = viper.GetBool("s3.use_ssl")

	// Stripe
	cfg.Stripe.SecretKey = viper.GetString("stripe.secret_key")
	cfg.Stripe.WebhookSecret = viper.GetString("stripe.webhook_secret")
	cfg.Stripe.PublishableKey = viper.GetString("stripe.publishable_key")

	// SMTP
	cfg.SMTP.Host = viper.GetString("smtp.host")
	cfg.SMTP.Port = viper.GetInt("smtp.port")
	cfg.SMTP.Username = viper.GetString("smtp.username")
	cfg.SMTP.Password = viper.GetString("smtp.password")
	cfg.SMTP.From = viper.GetString("smtp.from")
	cfg.SMTP.FromName = viper.GetString("smtp.from_name")

	// Features - Webhook Stream
	cfg.Features.WebhookStream.Enabled = viper.GetBool("features.webhook_stream.enabled")
	cfg.Features.WebhookStream.MaxLen = viper.GetInt64("features.webhook_stream.max_len")
	cfg.Features.WebhookStream.DLQMaxLen = viper.GetInt64("features.webhook_stream.dlq_max_len")
	cfg.Features.WebhookStream.BatchSize = viper.GetInt("features.webhook_stream.batch_size")
	cfg.Features.WebhookStream.MaxRetries = viper.GetInt("features.webhook_stream.max_retries")
	cfg.Features.WebhookStream.StaleTimeout = viper.GetInt("features.webhook_stream.stale_timeout")
	cfg.Features.WebhookStream.ConsumerCount = viper.GetInt("features.webhook_stream.consumer_count")

	return &cfg, nil
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "linkflow")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", true)
	viper.SetDefault("app.url", "http://localhost:8080")
	viper.SetDefault("app.frontend_url", "http://localhost:3000")

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "linkflow")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.tls", false)

	// JWT defaults
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.access_expiry", "15m")
	viper.SetDefault("jwt.refresh_expiry", "7d")
	viper.SetDefault("jwt.issuer", "linkflow")

	// S3 defaults
	viper.SetDefault("s3.region", "us-east-1")
	viper.SetDefault("s3.use_ssl", true)

	// SMTP defaults
	viper.SetDefault("smtp.port", 587)
	viper.SetDefault("smtp.from_name", "LinkFlow")

	// Features - Webhook Stream defaults
	viper.SetDefault("features.webhook_stream.enabled", true)
	viper.SetDefault("features.webhook_stream.max_len", 100000)
	viper.SetDefault("features.webhook_stream.dlq_max_len", 10000)
	viper.SetDefault("features.webhook_stream.batch_size", 10)
	viper.SetDefault("features.webhook_stream.max_retries", 3)
	viper.SetDefault("features.webhook_stream.stale_timeout", 300)
	viper.SetDefault("features.webhook_stream.consumer_count", 2)
}
