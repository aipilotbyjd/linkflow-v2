package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

var ErrCommentNotFound = errors.New("comment not found")

type WorkflowCommentService struct {
	repo *repositories.BaseRepository[models.WorkflowComment]
}

// NewWorkflowCommentService creates a new WorkflowCommentService for managing workflow comments.
func NewWorkflowCommentService(repo *repositories.BaseRepository[models.WorkflowComment]) *WorkflowCommentService {
	return &WorkflowCommentService{repo: repo}
}

type CreateCommentInput struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	NodeID      *string
	ParentID    *uuid.UUID
	CreatedBy   uuid.UUID
	Content     string
}

func (s *WorkflowCommentService) Create(ctx context.Context, input CreateCommentInput) (*models.WorkflowComment, error) {
	comment := &models.WorkflowComment{
		WorkflowID:  input.WorkflowID,
		WorkspaceID: input.WorkspaceID,
		NodeID:      input.NodeID,
		ParentID:    input.ParentID,
		CreatedBy:   input.CreatedBy,
		Content:     input.Content,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *WorkflowCommentService) GetByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowComment, error) {
	var comments []models.WorkflowComment
	err := s.repo.DB().WithContext(ctx).
		Where("workflow_id = ? AND parent_id IS NULL", workflowID).
		Preload("Replies").
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}

func (s *WorkflowCommentService) GetByNode(ctx context.Context, workflowID uuid.UUID, nodeID string) ([]models.WorkflowComment, error) {
	var comments []models.WorkflowComment
	err := s.repo.DB().WithContext(ctx).
		Where("workflow_id = ? AND node_id = ? AND parent_id IS NULL", workflowID, nodeID).
		Preload("Replies").
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}

func (s *WorkflowCommentService) Update(ctx context.Context, id uuid.UUID, content string) error {
	return s.repo.DB().WithContext(ctx).Model(&models.WorkflowComment{}).
		Where("id = ?", id).
		Update("content", content).Error
}

func (s *WorkflowCommentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Delete(&models.WorkflowComment{}, "id = ?", id).Error
}

func (s *WorkflowCommentService) Resolve(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID) error {
	now := time.Now()
	return s.repo.DB().WithContext(ctx).Model(&models.WorkflowComment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_resolved": true,
			"resolved_by": resolvedBy,
			"resolved_at": now,
		}).Error
}

func (s *WorkflowCommentService) Unresolve(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Model(&models.WorkflowComment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_resolved": false,
			"resolved_by": nil,
			"resolved_at": nil,
		}).Error
}
