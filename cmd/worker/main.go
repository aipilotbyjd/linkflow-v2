package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"time"

	asynqAdapter "github.com/linkflow-ai/linkflow/internal/adapters/messaging/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	"github.com/linkflow-ai/linkflow/internal/adapters/websocket"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
	billingapp "github.com/linkflow-ai/linkflow/internal/core/application/billing"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/email"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	sentryPkg "github.com/linkflow-ai/linkflow/internal/infrastructure/observability/sentry"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// nodeAdapter adapts nodes.Node to executor.NodeHandler
type nodeAdapter struct {
	node nodes.Node
}

func (a *nodeAdapter) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	return a.node.Execute(ctx, runtime, node)
}

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "local"
		}
		configPath = fmt.Sprintf("configs/config.%s.yaml", env)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	var appLogger logger.Logger
	if cfg.App.IsDevelopment() {
		appLogger = logger.NewDevelopment()
	} else {
		appLogger = logger.NewDefault()
	}

	appLogger.Info().
		Str("app", cfg.App.Name).
		Str("service", "worker").
		Msg("Starting worker service")

	// Initialize Sentry
	if err := initSentry(cfg); err != nil {
		appLogger.Warn().Err(err).Msg("Failed to initialize Sentry, continuing without error tracking")
	} else if cfg.Sentry.Enabled {
		appLogger.Info().Msg("Sentry error tracking initialized")
		defer sentryPkg.Flush(2 * time.Second)
	}

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
		appLogger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}()

	// Initialize Redis
	redis, err := redisAdapter.NewClient(redisAdapter.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()

	// Initialize repositories
	gormDB, err := db.DB()
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to get raw DB")
	}
	_ = gormDB // db.DB() returns sql.DB, not gorm.DB for repositories

	// Initialize repositories
	executionRepo := repositories.NewExecutionRepository(db)
	nodeExecRepo := repositories.NewNodeExecutionRepository(db)
	workflowRepo := repositories.NewWorkflowRepository(db)
	usageRepo := repositories.NewUsageRepository(db)
	subscriptionRepo := repositories.NewSubscriptionRepository(db)

	// Initialize billing service and usage tracker
	usageService := billingapp.NewUsageService(usageRepo, subscriptionRepo)
	usageTracker := executor.NewUsageTracker(usageService)

	// Initialize email service
	emailService, err := email.NewService(email.Config{
		Provider:    cfg.Email.Provider,
		DefaultFrom: cfg.Email.From,
		SMTPHost:    cfg.Email.SMTPHost,
		SMTPPort:    cfg.Email.SMTPPort,
		SMTPUser:    cfg.Email.SMTPUser,
		SMTPPass:    cfg.Email.SMTPPass,
	})
	if err != nil {
		appLogger.Warn().Err(err).Msg("Failed to initialize email service, using noop")
		emailService, _ = email.NewService(email.Config{Provider: "noop"})
	}

	// Initialize node registry and load all nodes
	nodeRegistry := nodes.NewRegistry()
	if err := nodes.LoadAllNodes(nodeRegistry); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to load nodes")
	}
	appLogger.Info().Int("node_count", len(nodeRegistry.ListTypes())).Msg("Loaded workflow nodes")

	// Initialize processor and register node handlers
	processor := executor.NewProcessor(appLogger)
	for _, nodeType := range nodeRegistry.ListTypes() {
		node, _ := nodeRegistry.Get(nodeType)
		processor.RegisterHandler(nodeType, &nodeAdapter{node: node})
	}

	// Initialize Execution Stream Service with Redis publisher
	redisPublisher := websocket.NewRedisPublisher(redis.Redis())
	executionStreamService := websocket.NewExecutionStreamService(redisPublisher)

	// Initialize executor with usage tracking and streaming
	workflowExecutor := executor.NewExecutor(
		workflowRepo,
		executionRepo,
		nodeExecRepo,
		processor,
		usageTracker,
		executionStreamService,
		appLogger,
	)

	// Initialize Asynq server
	server, err := asynqAdapter.NewServer(asynqAdapter.Config{
		RedisAddr:     cfg.Redis.GetAddress(),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
		Concurrency:   cfg.Features.Execution.WorkerCount,
	})
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to create Asynq server")
	}

	// Register task handlers
	server.HandleFunc(asynqAdapter.TaskExecuteWorkflow, func(ctx context.Context, t *asynq.Task) error {
		var payload asynqAdapter.ExecuteWorkflowPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		appLogger.Info().
			Str("execution_id", payload.ExecutionID).
			Str("workflow_id", payload.WorkflowID).
			Msg("Processing workflow execution")

		execID, err := uuid.Parse(payload.ExecutionID)
		if err != nil {
			return err
		}

		return workflowExecutor.Execute(ctx, execID)
	})

	server.HandleFunc(asynqAdapter.TaskSendEmail, func(ctx context.Context, t *asynq.Task) error {
		var payload asynqAdapter.SendEmailPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		appLogger.Info().
			Str("to", payload.To).
			Str("subject", payload.Subject).
			Msg("Sending email")

		if payload.Template != "" {
			return emailService.SendTemplate(ctx, []string{payload.To}, payload.Template, payload.Data)
		}

		// Fallback for non-template emails (if payload supports it in future)
		return emailService.Send(ctx, &email.Message{
			To:       []string{payload.To},
			Subject:  payload.Subject,
			TextBody: fmt.Sprintf("%v", payload.Data), // Simple fallback
		})
	})

	// Keep references
	_ = redis

	// Start server (non-blocking)
	appLogger.Info().Msg("Starting Asynq worker server")
	if err := server.Start(); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to start worker server")
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info().Msg("Shutting down worker...")
	server.Shutdown()
	appLogger.Info().Msg("Worker stopped gracefully")
}

// initSentry initializes Sentry error tracking
func initSentry(cfg *config.Config) error {
	return sentryPkg.Init(cfg.Sentry, "linkflow-worker")
}
