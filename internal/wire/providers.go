package wire

import (
	"fmt"

	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/api"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"gorm.io/gorm"
)

// ProvideConfig loads the application configuration
func ProvideConfig() (*config.Config, error) {
	return config.Load()
}

// ProvideDB creates the database connection
func ProvideDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.NewGormDB(&cfg.Database)
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	if err := database.SeedPlans(db); err != nil {
		return nil, fmt.Errorf("failed to seed plans: %w", err)
	}
	return db, nil
}

// ProvideRedis creates the Redis client
func ProvideRedis(cfg *config.Config) (*pkgredis.Client, error) {
	return pkgredis.NewClient(&cfg.Redis)
}

// ProvideQueue creates the Asynq queue client
func ProvideQueue(cfg *config.Config) *queue.Client {
	return queue.NewClient(&cfg.Redis)
}

// ProvideJWTManager creates the JWT manager
func ProvideJWTManager(cfg *config.Config) *crypto.JWTManager {
	return crypto.NewJWTManager(crypto.JWTConfig{
		Secret:        cfg.JWT.Secret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
		Issuer:        cfg.JWT.Issuer,
	})
}

// ProvideEncryptor creates the AES encryptor
func ProvideEncryptor(cfg *config.Config) (*crypto.Encryptor, error) {
	return crypto.NewEncryptor(cfg.JWT.Secret[:32])
}

// ProvideOTPManager creates the OTP manager
func ProvideOTPManager(cfg *config.Config) *crypto.OTPManager {
	return crypto.NewOTPManager(cfg.App.Name)
}

// BaseURL is a type alias for dependency injection
type BaseURL string

// ProvideBaseURL returns the base URL for the application
func ProvideBaseURL(cfg *config.Config) BaseURL {
	if cfg.App.URL != "" {
		return BaseURL(cfg.App.URL)
	}
	return BaseURL(fmt.Sprintf("http://localhost:%d", cfg.Server.Port))
}

// InfraSet provides infrastructure dependencies
var InfraSet = wire.NewSet(
	ProvideConfig,
	ProvideDB,
	ProvideRedis,
	ProvideQueue,
	ProvideBaseURL,
)

// CryptoSet provides crypto dependencies
var CryptoSet = wire.NewSet(
	ProvideJWTManager,
	ProvideEncryptor,
	ProvideOTPManager,
)

// RepositorySet provides all repository dependencies
var RepositorySet = wire.NewSet(
	repositories.NewUserRepository,
	repositories.NewSessionRepository,
	repositories.NewWorkspaceRepository,
	repositories.NewWorkspaceMemberRepository,
	repositories.NewWorkspaceInvitationRepository,
	repositories.NewWorkflowRepository,
	repositories.NewWorkflowVersionRepository,
	repositories.NewExecutionRepository,
	repositories.NewNodeExecutionRepository,
	repositories.NewCredentialRepository,
	repositories.NewScheduleRepository,
	repositories.NewPlanRepository,
	repositories.NewSubscriptionRepository,
	repositories.NewUsageRepository,
	repositories.NewInvoiceRepository,
	repositories.NewPinnedDataRepository,
	repositories.NewWaitingExecutionRepository,
	repositories.NewTemplateRepository,
	repositories.NewOAuthStateRepository,
	repositories.NewWebhookEndpointRepository,
	ProvideAuditLogRepo,
	ProvideAlertRepo,
	ProvideAlertLogRepo,
	ProvideCommentRepo,
	ProvideExecShareRepo,
	ProvideEnvVarRepo,
	ProvideWorkspaceAnalyticsRepo,
	ProvideWorkflowAnalyticsRepo,
	ProvideWorkflowExportRepo,
	ProvideWorkflowImportRepo,
)

// Generic repository providers
func ProvideAuditLogRepo(db *gorm.DB) *repositories.BaseRepository[models.AuditLog] {
	return repositories.NewBaseRepository[models.AuditLog](db)
}

