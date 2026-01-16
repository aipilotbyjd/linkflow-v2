package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(ctx context.Context, s *schedule.Schedule) error {
	return postgres.GetTx(ctx, r.db).Create(s).Error
}

func (r *ScheduleRepository) Update(ctx context.Context, s *schedule.Schedule) error {
	return postgres.GetTx(ctx, r.db).Save(s).Error
}

func (r *ScheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&schedule.Schedule{}, "id = ?", id).Error
}

func (r *ScheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*schedule.Schedule, error) {
	var s schedule.Schedule
	if err := postgres.GetTx(ctx, r.db).First(&s, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, schedule.ErrScheduleNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *ScheduleRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]schedule.Schedule, error) {
	var schedules []schedule.Schedule
	if err := postgres.GetTx(ctx, r.db).Where("workflow_id = ?", workflowID).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *ScheduleRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]schedule.Schedule, int64, error) {
	var schedules []schedule.Schedule
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&schedule.Schedule{}).Where("workspace_id = ?", workspaceID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&schedules).Error; err != nil {
		return nil, 0, err
	}

	return schedules, total, nil
}

func (r *ScheduleRepository) FindDue(ctx context.Context, before time.Time) ([]schedule.Schedule, error) {
	var schedules []schedule.Schedule
	if err := postgres.GetTx(ctx, r.db).
		Where("is_active = ? AND next_run_at <= ?", true, before).
		Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *ScheduleRepository) FindActive(ctx context.Context) ([]schedule.Schedule, error) {
	var schedules []schedule.Schedule
	if err := postgres.GetTx(ctx, r.db).Where("is_active = ?", true).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *ScheduleRepository) UpdateNextRunAt(ctx context.Context, id uuid.UUID, nextRunAt time.Time) error {
	return postgres.GetTx(ctx, r.db).Model(&schedule.Schedule{}).
		Where("id = ?", id).
		Update("next_run_at", nextRunAt).Error
}

func (r *ScheduleRepository) RecordRun(ctx context.Context, id uuid.UUID, executionID uuid.UUID, nextRunAt time.Time) error {
	now := time.Now()
	return postgres.GetTx(ctx, r.db).Model(&schedule.Schedule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_run_at":       now,
			"last_execution_id": executionID,
			"next_run_at":       nextRunAt,
			"run_count":         gorm.Expr("run_count + 1"),
		}).Error
}

func (r *ScheduleRepository) SetActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	return postgres.GetTx(ctx, r.db).Model(&schedule.Schedule{}).
		Where("id = ?", id).
		Update("is_active", isActive).Error
}

func (r *ScheduleRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&schedule.Schedule{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

func (r *ScheduleRepository) CountActiveByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&schedule.Schedule{}).
		Where("workspace_id = ? AND is_active = ?", workspaceID, true).
		Count(&count).Error
	return count, err
}

func (r *ScheduleRepository) DeleteByWorkflowID(ctx context.Context, workflowID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&schedule.Schedule{}, "workflow_id = ?", workflowID).Error
}

func (r *ScheduleRepository) FindDueSchedules(ctx context.Context, from, to time.Time) ([]*schedule.Schedule, error) {
	var schedules []*schedule.Schedule
	if err := postgres.GetTx(ctx, r.db).
		Where("is_active = ? AND next_run_at >= ? AND next_run_at <= ?", true, from, to).
		Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}
