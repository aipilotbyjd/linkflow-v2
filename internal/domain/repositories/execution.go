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

func (r *ExecutionRepository) GetHourlyStatsByWorkspace(ctx context.Context, start, end time.Time) (map[uuid.UUID]int64, error) {
	type stat struct {
		WorkspaceID uuid.UUID `gorm:"column:workspace_id"`
		Count       int64     `gorm:"column:count"`
	}
	var stats []stat

	err := r.DB().WithContext(ctx).
		Model(&models.Execution{}).
		Select("workspace_id, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("workspace_id").
		Find(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int64)
	for _, s := range stats {
		result[s.WorkspaceID] = s.Count
	}
	return result, nil
}

// ExecutionFilter contains search criteria for executions
type ExecutionFilter struct {
	WorkspaceID *uuid.UUID
	WorkflowID  *uuid.UUID
	Status      *string
	TriggerType *string
	StartDate   *time.Time
	EndDate     *time.Time
	SearchQuery *string
}

// Search finds executions matching the given filter
func (r *ExecutionRepository) Search(ctx context.Context, filter ExecutionFilter, opts *ListOptions) ([]models.Execution, int64, error) {
	var executions []models.Execution
	var total int64

	query := r.DB().WithContext(ctx).Model(&models.Execution{})

	// Apply filters
	if filter.WorkspaceID != nil {
		query = query.Where("workspace_id = ?", *filter.WorkspaceID)
	}
	if filter.WorkflowID != nil {
		query = query.Where("workflow_id = ?", *filter.WorkflowID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.TriggerType != nil {
		query = query.Where("trigger_type = ?", *filter.TriggerType)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		searchTerm := "%" + *filter.SearchQuery + "%"
		query = query.Where("error_message ILIKE ? OR error_node_id ILIKE ?", searchTerm, searchTerm)
	}

	// Get total count
	query.Count(&total)

	// Apply pagination
	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order("created_at DESC")
	}

	err := query.Find(&executions).Error
	return executions, total, err
}

// DeleteByIDs deletes executions by their IDs within a workspace
func (r *ExecutionRepository) DeleteByIDs(ctx context.Context, workspaceID uuid.UUID, ids []uuid.UUID) (int64, error) {
	result := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND id IN ?", workspaceID, ids).
		Delete(&models.Execution{})
	return result.RowsAffected, result.Error
}

// GetStats returns execution statistics for a workspace
func (r *ExecutionRepository) GetStats(ctx context.Context, workspaceID uuid.UUID, start, end time.Time) (map[string]interface{}, error) {
	var result []struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}

	err := r.DB().WithContext(ctx).
		Model(&models.Execution{}).
		Select("status, COUNT(*) as count").
		Where("workspace_id = ? AND created_at BETWEEN ? AND ?", workspaceID, start, end).
		Group("status").
		Find(&result).Error

	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":      int64(0),
		"completed":  int64(0),
		"failed":     int64(0),
		"running":    int64(0),
		"queued":     int64(0),
		"cancelled":  int64(0),
		"timeout":    int64(0),
		"by_status":  map[string]int64{},
		"start_time": start.Format(time.RFC3339),
		"end_time":   end.Format(time.RFC3339),
	}

	byStatus := make(map[string]int64)
	var total int64
	for _, r := range result {
		byStatus[r.Status] = r.Count
		total += r.Count
		stats[r.Status] = r.Count
	}
	stats["total"] = total
	stats["by_status"] = byStatus

	// Get average duration for completed executions
	var avgDuration struct {
		AvgDuration float64 `gorm:"column:avg_duration"`
	}
	_ = r.DB().WithContext(ctx).
		Model(&models.Execution{}).
		Select("AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_duration").
		Where("workspace_id = ? AND status = 'completed' AND completed_at IS NOT NULL AND started_at IS NOT NULL AND created_at BETWEEN ? AND ?", workspaceID, start, end).
		Find(&avgDuration).Error

	stats["avg_duration_seconds"] = avgDuration.AvgDuration

	return stats, nil
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
