package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/linkflow-ai/linkflow/internal/pkg/email"
	"github.com/linkflow-ai/linkflow/internal/pkg/logger"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/linkflow-ai/linkflow/internal/pkg/streams"
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

	// Connect to Redis
	redisClient, err := pkgredis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	// Initialize Asynq client for email queue
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer asynqClient.Close()

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

	// Initialize email service
	emailCfg := &email.Config{
		SMTPHost:     cfg.SMTP.Host,
		SMTPPort:     cfg.SMTP.Port,
		SMTPUser:     cfg.SMTP.Username,
		SMTPPassword: cfg.SMTP.Password,
		FromEmail:    cfg.SMTP.From,
		FromName:     cfg.SMTP.FromName,
		QueueEnabled: true,
	}
	emailSvc := email.NewService(emailCfg, asynqClient)

	// Initialize queue client for webhook consumer
	queueClient := queue.NewClient(&cfg.Redis)
	defer queueClient.Close()

	// Initialize webhook stream consumers if enabled
	var webhookConsumers []*streams.WebhookConsumer
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.Features.WebhookStream.Enabled {
		webhookStream := streams.NewWebhookStream(redisClient.Client)

		// Start multiple consumers based on config
		consumerCount := cfg.Features.WebhookStream.ConsumerCount
		if consumerCount < 1 {
			consumerCount = 2
		}

		for i := 0; i < consumerCount; i++ {
			consumerName := fmt.Sprintf("worker-%d-consumer-%d", os.Getpid(), i)
			consumer := streams.NewWebhookConsumer(webhookStream, workflowSvc, queueClient, consumerName)

			if err := consumer.Start(ctx); err != nil {
				log.Error().Err(err).Int("consumer", i).Msg("Failed to start webhook consumer")
				continue
			}

			webhookConsumers = append(webhookConsumers, consumer)
			log.Info().Str("consumer", consumerName).Msg("Webhook stream consumer started")
		}

		log.Info().Int("count", len(webhookConsumers)).Msg("Webhook stream consumers running")
	}

	// Create worker
	w := worker.New(cfg, executionSvc, credentialSvc, workflowSvc, redisClient.Client, emailSvc)

	// Handle shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Info().Msg("Shutting down worker...")

		// Stop webhook consumers first
		cancel()
		for _, consumer := range webhookConsumers {
			consumer.Stop()
		}

		w.Shutdown()
	}()

	// Start worker
	if err := w.Start(); err != nil {
		log.Fatal().Err(err).Msg("Worker error")
	}
}
