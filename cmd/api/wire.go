//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/adapters/messaging/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	userCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/user"
	workflowCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workflow"
	workspaceCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workspace"
	executionCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	userQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/user"
	workflowQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/workflow"
	workspaceQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/workspace"
	executionQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
	"gorm.io/gorm"
)

// App holds all dependencies for the API server
type App struct {
	Config  *config.Config
	Logger  logger.Logger
	DB      *postgres.Client
	Redis   *redisAdapter.Client
	Queue   *asynq.Client
	Cache   cache.Cache
	JWT     *jwt.Manager
	Hasher  *crypto.Hasher

	// Repositories
	UserRepo       user.Repository
	SessionRepo    user.SessionRepository
	WorkspaceRepo  workspace.Repository
	MemberRepo     workspace.MemberRepository
	WorkflowRepo   workflow.Repository
	VersionRepo    workflow.VersionRepository
	ExecutionRepo  execution.Repository
	NodeExecRepo   execution.NodeExecutionRepository
	CredentialRepo credential.Repository
	ScheduleRepo   schedule.Repository
	WebhookRepo    webhook.Repository

	// Command Handlers
	RegisterUserHandler     *userCmd.RegisterUserHandler
	LoginUserHandler        *userCmd.LoginUserHandler
	CreateWorkflowHandler   *workflowCmd.CreateWorkflowHandler
	UpdateWorkflowHandler   *workflowCmd.UpdateWorkflowHandler
	ActivateWorkflowHandler *workflowCmd.ActivateWorkflowHandler
	CreateWorkspaceHandler  *workspaceCmd.CreateWorkspaceHandler
	StartExecutionHandler   *executionCmd.StartExecutionHandler

	// Query Handlers
	GetUserHandler        *userQuery.GetUserHandler
	GetWorkflowHandler    *workflowQuery.GetWorkflowHandler
	ListWorkflowsHandler  *workflowQuery.ListWorkflowsHandler
	GetVersionsHandler    *workflowQuery.GetVersionsHandler
	GetWorkspaceHandler   *workspaceQuery.GetWorkspaceHandler
	ListWorkspacesHandler *workspaceQuery.ListWorkspacesHandler
	ListMembersHandler    *workspaceQuery.ListMembersHandler
	GetExecutionHandler   *executionQuery.GetExecutionHandler
	ListExecutionsHandler *executionQuery.ListExecutionsHandler
}

// Infrastructure provider set
var infraSet = wire.NewSet(
	provideConfig,
	provideLogger,
	providePostgres,
	provideGormDB,
	provideRedis,
	provideAsynqClient,
	provideCache,
	provideJWTManager,
	provideHasher,
	provideEventBus,
)

// Repository provider set
var repoSet = wire.NewSet(
	repositories.NewUserRepository,
	repositories.NewSessionRepository,
	repositories.NewWorkspaceRepository,
	repositories.NewMemberRepository,
	repositories.NewWorkflowRepository,
	repositories.NewVersionRepository,
	repositories.NewExecutionRepository,
	repositories.NewNodeExecutionRepository,
	repositories.NewCredentialRepository,
	repositories.NewScheduleRepository,
	repositories.NewWebhookRepository,

	wire.Bind(new(user.Repository), new(*repositories.UserRepository)),
	wire.Bind(new(user.SessionRepository), new(*repositories.SessionRepository)),
	wire.Bind(new(workspace.Repository), new(*repositories.WorkspaceRepository)),
	wire.Bind(new(workspace.MemberRepository), new(*repositories.MemberRepository)),
	wire.Bind(new(workflow.Repository), new(*repositories.WorkflowRepository)),
	wire.Bind(new(workflow.VersionRepository), new(*repositories.VersionRepository)),
	wire.Bind(new(execution.Repository), new(*repositories.ExecutionRepository)),
	wire.Bind(new(execution.NodeExecutionRepository), new(*repositories.NodeExecutionRepository)),
	wire.Bind(new(credential.Repository), new(*repositories.CredentialRepository)),
	wire.Bind(new(schedule.Repository), new(*repositories.ScheduleRepository)),
	wire.Bind(new(webhook.Repository), new(*repositories.WebhookRepository)),
)

// Command handler provider set
var commandSet = wire.NewSet(
	userCmd.NewRegisterUserHandler,
	userCmd.NewLoginUserHandler,
	workflowCmd.NewCreateWorkflowHandler,
	workflowCmd.NewUpdateWorkflowHandler,
	workflowCmd.NewActivateWorkflowHandler,
	workflowCmd.NewDeactivateWorkflowHandler,
	workspaceCmd.NewCreateWorkspaceHandler,
	executionCmd.NewStartExecutionHandler,
)

// Query handler provider set
var querySet = wire.NewSet(
	userQuery.NewGetUserHandler,
	workflowQuery.NewGetWorkflowHandler,
	workflowQuery.NewListWorkflowsHandler,
	workflowQuery.NewGetVersionsHandler,
	workspaceQuery.NewGetWorkspaceHandler,
	workspaceQuery.NewListWorkspacesHandler,
	workspaceQuery.NewListMembersHandler,
	executionQuery.NewGetExecutionHandler,
	executionQuery.NewListExecutionsHandler,
	executionQuery.NewGetNodeExecutionsHandler,
)

