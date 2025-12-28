package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/linkflow-ai/linkflow/internal/pkg/logger"
	"github.com/linkflow-ai/linkflow/internal/scheduler"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize app with all dependencies (wire-generated)
	app, err := InitializeSchedulerApp()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize scheduler application")
	}

	// Initialize logger
	logger.Init(app.Config.App.Environment, app.Config.App.Debug)

	log.Info().
		Str("app", app.Config.App.Name).
		Str("service", "scheduler").
		Msg("Starting scheduler service")

	// Create scheduler config
	schedulerCfg := scheduler.DefaultConfig()

	// Create scheduler
	s := scheduler.New(schedulerCfg, &scheduler.Dependencies{
		DB:    app.DB,
		Redis: app.Redis,
		Queue: app.Queue,
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