func ProvideAlertRepo(db *gorm.DB) *repositories.BaseRepository[models.Alert] {
	return repositories.NewBaseRepository[models.Alert](db)
}

func ProvideAlertLogRepo(db *gorm.DB) *repositories.BaseRepository[models.AlertLog] {
	return repositories.NewBaseRepository[models.AlertLog](db)
}

func ProvideCommentRepo(db *gorm.DB) *repositories.BaseRepository[models.WorkflowComment] {
	return repositories.NewBaseRepository[models.WorkflowComment](db)
}

func ProvideExecShareRepo(db *gorm.DB) *repositories.BaseRepository[models.ExecutionShare] {
	return repositories.NewBaseRepository[models.ExecutionShare](db)
}

func ProvideEnvVarRepo(db *gorm.DB) *repositories.BaseRepository[models.EnvironmentVariable] {
	return repositories.NewBaseRepository[models.EnvironmentVariable](db)
}

func ProvideWorkspaceAnalyticsRepo(db *gorm.DB) *repositories.BaseRepository[models.WorkspaceAnalytics] {
	return repositories.NewBaseRepository[models.WorkspaceAnalytics](db)
}

func ProvideWorkflowAnalyticsRepo(db *gorm.DB) *repositories.BaseRepository[models.WorkflowAnalytics] {
	return repositories.NewBaseRepository[models.WorkflowAnalytics](db)
}

func ProvideWorkflowExportRepo(db *gorm.DB) *repositories.BaseRepository[models.WorkflowExport] {
	return repositories.NewBaseRepository[models.WorkflowExport](db)
}

func ProvideWorkflowImportRepo(db *gorm.DB) *repositories.BaseRepository[models.WorkflowImport] {
	return repositories.NewBaseRepository[models.WorkflowImport](db)
}

// ServiceSet provides all service dependencies
var ServiceSet = wire.NewSet(
	services.NewAuthService,
	services.NewUserService,
	services.NewWorkspaceService,
	ProvideWorkflowService,
	services.NewExecutionService,
	services.NewCredentialService,
	services.NewScheduleService,
	ProvideBillingService,
	ProvideOAuthService,
	services.NewTemplateService,
	ProvideWebhookManager,
	ProvideWaitResumeManager,
	services.NewAuditLogService,
	services.NewAlertService,
	services.NewWorkflowCommentService,
	services.NewExecutionShareService,
	ProvideEnvVarService,
	ProvideAnalyticsService,
	ProvideExportImportService,
)

// ProvideWorkflowService creates the workflow service with webhook repo
func ProvideWorkflowService(
	workflowRepo *repositories.WorkflowRepository,
	versionRepo *repositories.WorkflowVersionRepository,
	webhookRepo *repositories.WebhookEndpointRepository,
) *services.WorkflowService {
	svc := services.NewWorkflowService(workflowRepo, versionRepo)
	svc.SetWebhookEndpointRepo(webhookRepo)
	return svc
}

// ProvideBillingService creates the billing service with counting repos
func ProvideBillingService(
	planRepo *repositories.PlanRepository,
	subscriptionRepo *repositories.SubscriptionRepository,
	usageRepo *repositories.UsageRepository,
	invoiceRepo *repositories.InvoiceRepository,
	workspaceRepo *repositories.WorkspaceRepository,
	workflowRepo *repositories.WorkflowRepository,
	memberRepo *repositories.WorkspaceMemberRepository,
	credentialRepo *repositories.CredentialRepository,
) *services.BillingService {
	svc := services.NewBillingService(planRepo, subscriptionRepo, usageRepo, invoiceRepo, workspaceRepo)
	svc.SetCountingRepos(workflowRepo, memberRepo, credentialRepo)
	return svc
}

// ProvideOAuthService creates the OAuth service with base URL
func ProvideOAuthService(
	stateRepo *repositories.OAuthStateRepository,
	credentialRepo *repositories.CredentialRepository,
	baseURL BaseURL,
) *services.OAuthService {
	return services.NewOAuthService(stateRepo, credentialRepo, string(baseURL))
}

