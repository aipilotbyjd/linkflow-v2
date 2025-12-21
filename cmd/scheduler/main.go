package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/linkflow-ai/linkflow/internal/pkg/logger"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/linkflow-ai/linkflow/internal/scheduler"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.Init(cfg.App.Environment, cfg.App.Debug)

	log.Info().
		Str("app", cfg.App.Name).
		Str("service", "scheduler").
		Msg("Starting scheduler service")

	// Connect to database
	db, err := database.NewGormDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Connect to Redis
	redisClient, err := pkgredis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	// Initialize queue client
	queueClient := queue.NewClient(&cfg.Redis)

	// Create scheduler config
	schedulerCfg := scheduler.DefaultConfig()

	// Create scheduler
	s := scheduler.New(schedulerCfg, &scheduler.Dependencies{
		DB:    db,
		Redis: redisClient,
		Queue: queueClient,
	})

	// Start scheduler
	if err := s.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start scheduler")
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Received shutdown signal")

	// Stop scheduler
	if err := s.Stop(); err != nil {
		log.Error().Err(err).Msg("Error stopping scheduler")
	}

	log.Info().Msg("Scheduler stopped")
}
