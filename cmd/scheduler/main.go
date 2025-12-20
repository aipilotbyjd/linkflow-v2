package main

import (
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/linkflow-ai/linkflow/internal/pkg/logger"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/linkflow-ai/linkflow/internal/queue"
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

	// Initialize repositories
	workflowRepo := repositories.NewWorkflowRepository(db)
	executionRepo := repositories.NewExecutionRepository(db)
	nodeExecutionRepo := repositories.NewNodeExecutionRepository(db)
	scheduleRepo := repositories.NewScheduleRepository(db)

	// Initialize services
	executionSvc := services.NewExecutionService(executionRepo, nodeExecutionRepo, workflowRepo)
	scheduleSvc := services.NewScheduleService(scheduleRepo)

	// Create scheduler
	s := scheduler.New(cfg, redisClient, scheduleSvc, executionSvc, queueClient)

	// Start scheduler
	if err := s.Start(); err != nil {
		log.Fatal().Err(err).Msg("Scheduler error")
	}
}