// ProvideWebhookManager creates the webhook manager with base URL
func ProvideWebhookManager(
	webhookRepo *repositories.WebhookEndpointRepository,
	baseURL BaseURL,
) *services.WebhookManager {
	return services.NewWebhookManager(webhookRepo, string(baseURL))
}

// ProvideWaitResumeManager creates the wait/resume manager with base URL
func ProvideWaitResumeManager(
	waitingRepo *repositories.WaitingExecutionRepository,
	baseURL BaseURL,
) *services.WaitResumeManager {
	return services.NewWaitResumeManager(waitingRepo, string(baseURL))
}

// ProvideEnvVarService creates the environment variable service
func ProvideEnvVarService(
	repo *repositories.BaseRepository[models.EnvironmentVariable],
	encryptor *crypto.Encryptor,
) *services.EnvironmentVariableService {
	return services.NewEnvironmentVariableService(repo, encryptor)
}

// ProvideAnalyticsService creates the analytics service
func ProvideAnalyticsService(
	workspaceRepo *repositories.BaseRepository[models.WorkspaceAnalytics],
	workflowRepo *repositories.BaseRepository[models.WorkflowAnalytics],
	executionRepo *repositories.ExecutionRepository,
) *services.AnalyticsService {
	return services.NewAnalyticsService(workspaceRepo, workflowRepo, executionRepo)
}

// ProvideExportImportService creates the workflow export/import service
func ProvideExportImportService(
	exportRepo *repositories.BaseRepository[models.WorkflowExport],
	importRepo *repositories.BaseRepository[models.WorkflowImport],
	workflowRepo *repositories.WorkflowRepository,
) *services.WorkflowExportService {
	return services.NewWorkflowExportService(exportRepo, importRepo, workflowRepo)
}

// Services aggregates all services for the API server
type Services struct {
	Auth          *services.AuthService
	User          *services.UserService
	Workspace     *services.WorkspaceService
	Workflow      *services.WorkflowService
	Execution     *services.ExecutionService
	Credential    *services.CredentialService
	Schedule      *services.ScheduleService
	Billing       *services.BillingService
	OAuth         *services.OAuthService
	Template      *services.TemplateService
	WebhookMgr    *services.WebhookManager
	WaitResumeMgr *services.WaitResumeManager
	AuditLog      *services.AuditLogService
	Alert         *services.AlertService
	Comment       *services.WorkflowCommentService
	ExecShare     *services.ExecutionShareService
	EnvVar        *services.EnvironmentVariableService
	Analytics     *services.AnalyticsService
	ExportImport  *services.WorkflowExportService
}

// ProvideServices aggregates all services
func ProvideServices(
	auth *services.AuthService,
	user *services.UserService,
	workspace *services.WorkspaceService,
	workflow *services.WorkflowService,
	execution *services.ExecutionService,
	credential *services.CredentialService,
	schedule *services.ScheduleService,
	billing *services.BillingService,
	oauth *services.OAuthService,
	template *services.TemplateService,
	webhookMgr *services.WebhookManager,
	waitResumeMgr *services.WaitResumeManager,
	auditLog *services.AuditLogService,
	alert *services.AlertService,
	comment *services.WorkflowCommentService,
	execShare *services.ExecutionShareService,
	envVar *services.EnvironmentVariableService,
	analytics *services.AnalyticsService,
	exportImport *services.WorkflowExportService,
) *Services {
	return &Services{
		Auth:          auth,
		User:          user,
		Workspace:     workspace,
		Workflow:      workflow,
		Execution:     execution,
		Credential:    credential,
		Schedule:      schedule,
		Billing:       billing,
		OAuth:         oauth,
		Template:      template,
		WebhookMgr:    webhookMgr,
		WaitResumeMgr: waitResumeMgr,
		AuditLog:      auditLog,
		Alert:         alert,
		Comment:       comment,
		ExecShare:     execShare,
		EnvVar:        envVar,
		Analytics:     analytics,
		ExportImport:  exportImport,
	}
}

