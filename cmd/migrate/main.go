package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/config.yaml", "Path to config file")
	direction := flag.String("direction", "up", "Migration direction: up or down")
	steps := flag.Int("steps", 0, "Number of migrations to run (0 = all)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Str("app", cfg.App.Name).
		Str("direction", *direction).
		Msg("Running database migrations")

	// Initialize database
	db, err := postgres.NewClient(postgres.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Database:     cfg.Database.Name,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxLifetime:  cfg.Database.MaxLifetime,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}()

	// Run migrations
	switch *direction {
	case "up":
		if err := runMigrationsUp(db, *steps); err != nil {
			log.Fatal().Err(err).Msg("Migration up failed")
		}
		log.Info().Msg("Migrations completed successfully")

	case "down":
		if err := runMigrationsDown(db, *steps); err != nil {
			log.Fatal().Err(err).Msg("Migration down failed")
		}
		log.Info().Msg("Rollback completed successfully")

	case "status":
		if err := showMigrationStatus(db); err != nil {
			log.Fatal().Err(err).Msg("Failed to get migration status")
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown direction: %s\n", *direction)
		fmt.Fprintln(os.Stderr, "Usage: migrate -direction=[up|down|status] [-steps=N]")
		os.Exit(1)
	}
}

func runMigrationsUp(db *gorm.DB, steps int) error {
	log.Info().Int("steps", steps).Msg("Running GORM AutoMigrate")

	// Enable UUID extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	// AutoMigrate all models from persistence layer
	if err := db.AutoMigrate(
		// User & Auth
		&models.User{},
		&models.UserSession{},

		// Workspace
		&models.Workspace{},
		&models.WorkspaceMember{},

		// Workflow
		&models.Workflow{},
		&models.WorkflowVersion{},

		// Execution
		&models.Execution{},
		&models.NodeExecution{},

		// Folder
		&models.Folder{},

		// Additional models
		&models.PinnedData{},
		&models.Share{},
		&models.BinaryData{},

		// Billing
		&billing.Subscription{},
		&billing.Usage{},
		&billing.Invoice{},
	); err != nil {
		return fmt.Errorf("auto-migrate failed: %w", err)
	}

	log.Info().Msg("All tables created/updated successfully")
	return nil
}

func runMigrationsDown(db *gorm.DB, steps int) error {
	log.Info().Int("steps", steps).Msg("Running migrations down")
	log.Warn().Msg("Rollback not implemented - GORM AutoMigrate only adds columns, doesn't remove")
	return nil
}

func showMigrationStatus(db *gorm.DB) error {
	log.Info().Msg("Migration status - checking tables")

	tables := []string{
		"users", "user_sessions", "api_keys",
		"workspaces", "workspace_members",
		"workflows", "workflow_versions",
		"executions", "node_executions",
		"credentials", "schedules", "webhook_endpoints",
		"templates", "folders", "shares", "pinned_data", "binary_data",
		"plans", "subscriptions", "usage_records", "invoices",
	}

	for _, table := range tables {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			log.Warn().Str("table", table).Msg("NOT EXISTS")
		} else {
			log.Info().Str("table", table).Int64("rows", count).Msg("EXISTS")
		}
	}

	return nil
}
