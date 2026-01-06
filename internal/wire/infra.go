package wire

import (
	"fmt"

	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"gorm.io/gorm"
)

// ProvideConfig loads the application configuration
func ProvideConfig() (*config.Config, error) {
	return config.Load()
}

// ProvideDB creates the database connection
func ProvideDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.NewGormDB(&cfg.Database)
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	if err := database.SeedPlans(db); err != nil {
		return nil, fmt.Errorf("failed to seed plans: %w", err)
	}
	return db, nil
}

// ProvideRedis creates the Redis client
func ProvideRedis(cfg *config.Config) (*pkgredis.Client, error) {
	return pkgredis.NewClient(&cfg.Redis)
}

// ProvideQueue creates the Asynq queue client
func ProvideQueue(cfg *config.Config) *queue.Client {
	return queue.NewClient(&cfg.Redis)
}

// BaseURL is a type alias for dependency injection
type BaseURL string

// FrontendURL is a type alias for frontend URL dependency injection
type FrontendURL string

// ProvideBaseURL returns the base URL for the application
func ProvideBaseURL(cfg *config.Config) BaseURL {
	if cfg.App.URL != "" {
		return BaseURL(cfg.App.URL)
	}
	return BaseURL(fmt.Sprintf("http://localhost:%d", cfg.Server.Port))
}

// ProvideFrontendURL returns the frontend URL for redirects
func ProvideFrontendURL(cfg *config.Config) FrontendURL {
	if cfg.App.FrontendURL != "" {
		return FrontendURL(cfg.App.FrontendURL)
	}
	return FrontendURL("http://localhost:3000")
}

// InfraSet provides infrastructure dependencies
var InfraSet = wire.NewSet(
	ProvideConfig,
	ProvideDB,
	ProvideRedis,
	ProvideQueue,
	ProvideBaseURL,
	ProvideFrontendURL,
)
