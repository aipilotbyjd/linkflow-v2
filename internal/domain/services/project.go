package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

type CreateProjectInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Description *string
	Color       *string
	Icon        *string
	ParentID    *uuid.UUID
}

type UpdateProjectInput struct {
	Name        *string
	Description *string
	Color       *string
	Icon        *string
	ParentID    *uuid.UUID
}

func (s *ProjectService) Create(ctx context.Context, input CreateProjectInput) (*models.Project, error) {
	if input.ParentID != nil {
		var parent models.Project
		if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", *input.ParentID, input.WorkspaceID).First(&parent).Error; err != nil {
			return nil, fmt.Errorf("parent project not found")
		}
	}

	project := &models.Project{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
		Icon:        input.Icon,
		ParentID:    input.ParentID,
	}

	if err := s.db.WithContext(ctx).Create(project).Error; err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) List(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	query := s.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	err := query.Order("name ASC").Find(&projects).Error
	return projects, err
}

func (s *ProjectService) Get(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &project, nil
}

func (s *ProjectService) Update(ctx context.Context, id, workspaceID uuid.UUID, input UpdateProjectInput) (*models.Project, error) {
	project, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if project.WorkspaceID != workspaceID {
		return nil, ErrForbidden
	}

	if input.ParentID != nil {
		if *input.ParentID == id {
			return nil, fmt.Errorf("project cannot be its own parent")
		}
		var parent models.Project
		if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", *input.ParentID, workspaceID).First(&parent).Error; err != nil {
			return nil, fmt.Errorf("parent project not found")
		}
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Color != nil {
		updates["color"] = *input.Color
	}
	if input.Icon != nil {
		updates["icon"] = *input.Icon
	}
	if input.ParentID != nil {
		updates["parent_id"] = *input.ParentID
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(project).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.Get(ctx, id)
}

func (s *ProjectService) Delete(ctx context.Context, id, workspaceID uuid.UUID) error {
	project, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	if project.WorkspaceID != workspaceID {
		return ErrForbidden
	}

	var childCount int64
	s.db.WithContext(ctx).Model(&models.Project{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return fmt.Errorf("cannot delete project with sub-projects")
	}

	var workflowCount int64
	s.db.WithContext(ctx).Model(&models.Workflow{}).Where("project_id = ?", id).Count(&workflowCount)
	if workflowCount > 0 {
		return fmt.Errorf("cannot delete project containing workflows")
	}

	return s.db.WithContext(ctx).Delete(&models.Project{}, "id = ?", id).Error
}

func (s *ProjectService) GetTree(ctx context.Context, workspaceID uuid.UUID) ([]ProjectTreeNode, error) {
	var projects []models.Project
	err := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("name ASC").
		Find(&projects).Error
	if err != nil {
		return nil, err
	}

	return buildProjectTree(projects, nil), nil
}

type ProjectTreeNode struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Color       *string           `json:"color,omitempty"`
	Icon        *string           `json:"icon,omitempty"`
	ParentID    *uuid.UUID        `json:"parent_id,omitempty"`
	Children    []ProjectTreeNode `json:"children,omitempty"`
}

func buildProjectTree(projects []models.Project, parentID *uuid.UUID) []ProjectTreeNode {
	var nodes []ProjectTreeNode

	for _, p := range projects {
		matchesParent := (parentID == nil && p.ParentID == nil) ||
			(parentID != nil && p.ParentID != nil && *parentID == *p.ParentID)

		if matchesParent {
			node := ProjectTreeNode{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Color:       p.Color,
				Icon:        p.Icon,
				ParentID:    p.ParentID,
				Children:    buildProjectTree(projects, &p.ID),
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}
