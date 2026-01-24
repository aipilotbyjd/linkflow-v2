//go:build wireinject
// +build wireinject

package main

import (
	"fmt"
	"os"

	"github.com/google/wire"
	asynqAdapter "github.com/linkflow-ai/linkflow/internal/adapters/messaging/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	"github.com/linkflow-ai/linkflow/internal/adapters/websocket"
	executionCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
	"gorm.io/gorm"
)

// WorkerApp holds all dependencies for the worker service
type WorkerApp struct {
	Config    *config.Config
	Logger    logger.Logger
	DB        *postgres.Client
	Redis     *redisAdapter.Client
	Queue     *asynqAdapter.Client
	Encryptor *crypto.Encryptor

	// Repositories
	WorkflowRepo      workflow.Repository
	ExecutionRepo     execution.Repository
	NodeExecutionRepo execution.NodeExecutionRepository
	CredentialRepo    credential.Repository
	ScheduleRepo      schedule.Repository

	// Command Handlers
	StartExecutionHandler *executionCmd.StartExecutionHandler
}

// Infrastructure provider set
var workerInfraSet = wire.NewSet(
	provideWorkerConfig,
	provideWorkerLogger,
	provideWorkerGormDB,
	provideWorkerPostgresClient,
	provideWorkerRedis,
	provideWorkerAsynqClient,
	provideWorkerTaskQueue,
	provideWorkerEncryptor,
	provideWorkerEncryptor,
	provideWorkerEventBus,
	websocket.NewRedisPublisher,
	wire.Bind(new(websocket.EventPublisher), new(*websocket.RedisPublisher)),
	websocket.NewExecutionStreamService,
	wire.Bind(new(executionCmd.ExecutionStreamService), new(*websocket.ExecutionStreamService)),
)

// Repository provider set
var workerRepoSet = wire.NewSet(
	repositories.NewWorkflowRepository,
	repositories.NewExecutionRepository,
	repositories.NewNodeExecutionRepository,
	repositories.NewCredentialRepository,
	repositories.NewScheduleRepository,

	wire.Bind(new(workflow.Repository), new(*repositories.WorkflowRepository)),
	wire.Bind(new(execution.Repository), new(*repositories.ExecutionRepository)),
	wire.Bind(new(execution.NodeExecutionRepository), new(*repositories.NodeExecutionRepository)),
	wire.Bind(new(credential.Repository), new(*repositories.CredentialRepository)),
	wire.Bind(new(schedule.Repository), new(*repositories.ScheduleRepository)),
)

// Command handler provider set
var workerCommandSet = wire.NewSet(
	executionCmd.NewStartExecutionHandler,
)

func provideWorkerConfig() (*config.Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "local"
		}
		configPath = fmt.Sprintf("configs/config.%s.yaml", env)
	}
	return config.Load(configPath)
}

func provideWorkerLogger(cfg *config.Config) logger.Logger {
	if cfg.App.IsDevelopment() {
		return logger.NewDevelopment()
	}
	return logger.NewDefault()
}

func provideWorkerGormDB(cfg *config.Config) (*gorm.DB, error) {
	return postgres.NewClient(postgres.Config{
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
}

func provideWorkerPostgresClient(db *gorm.DB) *postgres.Client {
	return postgres.NewClientWrapper(db)
}

func provideWorkerRedis(cfg *config.Config) (*redisAdapter.Client, error) {
	return redisAdapter.NewClient(redisAdapter.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
}

func provideWorkerAsynqClient(cfg *config.Config) (*asynqAdapter.Client, error) {
	return asynqAdapter.NewClient(asynqAdapter.Config{
		RedisAddr:     cfg.Redis.GetAddress(),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
	})
}

func provideWorkerTaskQueue(client *asynqAdapter.Client) executionCmd.TaskQueue {
	return asynqAdapter.NewTaskQueueAdapter(client)
}

func provideWorkerEncryptor(cfg *config.Config) (*crypto.Encryptor, error) {
	return crypto.NewEncryptor(cfg.App.SecretKey)
}

func provideWorkerEventBus() events.Bus {
	return events.NewInMemoryBus()
}

func provideWorkerApp(
	cfg *config.Config,
	log logger.Logger,
	db *postgres.Client,
	redis *redisAdapter.Client,
	queue *asynqAdapter.Client,
	encryptor *crypto.Encryptor,
	workflowRepo workflow.Repository,
	executionRepo execution.Repository,
	nodeExecRepo execution.NodeExecutionRepository,
	credentialRepo credential.Repository,
	scheduleRepo schedule.Repository,
	startExecutionHandler *executionCmd.StartExecutionHandler,
) *WorkerApp {
	return &WorkerApp{
		Config:                cfg,
		Logger:                log,
		DB:                    db,
		Redis:                 redis,
		Queue:                 queue,
		Encryptor:             encryptor,
		WorkflowRepo:          workflowRepo,
		ExecutionRepo:         executionRepo,
		NodeExecutionRepo:     nodeExecRepo,
		CredentialRepo:        credentialRepo,
		ScheduleRepo:          scheduleRepo,
		StartExecutionHandler: startExecutionHandler,
	}
}

// InitializeWorkerApp wires all dependencies
func InitializeWorkerApp() (*WorkerApp, error) {
	wire.Build(
		workerInfraSet,
		workerRepoSet,
		workerCommandSet,
		provideWorkerApp,
	)
	return nil, nil
}