// Repositories aggregates repositories needed by handlers
type Repositories struct {
	PinnedData      *repositories.PinnedDataRepository
	WaitingExec     *repositories.WaitingExecutionRepository
	WebhookEndpoint *repositories.WebhookEndpointRepository
}

// ProvideRepositories aggregates handler repositories
func ProvideRepositories(
	pinnedData *repositories.PinnedDataRepository,
	waitingExec *repositories.WaitingExecutionRepository,
	webhookEndpoint *repositories.WebhookEndpointRepository,
) *Repositories {
	return &Repositories{
		PinnedData:      pinnedData,
		WaitingExec:     waitingExec,
		WebhookEndpoint: webhookEndpoint,
	}
}

// ProvideAPIServices converts wire.Services to api.Services
func ProvideAPIServices(s *Services) *api.Services {
	return &api.Services{
		Auth:          s.Auth,
		User:          s.User,
		Workspace:     s.Workspace,
		Workflow:      s.Workflow,
		Execution:     s.Execution,
		Credential:    s.Credential,
		Schedule:      s.Schedule,
		Billing:       s.Billing,
		OAuth:         s.OAuth,
		Template:      s.Template,
		WebhookMgr:    s.WebhookMgr,
		WaitResumeMgr: s.WaitResumeMgr,
		AuditLog:      s.AuditLog,
		Alert:         s.Alert,
		Comment:       s.Comment,
		ExecShare:     s.ExecShare,
		EnvVar:        s.EnvVar,
		Analytics:     s.Analytics,
		ExportImport:  s.ExportImport,
	}
}

// ProvideAPIRepositories converts wire.Repositories to api.Repositories
func ProvideAPIRepositories(r *Repositories) *api.Repositories {
	return &api.Repositories{
		PinnedData:      r.PinnedData,
		WaitingExec:     r.WaitingExec,
		WebhookEndpoint: r.WebhookEndpoint,
	}
}

// AppSet combines all provider sets for API server
var AppSet = wire.NewSet(
	InfraSet,
	CryptoSet,
	RepositorySet,
	ServiceSet,
	ProvideServices,
	ProvideRepositories,
	ProvideAPIServices,
	ProvideAPIRepositories,
)

// WorkerSet provides dependencies for worker service
var WorkerSet = wire.NewSet(
	InfraSet,
	CryptoSet,
	// Worker-specific repositories
	repositories.NewWorkflowRepository,
	repositories.NewWorkflowVersionRepository,
	repositories.NewExecutionRepository,
	repositories.NewNodeExecutionRepository,
	repositories.NewCredentialRepository,
	repositories.NewWorkspaceRepository,
	repositories.NewPlanRepository,
	repositories.NewSubscriptionRepository,
	repositories.NewUsageRepository,
	repositories.NewInvoiceRepository,
	// Worker-specific services
	ProvideWorkflowService,
	services.NewExecutionService,
	services.NewCredentialService,
	ProvideWorkerBillingService,
	// Webhook endpoint repo for workflow service
	repositories.NewWebhookEndpointRepository,
)

// ProvideWorkerBillingService creates billing service for worker (no counting repos needed)
func ProvideWorkerBillingService(
	planRepo *repositories.PlanRepository,
	subscriptionRepo *repositories.SubscriptionRepository,
	usageRepo *repositories.UsageRepository,
	invoiceRepo *repositories.InvoiceRepository,
	workspaceRepo *repositories.WorkspaceRepository,
) *services.BillingService {
	return services.NewBillingService(planRepo, subscriptionRepo, usageRepo, invoiceRepo, workspaceRepo)
}

// SchedulerSet provides dependencies for scheduler service
var SchedulerSet = wire.NewSet(
	InfraSet,
)
