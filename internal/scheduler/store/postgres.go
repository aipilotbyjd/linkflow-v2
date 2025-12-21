package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore(db *gorm.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) GetDue(ctx context.Context, limit int) ([]*Schedule, error) {
	var schedules []models.Schedule

	err := s.db.WithContext(ctx).
		Where("is_active = ? AND next_run_at <= ?", true, time.Now()).
		Order("next_run_at ASC").
		Limit(limit).
		Find(&schedules).Error

	if err != nil {
		return nil, err
	}

	return s.toSchedules(schedules), nil
}

func (s *PostgresStore) GetDueByPriority(ctx context.Context, priority string, limit int) ([]*Schedule, error) {
	var schedules []models.Schedule

	err := s.db.WithContext(ctx).
		Where("is_active = ? AND next_run_at <= ? AND priority = ?", true, time.Now(), priority).
		Order("next_run_at ASC").
		Limit(limit).
		Find(&schedules).Error

	if err != nil {
		return nil, err
	}

	return s.toSchedules(schedules), nil
}

func (s *PostgresStore) GetDueByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*Schedule, error) {
	var schedules []models.Schedule

	err := s.db.WithContext(ctx).
		Where("is_active = ? AND next_run_at <= ? AND workspace_id = ?", true, time.Now(), workspaceID).
		Order("next_run_at ASC").
		Limit(limit).
		Find(&schedules).Error

	if err != nil {
		return nil, err
	}

	return s.toSchedules(schedules), nil
}

func (s *PostgresStore) UpdateNextRun(ctx context.Context, id uuid.UUID, nextRun time.Time) error {
	return s.db.WithContext(ctx).
		Model(&models.Schedule{}).
		Where("id = ?", id).
		Update("next_run_at", nextRun).Error
}

func (s *PostgresStore) RecordRun(ctx context.Context, id uuid.UUID, nextRun time.Time) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&models.Schedule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_run_at": now,
			"next_run_at": nextRun,
			"run_count":   gorm.Expr("run_count + 1"),
		}).Error
}

func (s *PostgresStore) GetStale(ctx context.Context, threshold time.Duration) ([]*Schedule, error) {
	cutoff := time.Now().Add(-threshold)

	var schedules []models.Schedule
	err := s.db.WithContext(ctx).
		Where("is_active = ? AND next_run_at < ? AND (last_run_at IS NULL OR last_run_at < next_run_at)", true, cutoff).
		Find(&schedules).Error

	if err != nil {
		return nil, err
	}

	return s.toSchedules(schedules), nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error) {
	var schedule models.Schedule
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return s.toSchedule(&schedule), nil
}

func (s *PostgresStore) toSchedules(models []models.Schedule) []*Schedule {
	result := make([]*Schedule, len(models))
	for i := range models {
		result[i] = s.toSchedule(&models[i])
	}
	return result
}

func (s *PostgresStore) toSchedule(m *models.Schedule) *Schedule {
	sched := &Schedule{
		ID:             m.ID,
		WorkflowID:     m.WorkflowID,
		WorkspaceID:    m.WorkspaceID,
		Name:           m.Name,
		CronExpression: m.CronExpression,
		Timezone:       m.Timezone,
		RunCount:       m.RunCount,
		IsActive:       m.IsActive,
		NextRunAt:      m.NextRunAt,
		LastRunAt:      m.LastRunAt,
	}

	if m.InputData != nil {
		sched.InputData = m.InputData
	}

	return sched
}
