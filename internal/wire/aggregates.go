package wire

import (
	"github.com/linkflow-ai/linkflow/internal/api"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

// ProvideServices aggregates all services directly into api.Services
// This eliminates the need for intermediate wire.Services struct
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
	execReplay *services.ExecutionReplayService,
	versionDiff *services.VersionDiffService,
	apiKey *services.APIKeyService,
	workflowVar *services.WorkflowVariableService,
	folder *services.FolderService,
	workflowShare *services.WorkflowShareService,
	marketplace *services.MarketplaceService,
	templateRating *services.TemplateRatingService,
	binaryData *services.BinaryDataService,
	dashboard *services.DashboardService,
	note *services.NoteService,
) *api.Services {
	return &api.Services{
		Auth:           auth,
		User:           user,
		Workspace:      workspace,
		Workflow:       workflow,
		Execution:      execution,
		Credential:     credential,
		Schedule:       schedule,
		Billing:        billing,
		OAuth:          oauth,
		Template:       template,
		WebhookMgr:     webhookMgr,
		WaitResumeMgr:  waitResumeMgr,
		AuditLog:       auditLog,
		Alert:          alert,
		Comment:        comment,
		ExecShare:      execShare,
		EnvVar:         envVar,
		Analytics:      analytics,
		ExportImport:   exportImport,
		ExecReplay:     execReplay,
		VersionDiff:    versionDiff,
		APIKey:         apiKey,
		WorkflowVar:    workflowVar,
		Folder:         folder,
		WorkflowShare:  workflowShare,
		Marketplace:    marketplace,
		TemplateRating: templateRating,
		BinaryData:     binaryData,
		Dashboard:      dashboard,
		Note:           note,
	}
}

// ProvideRepositories aggregates repositories needed by handlers directly into api.Repositories
func ProvideRepositories(
	pinnedData *repositories.PinnedDataRepository,
	waitingExec *repositories.WaitingExecutionRepository,
	webhookEndpoint *repositories.WebhookEndpointRepository,
) *api.Repositories {
	return &api.Repositories{
		PinnedData:      pinnedData,
		WaitingExec:     waitingExec,
		WebhookEndpoint: webhookEndpoint,
	}
}
