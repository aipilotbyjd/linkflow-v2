package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type ExecutionLogRepository struct {
	db *gorm.DB
}

func NewExecutionLogRepository(db *gorm.DB) *ExecutionLogRepository {
	return &ExecutionLogRepository{db: db}
}

func (r *ExecutionLogRepository) Create(ctx context.Context, log *execution.Log) error {
	return postgres.GetTx(ctx, r.db).Create(log).Error
}

func (r *ExecutionLogRepository) CreateBatch(ctx context.Context, logs []execution.Log) error {
	if len(logs) == 0 {
		return nil
	}
	return postgres.GetTx(ctx, r.db).CreateInBatches(logs, 100).Error
}

func (r *ExecutionLogRepository) FindByExecutionID(ctx context.Context, executionID uuid.UUID, opts *types.ListOptions) ([]execution.Log, int64, error) {
	var logs []execution.Log
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&execution.Log{}).
		Where("execution_id = ?", executionID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Order("timestamp ASC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *ExecutionLogRepository) FindByNodeID(ctx context.Context, executionID uuid.UUID, nodeID string) ([]execution.Log, error) {
	var logs []execution.Log
	if err := postgres.GetTx(ctx, r.db).
		Where("execution_id = ? AND node_id = ?", executionID, nodeID).
		Order("timestamp ASC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *ExecutionLogRepository) DeleteByExecutionID(ctx context.Context, executionID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).
		Where("execution_id = ?", executionID).
		Delete(&execution.Log{}).Error
}

func (r *ExecutionLogRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := postgres.GetTx(ctx, r.db).
		Where("timestamp < ?", before).
		Delete(&execution.Log{})
	return result.RowsAffected, result.Error
}
