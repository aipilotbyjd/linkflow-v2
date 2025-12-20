package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/linkflow-ai/linkflow/internal/pkg/logger"
	"github.com/linkflow-ai/linkflow/internal/worker"
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
		Str("service", "worker").
		Msg("Starting worker service")

	// Connect to database
	db, err := database.NewGormDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Initialize repositories
	workflowRepo := repositories.NewWorkflowRepository(db)
	versionRepo := repositories.NewWorkflowVersionRepository(db)
	executionRepo := repositories.NewExecutionRepository(db)
	nodeExecutionRepo := repositories.NewNodeExecutionRepository(db)
	credentialRepo := repositories.NewCredentialRepository(db)

	// Initialize crypto
	encryptor, err := crypto.NewEncryptor(cfg.JWT.Secret[:32])
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create encryptor")
	}

	// Initialize services
	workflowSvc := services.NewWorkflowService(workflowRepo, versionRepo)
	executionSvc := services.NewExecutionService(executionRepo, nodeExecutionRepo, workflowRepo)
	credentialSvc := services.NewCredentialService(credentialRepo, encryptor)

	// Create worker
	w := worker.New(cfg, executionSvc, credentialSvc, workflowSvc)

	// Handle shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Info().Msg("Shutting down worker...")
		w.Shutdown()
	}()

	// Start worker
	if err := w.Start(); err != nil {
		log.Fatal().Err(err).Msg("Worker error")
	}
}
