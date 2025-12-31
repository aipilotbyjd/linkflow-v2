package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type WorkflowShareService struct {
	db *gorm.DB
}

func NewWorkflowShareService(db *gorm.DB) *WorkflowShareService {
	return &WorkflowShareService{db: db}
}

type ShareWorkflowInput struct {
	WorkflowID        uuid.UUID
	SourceWorkspaceID uuid.UUID
	TargetWorkspaceID uuid.UUID
	SharedBy          uuid.UUID
	Permission        string
	ExpiresAt         *time.Time
}

func (s *WorkflowShareService) Share(ctx context.Context, input ShareWorkflowInput) (*models.WorkflowShare, error) {
	if input.SourceWorkspaceID == input.TargetWorkspaceID {
		return nil, fmt.Errorf("cannot share workflow with the same workspace")
	}

	if input.Permission == "" {
		input.Permission = "read"
	}
	if input.Permission != "read" && input.Permission != "execute" && input.Permission != "edit" {
		return nil, fmt.Errorf("invalid permission: must be read, execute, or edit")
	}

	var workflow models.Workflow
	if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", input.WorkflowID, input.SourceWorkspaceID).First(&workflow).Error; err != nil {
		return nil, fmt.Errorf("workflow not found in source workspace")
	}

	var targetWs models.Workspace
	if err := s.db.WithContext(ctx).Where("id = ?", input.TargetWorkspaceID).First(&targetWs).Error; err != nil {
		return nil, fmt.Errorf("target workspace not found")
	}

	var existing models.WorkflowShare
	err := s.db.WithContext(ctx).
		Where("workflow_id = ? AND target_workspace_id = ?", input.WorkflowID, input.TargetWorkspaceID).
		First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("workflow already shared with this workspace")
	}

	share := &models.WorkflowShare{
		WorkflowID:        input.WorkflowID,
		SourceWorkspaceID: input.SourceWorkspaceID,
		TargetWorkspaceID: input.TargetWorkspaceID,
		SharedBy:          input.SharedBy,
		Permission:        input.Permission,
		ExpiresAt:         input.ExpiresAt,
	}

	if err := s.db.WithContext(ctx).Create(share).Error; err != nil {
		return nil, err
	}

	return share, nil
}

func (s *WorkflowShareService) Accept(ctx context.Context, shareID, userID, workspaceID uuid.UUID) (*models.WorkflowShare, error) {
	var share models.WorkflowShare
	if err := s.db.WithContext(ctx).Where("id = ? AND target_workspace_id = ?", shareID, workspaceID).First(&share).Error; err != nil {
		return nil, ErrNotFound
	}

	if share.AcceptedAt != nil {
		return nil, fmt.Errorf("share already accepted")
	}

	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("share invitation has expired")
	}

	now := time.Now()
	share.AcceptedAt = &now
	share.AcceptedBy = &userID

	if err := s.db.WithContext(ctx).Save(&share).Error; err != nil {
		return nil, err
	}

	return &share, nil
}

func (s *WorkflowShareService) Revoke(ctx context.Context, shareID, workspaceID uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND source_workspace_id = ?", shareID, workspaceID).
		Delete(&models.WorkflowShare{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *WorkflowShareService) ListSharedByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]models.WorkflowShare, error) {
	var shares []models.WorkflowShare
	err := s.db.WithContext(ctx).
		Preload("Workflow").
		Preload("TargetWorkspace").
		Where("source_workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

func (s *WorkflowShareService) ListSharedWithWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]models.WorkflowShare, error) {
	var shares []models.WorkflowShare
	err := s.db.WithContext(ctx).
		Preload("Workflow").
		Preload("SourceWorkspace").
		Where("target_workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

func (s *WorkflowShareService) GetPendingShares(ctx context.Context, workspaceID uuid.UUID) ([]models.WorkflowShare, error) {
	var shares []models.WorkflowShare
	err := s.db.WithContext(ctx).
		Preload("Workflow").
		Preload("SourceWorkspace").
		Where("target_workspace_id = ? AND accepted_at IS NULL", workspaceID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

func (s *WorkflowShareService) UpdatePermission(ctx context.Context, shareID, workspaceID uuid.UUID, permission string) error {
	if permission != "read" && permission != "execute" && permission != "edit" {
		return fmt.Errorf("invalid permission: must be read, execute, or edit")
	}

	result := s.db.WithContext(ctx).
		Model(&models.WorkflowShare{}).
		Where("id = ? AND source_workspace_id = ?", shareID, workspaceID).
		Update("permission", permission)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *WorkflowShareService) HasAccess(ctx context.Context, workflowID, workspaceID uuid.UUID, requiredPermission string) (bool, error) {
	var share models.WorkflowShare
	err := s.db.WithContext(ctx).
		Where("workflow_id = ? AND target_workspace_id = ? AND accepted_at IS NOT NULL", workflowID, workspaceID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		First(&share).Error
	if err != nil {
		return false, nil
	}

	permissionLevels := map[string]int{"read": 1, "execute": 2, "edit": 3}
	return permissionLevels[share.Permission] >= permissionLevels[requiredPermission], nil
}
