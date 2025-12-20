package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type WorkflowRepository struct {
	*BaseRepository[models.Workflow]
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{
		BaseRepository: NewBaseRepository[models.Workflow](db),
	}
}

func (r *WorkflowRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]models.Workflow, int64, error) {
	var workflows []models.Workflow
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	query.Model(&models.Workflow{}).Count(&total)

	if opts != nil {
		if opts.OrderBy != "" {
			query = query.Order(opts.OrderBy + " " + opts.Order)
		}
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	err := query.Find(&workflows).Error
	return workflows, total, err
}

func (r *WorkflowRepository) FindByStatus(ctx context.Context, workspaceID uuid.UUID, status string, opts *ListOptions) ([]models.Workflow, int64, error) {
	var workflows []models.Workflow
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ? AND status = ?", workspaceID, status)
	query.Model(&models.Workflow{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order(opts.OrderBy + " " + opts.Order)
	}

	err := query.Find(&workflows).Error
	return workflows, total, err
}

func (r *WorkflowRepository) FindByTags(ctx context.Context, workspaceID uuid.UUID, tags []string) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND tags && ?", workspaceID, tags).
		Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepository) UpdateStatus(ctx context.Context, workflowID uuid.UUID, status string) error {
	updates := map[string]interface{}{"status": status}
	if status == models.WorkflowStatusActive {
		now := time.Now()
		updates["activated_at"] = now
	} else if status == models.WorkflowStatusArchived {
		now := time.Now()
		updates["archived_at"] = now
	}

	return r.DB().WithContext(ctx).Model(&models.Workflow{}).
		Where("id = ?", workflowID).
		Updates(updates).Error
}

func (r *WorkflowRepository) IncrementExecutionCount(ctx context.Context, workflowID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.Workflow{}).
		Where("id = ?", workflowID).
		Updates(map[string]interface{}{
			"execution_count":  gorm.Expr("execution_count + 1"),
			"last_executed_at": time.Now(),
		}).Error
}

func (r *WorkflowRepository) IncrementVersion(ctx context.Context, workflowID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.Workflow{}).
		Where("id = ?", workflowID).
		Update("version", gorm.Expr("version + 1")).Error
}

func (r *WorkflowRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.Workflow{}).
		Where("workspace_id = ?", workspaceID).
		Count(&count).Error
	return count, err
}

func (r *WorkflowRepository) Search(ctx context.Context, workspaceID uuid.UUID, query string, opts *ListOptions) ([]models.Workflow, int64, error) {
	var workflows []models.Workflow
	var total int64

	dbQuery := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND (name ILIKE ? OR description ILIKE ?)", workspaceID, "%"+query+"%", "%"+query+"%")
	dbQuery.Model(&models.Workflow{}).Count(&total)

	if opts != nil {
		dbQuery = dbQuery.Offset(opts.Offset).Limit(opts.Limit).Order(opts.OrderBy + " " + opts.Order)
	}

	err := dbQuery.Find(&workflows).Error
	return workflows, total, err
}

func (r *WorkflowRepository) FindActiveWithWebhook(ctx context.Context, endpointID string) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := r.DB().WithContext(ctx).
		Where("status = ? AND settings->>'webhookEndpoint' = ?", models.WorkflowStatusActive, endpointID).
		Find(&workflows).Error
	return workflows, err
}

// Version methods
type WorkflowVersionRepository struct {
	*BaseRepository[models.WorkflowVersion]
}

func NewWorkflowVersionRepository(db *gorm.DB) *WorkflowVersionRepository {
	return &WorkflowVersionRepository{
		BaseRepository: NewBaseRepository[models.WorkflowVersion](db),
	}
}

func (r *WorkflowVersionRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowVersion, error) {
	var versions []models.WorkflowVersion
	err := r.DB().WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("version DESC").
		Find(&versions).Error
	return versions, err
}

func (r *WorkflowVersionRepository) FindByWorkflowAndVersion(ctx context.Context, workflowID uuid.UUID, version int) (*models.WorkflowVersion, error) {
	var v models.WorkflowVersion
	err := r.DB().WithContext(ctx).
		Where("workflow_id = ? AND version = ?", workflowID, version).
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *WorkflowVersionRepository) GetLatestVersion(ctx context.Context, workflowID uuid.UUID) (int, error) {
	var version models.WorkflowVersion
	err := r.DB().WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("version DESC").
		First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return version.Version, nil
}
