package wire

import (
	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"gorm.io/gorm"
)

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
	ProvideExecutionReplayService,
	ProvideVersionDiffService,
	services.NewAPIKeyService,
	services.NewWorkflowVariableService,
	services.NewFolderService,
	services.NewWorkflowShareService,
	services.NewMarketplaceService,
	services.NewTemplateRatingService,
	ProvideBinaryDataService,
	ProvideDashboardService,
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

// ProvideOAuthService creates the OAuth service with encryption and URL configuration
func ProvideOAuthService(
	stateRepo *repositories.OAuthStateRepository,
	credentialRepo *repositories.CredentialRepository,
	encryptor *crypto.Encryptor,
	cfg *config.Config,
	baseURL BaseURL,
	frontendURL FrontendURL,
) *services.OAuthService {
	svc := services.NewOAuthService(stateRepo, credentialRepo, encryptor, string(baseURL), string(frontendURL))

	// Configure OAuth providers from application config
	if cfg.HasOAuth("google") {
		svc.ConfigureProvider("google", cfg.OAuth.Google.ClientID, cfg.OAuth.Google.ClientSecret)
	}
	if cfg.HasOAuth("github") {
		svc.ConfigureProvider("github", cfg.OAuth.GitHub.ClientID, cfg.OAuth.GitHub.ClientSecret)
	}
	if cfg.HasOAuth("microsoft") {
		svc.ConfigureProvider("microsoft", cfg.OAuth.Microsoft.ClientID, cfg.OAuth.Microsoft.ClientSecret)
	}
	if cfg.HasOAuth("slack") {
		svc.ConfigureProvider("slack", cfg.OAuth.Slack.ClientID, cfg.OAuth.Slack.ClientSecret)
	}
	if cfg.HasOAuth("notion") {
		svc.ConfigureProvider("notion", cfg.OAuth.Notion.ClientID, cfg.OAuth.Notion.ClientSecret)
	}
	if cfg.HasOAuth("hubspot") {
		svc.ConfigureProvider("hubspot", cfg.OAuth.HubSpot.ClientID, cfg.OAuth.HubSpot.ClientSecret)
	}
	if cfg.HasOAuth("salesforce") {
		svc.ConfigureProvider("salesforce", cfg.OAuth.Salesforce.ClientID, cfg.OAuth.Salesforce.ClientSecret)
	}
	if cfg.HasOAuth("airtable") {
		svc.ConfigureProvider("airtable", cfg.OAuth.Airtable.ClientID, cfg.OAuth.Airtable.ClientSecret)
	}
	if cfg.HasOAuth("stripe") {
		svc.ConfigureProvider("stripe", cfg.OAuth.Stripe.ClientID, cfg.OAuth.Stripe.ClientSecret)
	}

	return svc
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

// ProvideExecutionReplayService creates the execution replay service
func ProvideExecutionReplayService(
	execService *services.ExecutionService,
	execRepo *repositories.ExecutionRepository,
) *services.ExecutionReplayService {
	return services.NewExecutionReplayService(execService, execRepo)
}

// ProvideVersionDiffService creates the version diff service
func ProvideVersionDiffService(
	workflowRepo *repositories.WorkflowRepository,
	versionRepo *repositories.WorkflowVersionRepository,
) *services.VersionDiffService {
	return services.NewVersionDiffService(workflowRepo, versionRepo)
}

// WorkerServiceSet provides services needed by worker
var WorkerServiceSet = wire.NewSet(
	ProvideWorkflowService,
	services.NewExecutionService,
	services.NewCredentialService,
	ProvideWorkerBillingService,
	ProvideOAuthService,
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

// ProvideBinaryDataService creates the binary data service with storage path
func ProvideBinaryDataService(db *gorm.DB) *services.BinaryDataService {
	return services.NewBinaryDataService(db, "")
}

// ProvideDashboardService creates the dashboard service
func ProvideDashboardService(
	db *gorm.DB,
	workflowRepo *repositories.WorkflowRepository,
	executionRepo *repositories.ExecutionRepository,
	credentialRepo *repositories.CredentialRepository,
	scheduleRepo *repositories.ScheduleRepository,
) *services.DashboardService {
	return services.NewDashboardService(db, workflowRepo, executionRepo, credentialRepo, scheduleRepo)
}
