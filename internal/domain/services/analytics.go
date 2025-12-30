package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type AnalyticsService struct {
	workspaceRepo *repositories.BaseRepository[models.WorkspaceAnalytics]
	workflowRepo  *repositories.BaseRepository[models.WorkflowAnalytics]
	execRepo      *repositories.ExecutionRepository
}

func NewAnalyticsService(
	workspaceRepo *repositories.BaseRepository[models.WorkspaceAnalytics],
	workflowRepo *repositories.BaseRepository[models.WorkflowAnalytics],
	execRepo *repositories.ExecutionRepository,
) *AnalyticsService {
	return &AnalyticsService{
		workspaceRepo: workspaceRepo,
		workflowRepo:  workflowRepo,
		execRepo:      execRepo,
	}
}

func (s *AnalyticsService) GetWorkspaceAnalytics(ctx context.Context, workspaceID uuid.UUID, start, end time.Time) ([]models.WorkspaceAnalytics, error) {
	var analytics []models.WorkspaceAnalytics
	err := s.workspaceRepo.DB().WithContext(ctx).
		Where("workspace_id = ? AND date >= ? AND date <= ?", workspaceID, start, end).
		Order("date").
		Find(&analytics).Error
	return analytics, err
}

func (s *AnalyticsService) GetWorkflowAnalytics(ctx context.Context, workflowID uuid.UUID, start, end time.Time) ([]models.WorkflowAnalytics, error) {
	var analytics []models.WorkflowAnalytics
	err := s.workflowRepo.DB().WithContext(ctx).
		Where("workflow_id = ? AND date >= ? AND date <= ?", workflowID, start, end).
		Order("date").
		Find(&analytics).Error
	return analytics, err
}

func (s *AnalyticsService) AggregateWorkspaceDaily(ctx context.Context, workspaceID uuid.UUID, date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	stats, err := s.execRepo.GetStats(ctx, workspaceID, startOfDay, endOfDay)
	if err != nil {
		return err
	}

	analytics := &models.WorkspaceAnalytics{
		WorkspaceID:         workspaceID,
		Date:                startOfDay,
		ExecutionsTotal:     getIntStat(stats, "total"),
		ExecutionsSuccess:   getIntStat(stats, "success"),
		ExecutionsFailed:    getIntStat(stats, "failed"),
		ExecutionsCancelled: getIntStat(stats, "cancelled"),
	}

	// Upsert
	return s.workspaceRepo.DB().WithContext(ctx).
		Where("workspace_id = ? AND date = ?", workspaceID, startOfDay).
		Assign(analytics).
		FirstOrCreate(analytics).Error
}

func getIntStat(stats map[string]interface{}, key string) int {
	if v, ok := stats[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}
