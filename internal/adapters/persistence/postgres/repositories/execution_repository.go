package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/mappers"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"gorm.io/gorm"
)

type ExecutionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) *ExecutionRepository {
	return &ExecutionRepository{db: db}
}

func (r *ExecutionRepository) Create(ctx context.Context, exec *execution.Execution) error {
	model := mappers.ExecutionToModel(exec)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *ExecutionRepository) Update(ctx context.Context, exec *execution.Execution) error {
	model := mappers.ExecutionToModel(exec)
	return postgres.GetTx(ctx, r.db).Save(model).Error
}

func (r *ExecutionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.Execution{}, "id = ?", id).Error
}

func (r *ExecutionRepository) FindByID(ctx context.Context, id uuid.UUID) (*execution.Execution, error) {
	var model models.Execution
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, execution.ErrExecutionNotFound
		}
		return nil, err
	}
	return mappers.ExecutionToDomain(&model), nil
}

func (r *ExecutionRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID, opts *execution.ListOptions) ([]execution.Execution, int64, error) {
	var modelList []models.Execution
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Execution{}).Where("workflow_id = ?", workflowID)

	if opts != nil && opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
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

	executions := make([]execution.Execution, len(modelList))
	for i, m := range modelList {
		executions[i] = *mappers.ExecutionToDomain(&m)
	}

	return executions, total, nil
}

func (r *ExecutionRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *execution.ListOptions) ([]execution.Execution, int64, error) {
	var modelList []models.Execution
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Execution{}).Where("workspace_id = ?", workspaceID)

	if opts != nil && opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
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

	executions := make([]execution.Execution, len(modelList))
	for i, m := range modelList {
		executions[i] = *mappers.ExecutionToDomain(&m)
	}

	return executions, total, nil
}

func (r *ExecutionRepository) FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]execution.Execution, error) {
	var modelList []models.Execution
	if err := postgres.GetTx(ctx, r.db).Where("batch_id = ?", batchID).Find(&modelList).Error; err != nil {
		return nil, err
	}

	executions := make([]execution.Execution, len(modelList))
	for i, m := range modelList {
		executions[i] = *mappers.ExecutionToDomain(&m)
	}

	return executions, nil
}

func (r *ExecutionRepository) FindRunning(ctx context.Context) ([]execution.Execution, error) {
	var modelList []models.Execution
	if err := postgres.GetTx(ctx, r.db).Where("status = ?", execution.StatusRunning).Find(&modelList).Error; err != nil {
		return nil, err
	}

	executions := make([]execution.Execution, len(modelList))
	for i, m := range modelList {
		executions[i] = *mappers.ExecutionToDomain(&m)
	}

	return executions, nil
}

func (r *ExecutionRepository) FindStale(ctx context.Context, timeout time.Duration) ([]execution.Execution, error) {
	var modelList []models.Execution
	staleTime := time.Now().Add(-timeout)
	if err := postgres.GetTx(ctx, r.db).
		Where("status = ? AND started_at < ?", execution.StatusRunning, staleTime).
		Find(&modelList).Error; err != nil {
		return nil, err
	}

	executions := make([]execution.Execution, len(modelList))
	for i, m := range modelList {
		executions[i] = *mappers.ExecutionToDomain(&m)
	}

	return executions, nil
}

func (r *ExecutionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status execution.Status) error {
	return postgres.GetTx(ctx, r.db).Model(&models.Execution{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *ExecutionRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Execution{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

func (r *ExecutionRepository) CountByWorkflowID(ctx context.Context, workflowID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Execution{}).Where("workflow_id = ?", workflowID).Count(&count).Error
	return count, err
}

func (r *ExecutionRepository) CountByStatus(ctx context.Context, workspaceID uuid.UUID, status execution.Status) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Execution{}).
		Where("workspace_id = ? AND status = ?", workspaceID, status).
		Count(&count).Error
	return count, err
}

func (r *ExecutionRepository) DeleteOlderThan(ctx context.Context, workspaceID uuid.UUID, before time.Time) (int64, error) {
	result := postgres.GetTx(ctx, r.db).
		Where("workspace_id = ? AND created_at < ?", workspaceID, before).
		Delete(&models.Execution{})
	return result.RowsAffected, result.Error
}

func (r *ExecutionRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) (int64, error) {
	result := postgres.GetTx(ctx, r.db).Delete(&models.Execution{}, "id IN ?", ids)
	return result.RowsAffected, result.Error
}
