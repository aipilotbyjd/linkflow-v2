//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	asynqAdapter "github.com/linkflow-ai/linkflow/internal/adapters/messaging/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"gorm.io/gorm"
)

// SchedulerApp holds all dependencies for the scheduler service
type SchedulerApp struct {
	Config *config.Config
	Logger logger.Logger
	DB     *postgres.Client
	Redis  *redisAdapter.Client
	Queue  *asynqAdapter.Client

	// Repositories
	ScheduleRepo schedule.Repository
	WorkflowRepo workflow.Repository
}

// Infrastructure provider set
var schedulerInfraSet = wire.NewSet(
	provideSchedulerConfig,
	provideSchedulerLogger,
	provideSchedulerPostgres,
	provideSchedulerGormDB,
	provideSchedulerRedis,
	provideSchedulerAsynqClient,
)

// Repository provider set
var schedulerRepoSet = wire.NewSet(
	repositories.NewScheduleRepository,
	repositories.NewWorkflowRepository,

	wire.Bind(new(schedule.Repository), new(*repositories.ScheduleRepository)),
	wire.Bind(new(workflow.Repository), new(*repositories.WorkflowRepository)),
)

func provideSchedulerConfig() (*config.Config, error) {
	configPath := "configs/config.yaml"
	return config.Load(configPath)
}

func provideSchedulerLogger(cfg *config.Config) logger.Logger {
	if cfg.App.IsDevelopment() {
		return logger.NewDevelopment()
	}
	return logger.NewDefault()
}

func provideSchedulerPostgres(cfg *config.Config) (*postgres.Client, error) {
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

func provideSchedulerGormDB(client *postgres.Client) *gorm.DB {
	return client.DB()
}

func provideSchedulerRedis(cfg *config.Config) (*redisAdapter.Client, error) {
	return redisAdapter.NewClient(redisAdapter.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
}

func provideSchedulerAsynqClient(cfg *config.Config) (*asynqAdapter.Client, error) {
	return asynqAdapter.NewClient(asynqAdapter.Config{
		RedisAddr:     cfg.Redis.GetAddress(),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
	})
}

func provideSchedulerApp(
	cfg *config.Config,
	log logger.Logger,
	db *postgres.Client,
	redis *redisAdapter.Client,
	queue *asynqAdapter.Client,
	scheduleRepo schedule.Repository,
	workflowRepo workflow.Repository,
) *SchedulerApp {
	return &SchedulerApp{
		Config:       cfg,
		Logger:       log,
		DB:           db,
		Redis:        redis,
		Queue:        queue,
		ScheduleRepo: scheduleRepo,
		WorkflowRepo: workflowRepo,
	}
}

// InitializeSchedulerApp wires all dependencies
func InitializeSchedulerApp() (*SchedulerApp, error) {
	wire.Build(
		schedulerInfraSet,
		schedulerRepoSet,
		provideSchedulerApp,
	)
	return nil, nil
}
