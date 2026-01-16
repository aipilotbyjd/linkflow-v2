package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
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
	log.Info().Int("steps", steps).Msg("Running migrations up")

	// Enable auto-migrate for development
	// For production, implement file-based migrations
	if err := db.AutoMigrate(
	// Add your domain models here for auto-migration
	// &user.User{},
	// &workspace.Workspace{},
	// etc.
	); err != nil {
		return fmt.Errorf("auto-migrate failed: %w", err)
	}

	return nil
}

func runMigrationsDown(db *gorm.DB, steps int) error {
	log.Info().Int("steps", steps).Msg("Running migrations down")
	// Implement rollback logic using migration files
	// For now, this is a placeholder
	log.Warn().Msg("Rollback not implemented - use migration files")
	return nil
}

func showMigrationStatus(db *gorm.DB) error {
	log.Info().Msg("Migration status")
	// Show current migration version
	// Implement based on your migration tracking table
	return nil
}
