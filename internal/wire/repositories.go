package wire

import (
	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"gorm.io/gorm"
)

// RepositorySet provides all repository dependencies
var RepositorySet = wire.NewSet(
	// Core repositories
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
	repositories.NewAPIKeyRepository,
	// Generic repositories (for newer features)
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

// Generic repository providers for BaseRepository[T]

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

// WorkerRepositorySet provides repositories needed by worker service
var WorkerRepositorySet = wire.NewSet(
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
	repositories.NewWebhookEndpointRepository,
)
