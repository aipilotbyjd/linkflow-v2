package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type AuditLogService struct {
	repo *repositories.BaseRepository[models.AuditLog]
}

// NewAuditLogService creates a new AuditLogService for tracking user actions.
func NewAuditLogService(repo *repositories.BaseRepository[models.AuditLog]) *AuditLogService {
	return &AuditLogService{repo: repo}
}

type AuditLogInput struct {
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	ResourceName *string
	OldValue     interface{}
	NewValue     interface{}
	Metadata     map[string]interface{}
	IPAddress    *string
	UserAgent    *string
}

func (s *AuditLogService) Log(ctx context.Context, input AuditLogInput) error {
	var oldVal, newVal, meta models.JSON

	if input.OldValue != nil {
		if m, ok := input.OldValue.(map[string]interface{}); ok {
			oldVal = m
		} else if b, err := json.Marshal(input.OldValue); err == nil {
			json.Unmarshal(b, &oldVal)
		}
	}
	if input.NewValue != nil {
		if m, ok := input.NewValue.(map[string]interface{}); ok {
			newVal = m
		} else if b, err := json.Marshal(input.NewValue); err == nil {
			json.Unmarshal(b, &newVal)
		}
	}
	if input.Metadata != nil {
		meta = input.Metadata
	}

	log := &models.AuditLog{
		WorkspaceID:  input.WorkspaceID,
		UserID:       input.UserID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		ResourceName: input.ResourceName,
		OldValue:     oldVal,
		NewValue:     newVal,
		Metadata:     meta,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
	}

	return s.repo.Create(ctx, log)
}

func (s *AuditLogService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.repo.DB().WithContext(ctx).Model(&models.AuditLog{}).Where("workspace_id = ?", workspaceID)
	query.Count(&total)

	if opts != nil {
		if opts.Limit > 0 {
			query = query.Limit(opts.Limit)
		}
		if opts.Offset > 0 {
			query = query.Offset(opts.Offset)
		}
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (s *AuditLogService) Search(ctx context.Context, workspaceID uuid.UUID, action, resourceType string, userID *uuid.UUID, start, end *time.Time, opts *repositories.ListOptions) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.repo.DB().WithContext(ctx).Model(&models.AuditLog{}).Where("workspace_id = ?", workspaceID)

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if start != nil {
		query = query.Where("created_at >= ?", *start)
	}
	if end != nil {
		query = query.Where("created_at <= ?", *end)
	}

	query.Count(&total)

	if opts != nil {
		if opts.Limit > 0 {
			query = query.Limit(opts.Limit)
		}
		if opts.Offset > 0 {
			query = query.Offset(opts.Offset)
		}
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}
