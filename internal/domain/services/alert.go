package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

var ErrAlertNotFound = errors.New("alert not found")

type AlertService struct {
	repo    *repositories.BaseRepository[models.Alert]
	logRepo *repositories.BaseRepository[models.AlertLog]
}

func NewAlertService(repo *repositories.BaseRepository[models.Alert], logRepo *repositories.BaseRepository[models.AlertLog]) *AlertService {
	return &AlertService{repo: repo, logRepo: logRepo}
}

type CreateAlertInput struct {
	WorkspaceID  uuid.UUID
	WorkflowID   *uuid.UUID
	CreatedBy    uuid.UUID
	Name         string
	Type         string
	Trigger      string
	Config       map[string]interface{}
	Conditions   map[string]interface{}
	CooldownMins int
}

func (s *AlertService) Create(ctx context.Context, input CreateAlertInput) (*models.Alert, error) {
	alert := &models.Alert{
		WorkspaceID:  input.WorkspaceID,
		WorkflowID:   input.WorkflowID,
		CreatedBy:    input.CreatedBy,
		Name:         input.Name,
		Type:         input.Type,
		Trigger:      input.Trigger,
		Config:       input.Config,
		Conditions:   input.Conditions,
		CooldownMins: input.CooldownMins,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, alert); err != nil {
		return nil, err
	}
	return alert, nil
}

func (s *AlertService) GetByID(ctx context.Context, id uuid.UUID) (*models.Alert, error) {
	alert, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrAlertNotFound
	}
	return alert, nil
}

func (s *AlertService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]models.Alert, error) {
	var alerts []models.Alert
	err := s.repo.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&alerts).Error
	return alerts, err
}

func (s *AlertService) GetByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Alert, error) {
	var alerts []models.Alert
	err := s.repo.DB().WithContext(ctx).Where("workflow_id = ?", workflowID).Find(&alerts).Error
	return alerts, err
}

func (s *AlertService) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return s.repo.DB().WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AlertService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Delete(&models.Alert{}, "id = ?", id).Error
}

func (s *AlertService) ShouldFire(ctx context.Context, alert *models.Alert) bool {
	if !alert.IsActive {
		return false
	}
	if alert.LastFiredAt != nil && alert.CooldownMins > 0 {
		cooldown := time.Duration(alert.CooldownMins) * time.Minute
		if time.Since(*alert.LastFiredAt) < cooldown {
			return false
		}
	}
	return true
}

func (s *AlertService) Fire(ctx context.Context, alertID uuid.UUID, executionID *uuid.UUID, message string) error {
	now := time.Now()
	if err := s.repo.DB().WithContext(ctx).Model(&models.Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"last_fired_at": now,
		}).Error; err != nil {
		return err
	}
	// Increment fire count separately
	s.repo.DB().WithContext(ctx).Model(&models.Alert{}).
		Where("id = ?", alertID).
		UpdateColumn("fire_count", s.repo.DB().Raw("fire_count + 1"))

	log := &models.AlertLog{
		AlertID:     alertID,
		ExecutionID: executionID,
		Status:      "sent",
		Message:     message,
	}
	return s.logRepo.Create(ctx, log)
}
