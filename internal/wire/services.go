package wire

import (
	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
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
