package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type ExecutionRepository struct {
	*BaseRepository[models.Execution]
}

func NewExecutionRepository(db *gorm.DB) *ExecutionRepository {
	return &ExecutionRepository{
		BaseRepository: NewBaseRepository[models.Execution](db),
	}
}

func (r *ExecutionRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID, opts *ListOptions) ([]models.Execution, int64, error) {
	var executions []models.Execution
	var total int64

	query := r.DB().WithContext(ctx).Where("workflow_id = ?", workflowID)
	query.Model(&models.Execution{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order("created_at DESC")
	}

	err := query.Find(&executions).Error
	return executions, total, err
}

func (r *ExecutionRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]models.Execution, int64, error) {
	var executions []models.Execution
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	query.Model(&models.Execution{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order("created_at DESC")
	}

	err := query.Find(&executions).Error
	return executions, total, err
}

func (r *ExecutionRepository) FindByStatus(ctx context.Context, status string, opts *ListOptions) ([]models.Execution, int64, error) {
	var executions []models.Execution
	var total int64

	query := r.DB().WithContext(ctx).Where("status = ?", status)
	query.Model(&models.Execution{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order("created_at DESC")
	}

	err := query.Find(&executions).Error
	return executions, total, err
}

func (r *ExecutionRepository) FindRunning(ctx context.Context) ([]models.Execution, error) {
	var executions []models.Execution
	err := r.DB().WithContext(ctx).
		Where("status = ?", models.ExecutionStatusRunning).
		Find(&executions).Error
	return executions, err
}

func (r *ExecutionRepository) FindStale(ctx context.Context, threshold time.Duration) ([]models.Execution, error) {
	var executions []models.Execution
	cutoff := time.Now().Add(-threshold)
	err := r.DB().WithContext(ctx).
		Where("status = ? AND started_at < ?", models.ExecutionStatusRunning, cutoff).
		Find(&executions).Error
	return executions, err
}

func (r *ExecutionRepository) UpdateStatus(ctx context.Context, executionID uuid.UUID, status string) error {
	updates := map[string]interface{}{"status": status}
	
	if status == models.ExecutionStatusRunning {
		now := time.Now()
		updates["started_at"] = now
	} else if status == models.ExecutionStatusCompleted || status == models.ExecutionStatusFailed || status == models.ExecutionStatusCancelled || status == models.ExecutionStatusTimeout {
		now := time.Now()
		updates["completed_at"] = now
	}

	return r.DB().WithContext(ctx).Model(&models.Execution{}).
		Where("id = ?", executionID).
		Updates(updates).Error
}

func (r *ExecutionRepository) SetError(ctx context.Context, executionID uuid.UUID, errorMessage string, errorNodeID *string) error {
	updates := map[string]interface{}{
		"status":        models.ExecutionStatusFailed,
		"error_message": errorMessage,
		"completed_at":  time.Now(),
	}
	if errorNodeID != nil {
		updates["error_node_id"] = *errorNodeID
	}

	return r.DB().WithContext(ctx).Model(&models.Execution{}).
		Where("id = ?", executionID).
		Updates(updates).Error
}

func (r *ExecutionRepository) UpdateProgress(ctx context.Context, executionID uuid.UUID, nodesCompleted int) error {
	return r.DB().WithContext(ctx).Model(&models.Execution{}).
		Where("id = ?", executionID).
		Update("nodes_completed", nodesCompleted).Error
}

func (r *ExecutionRepository) SetOutput(ctx context.Context, executionID uuid.UUID, output models.JSON) error {
	return r.DB().WithContext(ctx).Model(&models.Execution{}).
		Where("id = ?", executionID).
		Update("output_data", output).Error
}

func (r *ExecutionRepository) CountByWorkspaceInPeriod(ctx context.Context, workspaceID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.Execution{}).
		Where("workspace_id = ? AND created_at BETWEEN ? AND ?", workspaceID, start, end).
		Count(&count).Error
	return count, err
}

func (r *ExecutionRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.DB().WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&models.Execution{})
	return result.RowsAffected, result.Error
}

// Node Execution methods
type NodeExecutionRepository struct {
	*BaseRepository[models.NodeExecution]
}

func NewNodeExecutionRepository(db *gorm.DB) *NodeExecutionRepository {
	return &NodeExecutionRepository{
		BaseRepository: NewBaseRepository[models.NodeExecution](db),
	}
}

func (r *NodeExecutionRepository) FindByExecutionID(ctx context.Context, executionID uuid.UUID) ([]models.NodeExecution, error) {
	var nodeExecutions []models.NodeExecution
	err := r.DB().WithContext(ctx).
		Where("execution_id = ?", executionID).
		Order("created_at ASC").
		Find(&nodeExecutions).Error
	return nodeExecutions, err
}

func (r *NodeExecutionRepository) FindByExecutionAndNode(ctx context.Context, executionID uuid.UUID, nodeID string) (*models.NodeExecution, error) {
	var nodeExecution models.NodeExecution
	err := r.DB().WithContext(ctx).
		Where("execution_id = ? AND node_id = ?", executionID, nodeID).
		First(&nodeExecution).Error
	if err != nil {
		return nil, err
	}
	return &nodeExecution, nil
}

func (r *NodeExecutionRepository) UpdateStatus(ctx context.Context, nodeExecutionID uuid.UUID, status string) error {
	updates := map[string]interface{}{"status": status}
	
	if status == models.NodeStatusRunning {
		now := time.Now()
		updates["started_at"] = now
	} else if status == models.NodeStatusCompleted || status == models.NodeStatusFailed || status == models.NodeStatusSkipped {
		now := time.Now()
		updates["completed_at"] = now
	}

	return r.DB().WithContext(ctx).Model(&models.NodeExecution{}).
		Where("id = ?", nodeExecutionID).
		Updates(updates).Error
}

func (r *NodeExecutionRepository) SetResult(ctx context.Context, nodeExecutionID uuid.UUID, status string, output models.JSON, durationMs int) error {
	now := time.Now()
	return r.DB().WithContext(ctx).Model(&models.NodeExecution{}).
		Where("id = ?", nodeExecutionID).
		Updates(map[string]interface{}{
			"status":       status,
			"output_data":  output,
			"completed_at": now,
			"duration_ms":  durationMs,
		}).Error
}

func (r *NodeExecutionRepository) SetError(ctx context.Context, nodeExecutionID uuid.UUID, errorMessage string) error {
	now := time.Now()
	return r.DB().WithContext(ctx).Model(&models.NodeExecution{}).
		Where("id = ?", nodeExecutionID).
		Updates(map[string]interface{}{
			"status":        models.NodeStatusFailed,
			"error_message": errorMessage,
			"completed_at":  now,
		}).Error
}
