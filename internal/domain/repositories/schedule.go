package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type ScheduleRepository struct {
	*BaseRepository[models.Schedule]
}

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository {
	return &ScheduleRepository{
		BaseRepository: NewBaseRepository[models.Schedule](db),
	}
}

func (r *ScheduleRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.DB().WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("created_at DESC").
		Find(&schedules).Error
	return schedules, err
}

func (r *ScheduleRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]models.Schedule, int64, error) {
	var schedules []models.Schedule
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	query.Model(&models.Schedule{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order(opts.OrderBy + " " + opts.Order)
	}

	err := query.Find(&schedules).Error
	return schedules, total, err
}

func (r *ScheduleRepository) FindDue(ctx context.Context) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.DB().WithContext(ctx).
		Preload("Workflow").
		Where("is_active = ? AND next_run_at <= ?", true, time.Now()).
		Find(&schedules).Error
	return schedules, err
}

func (r *ScheduleRepository) FindActive(ctx context.Context) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.DB().WithContext(ctx).
		Where("is_active = ?", true).
		Find(&schedules).Error
	return schedules, err
}

func (r *ScheduleRepository) UpdateNextRun(ctx context.Context, scheduleID uuid.UUID, nextRunAt time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Schedule{}).
		Where("id = ?", scheduleID).
		Update("next_run_at", nextRunAt).Error
}

func (r *ScheduleRepository) RecordRun(ctx context.Context, scheduleID uuid.UUID, executionID uuid.UUID, nextRunAt time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Schedule{}).
		Where("id = ?", scheduleID).
		Updates(map[string]interface{}{
			"last_run_at":       time.Now(),
			"last_execution_id": executionID,
			"next_run_at":       nextRunAt,
			"run_count":         gorm.Expr("run_count + 1"),
		}).Error
}

func (r *ScheduleRepository) SetActive(ctx context.Context, scheduleID uuid.UUID, isActive bool) error {
	return r.DB().WithContext(ctx).Model(&models.Schedule{}).
		Where("id = ?", scheduleID).
		Update("is_active", isActive).Error
}

func (r *ScheduleRepository) DeactivateByWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.Schedule{}).
		Where("workflow_id = ?", workflowID).
		Update("is_active", false).Error
}

func (r *ScheduleRepository) FindDueBatch(ctx context.Context, limit, offset int) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.DB().WithContext(ctx).
		Preload("Workflow").
		Where("is_active = ? AND next_run_at <= ?", true, time.Now()).
		Order("next_run_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&schedules).Error
	return schedules, err
}

func (r *ScheduleRepository) FindDueByPriority(ctx context.Context, priority string) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.DB().WithContext(ctx).
		Preload("Workflow").
		Where("is_active = ? AND next_run_at <= ? AND priority = ?", true, time.Now(), priority).
		Order("next_run_at ASC").
		Find(&schedules).Error
	return schedules, err
}
