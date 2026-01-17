package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"gorm.io/gorm"
)

type ExecutionStatsRepository struct {
	db *gorm.DB
}

func NewExecutionStatsRepository(db *gorm.DB) *ExecutionStatsRepository {
	return &ExecutionStatsRepository{db: db}
}

func (r *ExecutionStatsRepository) GetStats(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (*execution.Stats, error) {
	var stats execution.Stats

	query := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workspace_id = ? AND created_at BETWEEN ? AND ?", workspaceID, from, to)

	if err := query.Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	if err := query.Where("status = ?", execution.StatusCompleted).Count(&stats.Completed).Error; err != nil {
		return nil, err
	}

	if err := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workspace_id = ? AND created_at BETWEEN ? AND ? AND status = ?", workspaceID, from, to, execution.StatusFailed).
		Count(&stats.Failed).Error; err != nil {
		return nil, err
	}

	if err := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workspace_id = ? AND created_at BETWEEN ? AND ? AND status = ?", workspaceID, from, to, execution.StatusCancelled).
		Count(&stats.Cancelled).Error; err != nil {
		return nil, err
	}

	if err := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workspace_id = ? AND created_at BETWEEN ? AND ? AND status = ?", workspaceID, from, to, execution.StatusRunning).
		Count(&stats.Running).Error; err != nil {
		return nil, err
	}

	if err := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workspace_id = ? AND created_at BETWEEN ? AND ? AND status = ?", workspaceID, from, to, execution.StatusQueued).
		Count(&stats.Queued).Error; err != nil {
		return nil, err
	}

	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Completed) / float64(stats.Total) * 100
	}

	var avgDuration float64
	postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workspace_id = ? AND created_at BETWEEN ? AND ? AND status = ? AND completed_at IS NOT NULL", workspaceID, from, to, execution.StatusCompleted).
		Select("AVG(EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000)").
		Scan(&avgDuration)
	stats.AvgDuration = time.Duration(avgDuration) * time.Millisecond

	return &stats, nil
}

func (r *ExecutionStatsRepository) GetWorkflowStats(ctx context.Context, workflowID uuid.UUID, from, to time.Time) (*execution.Stats, error) {
	var stats execution.Stats

	query := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workflow_id = ? AND created_at BETWEEN ? AND ?", workflowID, from, to)

	if err := query.Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	if err := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workflow_id = ? AND created_at BETWEEN ? AND ? AND status = ?", workflowID, from, to, execution.StatusCompleted).
		Count(&stats.Completed).Error; err != nil {
		return nil, err
	}

	if err := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Where("workflow_id = ? AND created_at BETWEEN ? AND ? AND status = ?", workflowID, from, to, execution.StatusFailed).
		Count(&stats.Failed).Error; err != nil {
		return nil, err
	}

	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Completed) / float64(stats.Total) * 100
	}

	return &stats, nil
}

func (r *ExecutionStatsRepository) GetDailyStats(ctx context.Context, workspaceID uuid.UUID, days int) ([]execution.DailyStat, error) {
	var results []execution.DailyStat

	startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)

	rows, err := postgres.GetTx(ctx, r.db).Model(&execution.Execution{}).
		Select("DATE(created_at) as date, COUNT(*) as total, "+
			"SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as completed, "+
			"SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as failed",
			execution.StatusCompleted, execution.StatusFailed).
		Where("workspace_id = ? AND created_at >= ?", workspaceID, startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat execution.DailyStat
		if err := rows.Scan(&stat.Date, &stat.Total, &stat.Completed, &stat.Failed); err != nil {
			return nil, err
		}
		results = append(results, stat)
	}

	return results, nil
}
