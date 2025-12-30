package config

import (
	"os"
	"testing"
	"time"
)

// TestConfig returns a minimal valid config for testing
func TestConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:                   "linkflow-test",
			Environment:            "test",
			Debug:                  true,
			URL:                    "http://localhost:8080",
			FrontendURL:            "http://localhost:3000",
			ExecutionRetentionDays: 30,
			GracefulShutdownDelay:  5 * time.Second,
		},
		Server: ServerConfig{
			Host:            "localhost",
			Port:            8080,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			MaxHeaderBytes:  1 << 20,
		},
		Database: DatabaseConfig{
			Host:              "localhost",
			Port:              5432,
			User:              "postgres",
			Password:          "postgres",
			Name:              "linkflow_test",
			SSLMode:           "disable",
			MaxOpenConns:      10,
			MaxIdleConns:      5,
			ConnMaxLifetime:   30 * time.Minute,
			ConnMaxIdleTime:   5 * time.Minute,
			SlowQueryThreshold: 200 * time.Millisecond,
		},
		Redis: RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			DB:           1, // Use different DB for tests
			TLS:          false,
			PoolSize:     10,
			MinIdleConns: 5,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolTimeout:  4 * time.Second,
		},
		JWT: JWTConfig{
			Secret:                 "test-secret-key-that-is-at-least-32-characters",
			AccessExpiry:           15 * time.Minute,
			RefreshExpiry:          168 * time.Hour,
			Issuer:                 "linkflow-test",
			SigningMethod:          "HS256",
			RefreshTokenRotation:   true,
		},
		Queue: QueueConfig{
			Concurrency:     5,
			ShutdownTimeout: 10 * time.Second,
			RetryLimit:      3,
			RetryDelay:      1 * time.Second,
			Retention:       1 * time.Hour,
			Queues: QueuePriorities{
				Critical: 6,
				Default:  3,
				Low:      1,
			},
		},
		RateLimit: RateLimitConfig{
			Enabled:           false, // Disable for tests
			RequestsPerSecond: 1000,
			Burst:             2000,
		},
		CORS: CORSConfig{
			Enabled:          true,
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: false,
			MaxAge:           86400,
		},
		Logging: LoggingConfig{
			Level:         "debug",
			Format:        "console",
			Output:        "stdout",
			IncludeCaller: true,
			RedactSecrets: false,
		},
		Metrics: MetricsConfig{
			Enabled:   false,
			Path:      "/metrics",
			Namespace: "linkflow_test",
		},
		Health: HealthConfig{
			Path:              "/health",
			ReadyPath:         "/ready",
			LivePath:          "/live",
			DetailedResponse:  true,
			IncludeDBCheck:    false,
			IncludeRedisCheck: false,
			Timeout:           5 * time.Second,
		},
		Features: FeaturesConfig{
			WebhookStream: WebhookStreamConfig{
				Enabled:           false,
				MaxLen:            1000,
				DLQMaxLen:         100,
				BatchSize:         10,
				MaxRetries:        3,
				StaleTimeout:      60,
				ConsumerCount:     1,
				ProcessingTimeout: 10 * time.Second,
			},
			Execution: ExecutionConfig{
				MaxConcurrent:   10,
				DefaultTimeout:  30 * time.Second,
				MaxTimeout:      60 * time.Second,
				RetentionDays:   7,
				CleanupInterval: 1 * time.Hour,
				MaxPayloadSize:  1024 * 1024,
			},
		},
	}
}

// SetupTestEnv sets up environment variables for testing and returns a cleanup function
func SetupTestEnv(t *testing.T) func() {
	t.Helper()

	origEnv := map[string]string{}
	testEnv := map[string]string{
		"APP_ENVIRONMENT":    "test",
		"APP_DEBUG":          "true",
		"DATABASE_HOST":      "localhost",
		"DATABASE_PORT":      "5432",
		"DATABASE_USER":      "postgres",
		"DATABASE_PASSWORD":  "postgres",
		"DATABASE_NAME":      "linkflow_test",
		"DATABASE_SSLMODE":   "disable",
		"REDIS_HOST":         "localhost",
		"REDIS_PORT":         "6379",
		"REDIS_DB":           "1",
		"JWT_SECRET":         "test-secret-key-that-is-at-least-32-characters",
	}

	// Save original values
	for k := range testEnv {
		if v, exists := os.LookupEnv(k); exists {
			origEnv[k] = v
		}
	}

	// Set test values
	for k, v := range testEnv {
		os.Setenv(k, v)
	}

	// Return cleanup function
	return func() {
		for k := range testEnv {
			if v, exists := origEnv[k]; exists {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
		// Reset singleton
		Reset()
	}
}

// WithTestConfig sets a test config as the singleton and returns a cleanup function
func WithTestConfig(t *testing.T, cfg *Config) func() {
	t.Helper()
	SetInstance(cfg)
	return func() {
		Reset()
	}
}

// TestConfigWithOverrides returns a test config with custom overrides
func TestConfigWithOverrides(overrides func(*Config)) *Config {
	cfg := TestConfig()
	if overrides != nil {
		overrides(cfg)
	}
	return cfg
}

// TestDatabaseConfig returns a test database config
func TestDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:              getEnvOrDefault("TEST_DATABASE_HOST", "localhost"),
		Port:              5432,
		User:              getEnvOrDefault("TEST_DATABASE_USER", "postgres"),
		Password:          getEnvOrDefault("TEST_DATABASE_PASSWORD", "postgres"),
		Name:              getEnvOrDefault("TEST_DATABASE_NAME", "linkflow_test"),
		SSLMode:           "disable",
		MaxOpenConns:      5,
		MaxIdleConns:      2,
		ConnMaxLifetime:   5 * time.Minute,
		ConnMaxIdleTime:   1 * time.Minute,
		SlowQueryThreshold: 100 * time.Millisecond,
	}
}

// TestRedisConfig returns a test Redis config
func TestRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         getEnvOrDefault("TEST_REDIS_HOST", "localhost"),
		Port:         6379,
		Password:     getEnvOrDefault("TEST_REDIS_PASSWORD", ""),
		DB:           1,
		TLS:          false,
		PoolSize:     5,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// SkipIfNoDatabase skips the test if database is not available
func SkipIfNoDatabase(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_DATABASE_HOST") == "" && os.Getenv("CI") == "" {
		t.Skip("Skipping test: database not available (set TEST_DATABASE_HOST)")
	}
}

// SkipIfNoRedis skips the test if Redis is not available
func SkipIfNoRedis(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_REDIS_HOST") == "" && os.Getenv("CI") == "" {
		t.Skip("Skipping test: Redis not available (set TEST_REDIS_HOST)")
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
