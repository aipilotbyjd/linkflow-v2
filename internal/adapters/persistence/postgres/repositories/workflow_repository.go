package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/mappers"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type WorkflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) Create(ctx context.Context, wf *workflow.Workflow) error {
	model := mappers.WorkflowToModel(wf)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *WorkflowRepository) Update(ctx context.Context, wf *workflow.Workflow) error {
	model := mappers.WorkflowToModel(wf)
	return postgres.GetTx(ctx, r.db).Save(model).Error
}

func (r *WorkflowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.Workflow{}, "id = ?", id).Error
}

func (r *WorkflowRepository) FindByID(ctx context.Context, id uuid.UUID) (*workflow.Workflow, error) {
	var model models.Workflow
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workflow.ErrWorkflowNotFound
		}
		return nil, err
	}
	return mappers.WorkflowToDomain(&model), nil
}

func (r *WorkflowRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *workflow.ListOptions) ([]workflow.Workflow, int64, error) {
	var modelList []models.Workflow
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).Where("workspace_id = ?", workspaceID)

	if opts != nil {
		if opts.Status != nil {
			query = query.Where("status = ?", *opts.Status)
		}
		if opts.Search != "" {
			query = query.Where("name ILIKE ?", "%"+opts.Search+"%")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil && opts.ListOptions != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	workflows := make([]workflow.Workflow, len(modelList))
	for i, m := range modelList {
		workflows[i] = *mappers.WorkflowToDomain(&m)
	}

	return workflows, total, nil
}

func (r *WorkflowRepository) ExistsByName(ctx context.Context, workspaceID uuid.UUID, name string) (bool, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("workspace_id = ? AND name = ?", workspaceID, name).
		Count(&count).Error
	return count > 0, err
}

func (r *WorkflowRepository) IncrementExecutionCount(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("id = ?", id).
		UpdateColumn("execution_count", gorm.Expr("execution_count + 1")).
		UpdateColumn("last_executed_at", gorm.Expr("NOW()")).
		Error
}

func (r *WorkflowRepository) FindByFolderID(ctx context.Context, folderID uuid.UUID, opts *types.ListOptions) ([]workflow.Workflow, int64, error) {
	var modelList []models.Workflow
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).Where("project_id = ?", folderID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	workflows := make([]workflow.Workflow, len(modelList))
	for i, m := range modelList {
		workflows[i] = *mappers.WorkflowToDomain(&m)
	}

	return workflows, total, nil
}

func (r *WorkflowRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status workflow.Status) error {
	return postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *WorkflowRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

func (r *WorkflowRepository) CountActiveByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("workspace_id = ? AND status = ?", workspaceID, workflow.StatusActive).
		Count(&count).Error
	return count, err
}
