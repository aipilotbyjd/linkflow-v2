package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type WorkflowFolderService struct {
	db *gorm.DB
}

func NewWorkflowFolderService(db *gorm.DB) *WorkflowFolderService {
	return &WorkflowFolderService{db: db}
}

type CreateFolderInput struct {
	WorkspaceID uuid.UUID
	Name        string
	ParentID    *uuid.UUID
}

type UpdateFolderInput struct {
	Name     *string
	ParentID *uuid.UUID
}

func (s *WorkflowFolderService) Create(ctx context.Context, input CreateFolderInput) (*models.WorkflowFolder, error) {
	if input.ParentID != nil {
		var parent models.WorkflowFolder
		if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", *input.ParentID, input.WorkspaceID).First(&parent).Error; err != nil {
			return nil, fmt.Errorf("parent folder not found")
		}
	}

	folder := &models.WorkflowFolder{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		ParentID:    input.ParentID,
	}

	if err := s.db.WithContext(ctx).Create(folder).Error; err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *WorkflowFolderService) List(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID) ([]models.WorkflowFolder, error) {
	var folders []models.WorkflowFolder
	query := s.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	err := query.Order("name ASC").Find(&folders).Error
	return folders, err
}

func (s *WorkflowFolderService) Get(ctx context.Context, id uuid.UUID) (*models.WorkflowFolder, error) {
	var folder models.WorkflowFolder
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&folder).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &folder, nil
}

func (s *WorkflowFolderService) Update(ctx context.Context, id, workspaceID uuid.UUID, input UpdateFolderInput) (*models.WorkflowFolder, error) {
	folder, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if folder.WorkspaceID != workspaceID {
		return nil, ErrForbidden
	}

	if input.ParentID != nil {
		if *input.ParentID == id {
			return nil, fmt.Errorf("folder cannot be its own parent")
		}
		var parent models.WorkflowFolder
		if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", *input.ParentID, workspaceID).First(&parent).Error; err != nil {
			return nil, fmt.Errorf("parent folder not found")
		}
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.ParentID != nil {
		updates["parent_id"] = *input.ParentID
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(folder).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.Get(ctx, id)
}

func (s *WorkflowFolderService) Delete(ctx context.Context, id, workspaceID uuid.UUID) error {
	folder, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	if folder.WorkspaceID != workspaceID {
		return ErrForbidden
	}

	var childCount int64
	s.db.WithContext(ctx).Model(&models.WorkflowFolder{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return fmt.Errorf("cannot delete folder with subfolders")
	}

	var workflowCount int64
	s.db.WithContext(ctx).Model(&models.Workflow{}).Where("folder_id = ?", id).Count(&workflowCount)
	if workflowCount > 0 {
		return fmt.Errorf("cannot delete folder containing workflows")
	}

	return s.db.WithContext(ctx).Delete(&models.WorkflowFolder{}, "id = ?", id).Error
}

func (s *WorkflowFolderService) GetTree(ctx context.Context, workspaceID uuid.UUID) ([]FolderTreeNode, error) {
	var folders []models.WorkflowFolder
	err := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("name ASC").
		Find(&folders).Error
	if err != nil {
		return nil, err
	}

	return buildFolderTree(folders, nil), nil
}

type FolderTreeNode struct {
	ID       uuid.UUID        `json:"id"`
	Name     string           `json:"name"`
	ParentID *uuid.UUID       `json:"parent_id,omitempty"`
	Children []FolderTreeNode `json:"children,omitempty"`
}

func buildFolderTree(folders []models.WorkflowFolder, parentID *uuid.UUID) []FolderTreeNode {
	var nodes []FolderTreeNode

	for _, f := range folders {
		matchesParent := (parentID == nil && f.ParentID == nil) ||
			(parentID != nil && f.ParentID != nil && *parentID == *f.ParentID)

		if matchesParent {
			node := FolderTreeNode{
				ID:       f.ID,
				Name:     f.Name,
				ParentID: f.ParentID,
				Children: buildFolderTree(folders, &f.ID),
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}
