package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Parse flags
	clean := flag.Bool("clean", true, "Clean existing dev data before seeding")
	adminEmail := flag.String("admin-email", "admin@linkflow.dev", "Admin user email")
	adminPassword := flag.String("admin-password", "Admin123!", "Admin user password")
	flag.Parse()

	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Connect to database
	db, err := database.NewGormDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Run base seeders first (plans, permissions, roles)
	log.Info().Msg("Running base seeders...")
	if err := database.SeedAll(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to run base seeders")
	}

	// Run development seeder
	log.Info().Msg("Running development seeder...")
	seedCfg := database.DevSeedConfig{
		AdminEmail:    *adminEmail,
		AdminPassword: *adminPassword,
		CleanFirst:    *clean,
	}

	if err := database.SeedDevelopment(db, seedCfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed development data")
	}

	fmt.Println("\n✅ Seeding completed successfully!")
	fmt.Println("\nYou can now login with:")
	fmt.Printf("  Email:    %s\n", *adminEmail)
	fmt.Printf("  Password: %s\n", *adminPassword)
}