func provideConfig() (*config.Config, error) {
	configPath := "configs/config.yaml"
	return config.Load(configPath)
}

func provideLogger(cfg *config.Config) logger.Logger {
	if cfg.App.IsDevelopment() {
		return logger.NewDevelopment()
	}
	return logger.NewDefault()
}

func providePostgres(cfg *config.Config) (*postgres.Client, error) {
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

func provideGormDB(client *postgres.Client) *gorm.DB {
	return client.DB()
}

func provideRedis(cfg *config.Config) (*redisAdapter.Client, error) {
	return redisAdapter.NewClient(redisAdapter.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
}

func provideAsynqClient(cfg *config.Config) (*asynq.Client, error) {
	return asynq.NewClient(asynq.Config{
		RedisAddr:     cfg.Redis.GetAddress(),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
	})
}

func provideCache(redis *redisAdapter.Client) cache.Cache {
	return cache.NewRedisCache(redis.Redis(), "linkflow")
}

func provideJWTManager(cfg *config.Config) *jwt.Manager {
	return jwt.NewManager(jwt.Config{
		Secret:        cfg.JWT.Secret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
		Issuer:        cfg.JWT.Issuer,
	})
}

func provideHasher() *crypto.Hasher {
	return crypto.NewHasher(10)
}

func provideEventBus() events.Bus {
	return events.NewInMemoryBus()
}

func provideApp(
	cfg *config.Config,
	log logger.Logger,
	db *postgres.Client,
	redis *redisAdapter.Client,
	queue *asynq.Client,
	cacheClient cache.Cache,
	jwtMgr *jwt.Manager,
	hasher *crypto.Hasher,
	userRepo user.Repository,
	sessionRepo user.SessionRepository,
	workspaceRepo workspace.Repository,
	memberRepo workspace.MemberRepository,
	workflowRepo workflow.Repository,
	versionRepo workflow.VersionRepository,
	executionRepo execution.Repository,
	nodeExecRepo execution.NodeExecutionRepository,
	credentialRepo credential.Repository,
	scheduleRepo schedule.Repository,
	webhookRepo webhook.Repository,
	registerUserHandler *userCmd.RegisterUserHandler,
	loginUserHandler *userCmd.LoginUserHandler,
	createWorkflowHandler *workflowCmd.CreateWorkflowHandler,
	updateWorkflowHandler *workflowCmd.UpdateWorkflowHandler,
	activateWorkflowHandler *workflowCmd.ActivateWorkflowHandler,
	createWorkspaceHandler *workspaceCmd.CreateWorkspaceHandler,
	startExecutionHandler *executionCmd.StartExecutionHandler,
	getUserHandler *userQuery.GetUserHandler,
	getWorkflowHandler *workflowQuery.GetWorkflowHandler,
	listWorkflowsHandler *workflowQuery.ListWorkflowsHandler,
	getVersionsHandler *workflowQuery.GetVersionsHandler,
	getWorkspaceHandler *workspaceQuery.GetWorkspaceHandler,
	listWorkspacesHandler *workspaceQuery.ListWorkspacesHandler,
	listMembersHandler *workspaceQuery.ListMembersHandler,
	getExecutionHandler *executionQuery.GetExecutionHandler,
	listExecutionsHandler *executionQuery.ListExecutionsHandler,
) *App {
	return &App{
		Config:                  cfg,
		Logger:                  log,
		DB:                      db,
		Redis:                   redis,
		Queue:                   queue,
		Cache:                   cacheClient,
		JWT:                     jwtMgr,
		Hasher:                  hasher,
		UserRepo:                userRepo,
		SessionRepo:             sessionRepo,
		WorkspaceRepo:           workspaceRepo,
		MemberRepo:              memberRepo,
		WorkflowRepo:            workflowRepo,
		VersionRepo:             versionRepo,
		ExecutionRepo:           executionRepo,
		NodeExecRepo:            nodeExecRepo,
		CredentialRepo:          credentialRepo,
		ScheduleRepo:            scheduleRepo,
		WebhookRepo:             webhookRepo,
		RegisterUserHandler:     registerUserHandler,
		LoginUserHandler:        loginUserHandler,
		CreateWorkflowHandler:   createWorkflowHandler,
		UpdateWorkflowHandler:   updateWorkflowHandler,
		ActivateWorkflowHandler: activateWorkflowHandler,
		CreateWorkspaceHandler:  createWorkspaceHandler,
		StartExecutionHandler:   startExecutionHandler,
		GetUserHandler:          getUserHandler,
		GetWorkflowHandler:      getWorkflowHandler,
		ListWorkflowsHandler:    listWorkflowsHandler,
		GetVersionsHandler:      getVersionsHandler,
		GetWorkspaceHandler:     getWorkspaceHandler,
		ListWorkspacesHandler:   listWorkspacesHandler,
		ListMembersHandler:      listMembersHandler,
		GetExecutionHandler:     getExecutionHandler,
		ListExecutionsHandler:   listExecutionsHandler,
	}
}

// InitializeApp wires all dependencies
func InitializeApp() (*App, error) {
	wire.Build(
		infraSet,
		repoSet,
		commandSet,
		querySet,
		provideApp,
	)
	return nil, nil
}
