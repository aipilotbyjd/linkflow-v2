package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type MarketplaceService struct {
	db *gorm.DB
}

func NewMarketplaceService(db *gorm.DB) *MarketplaceService {
	return &MarketplaceService{db: db}
}

type PublishTemplateInput struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	PublishedBy uuid.UUID
	Name        string
	Description string
	Category    string
	Tags        []string
	Icon        string
	IsPublic    bool
}

func (s *MarketplaceService) Publish(ctx context.Context, input PublishTemplateInput) (*models.TemplateMarketplace, error) {
	var workflow models.Workflow
	if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", input.WorkflowID, input.WorkspaceID).First(&workflow).Error; err != nil {
		return nil, fmt.Errorf("workflow not found")
	}

	var existing models.TemplateMarketplace
	err := s.db.WithContext(ctx).Where("workflow_id = ?", input.WorkflowID).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("workflow already published to marketplace")
	}

	template := &models.TemplateMarketplace{
		WorkflowID:  input.WorkflowID,
		WorkspaceID: input.WorkspaceID,
		PublishedBy: input.PublishedBy,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Tags:        input.Tags,
		Icon:        input.Icon,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		IsPublic:    input.IsPublic,
		Version:     "1.0.0",
		PublishedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(template).Error; err != nil {
		return nil, err
	}

	return template, nil
}

func (s *MarketplaceService) Update(ctx context.Context, templateID, workspaceID uuid.UUID, input PublishTemplateInput) (*models.TemplateMarketplace, error) {
	var template models.TemplateMarketplace
	if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", templateID, workspaceID).First(&template).Error; err != nil {
		return nil, ErrNotFound
	}

	updates := map[string]interface{}{
		"name":        input.Name,
		"description": input.Description,
		"category":    input.Category,
		"tags":        input.Tags,
		"icon":        input.Icon,
		"is_public":   input.IsPublic,
		"updated_at":  time.Now(),
	}

	if err := s.db.WithContext(ctx).Model(&template).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

func (s *MarketplaceService) SyncFromWorkflow(ctx context.Context, templateID, workspaceID uuid.UUID) (*models.TemplateMarketplace, error) {
	var template models.TemplateMarketplace
	if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", templateID, workspaceID).First(&template).Error; err != nil {
		return nil, ErrNotFound
	}

	var workflow models.Workflow
	if err := s.db.WithContext(ctx).Where("id = ?", template.WorkflowID).First(&workflow).Error; err != nil {
		return nil, fmt.Errorf("source workflow not found")
	}

	template.Nodes = workflow.Nodes
	template.Connections = workflow.Connections
	template.Settings = workflow.Settings
	template.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&template).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

func (s *MarketplaceService) Unpublish(ctx context.Context, templateID, workspaceID uuid.UUID) error {
	result := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", templateID, workspaceID).Delete(&models.TemplateMarketplace{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MarketplaceService) Get(ctx context.Context, templateID uuid.UUID) (*models.TemplateMarketplace, error) {
	var template models.TemplateMarketplace
	if err := s.db.WithContext(ctx).Where("id = ?", templateID).First(&template).Error; err != nil {
		return nil, ErrNotFound
	}
	return &template, nil
}

func (s *MarketplaceService) ListPublic(ctx context.Context, category string, limit, offset int) ([]models.TemplateMarketplace, int64, error) {
	var templates []models.TemplateMarketplace
	var total int64

	query := s.db.WithContext(ctx).Model(&models.TemplateMarketplace{}).Where("is_public = ?", true)
	if category != "" {
		query = query.Where("category = ?", category)
	}

	query.Count(&total)

	err := query.Order("usage_count DESC, rating DESC").Limit(limit).Offset(offset).Find(&templates).Error
	return templates, total, err
}

func (s *MarketplaceService) ListFeatured(ctx context.Context, limit int) ([]models.TemplateMarketplace, error) {
	var templates []models.TemplateMarketplace
	err := s.db.WithContext(ctx).
		Where("is_public = ? AND is_featured = ?", true, true).
		Order("rating DESC, usage_count DESC").
		Limit(limit).
		Find(&templates).Error
	return templates, err
}

func (s *MarketplaceService) Search(ctx context.Context, query string, limit, offset int) ([]models.TemplateMarketplace, int64, error) {
	var templates []models.TemplateMarketplace
	var total int64

	searchQuery := s.db.WithContext(ctx).Model(&models.TemplateMarketplace{}).
		Where("is_public = ?", true).
		Where("name ILIKE ? OR description ILIKE ? OR category ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")

	searchQuery.Count(&total)

	err := searchQuery.Order("rating DESC, usage_count DESC").Limit(limit).Offset(offset).Find(&templates).Error
	return templates, total, err
}

func (s *MarketplaceService) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]models.TemplateMarketplace, error) {
	var templates []models.TemplateMarketplace
	err := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&templates).Error
	return templates, err
}

func (s *MarketplaceService) GetCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := s.db.WithContext(ctx).
		Model(&models.TemplateMarketplace{}).
		Where("is_public = ?", true).
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}

func (s *MarketplaceService) IncrementUsage(ctx context.Context, templateID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Model(&models.TemplateMarketplace{}).
		Where("id = ?", templateID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (s *MarketplaceService) SetFeatured(ctx context.Context, templateID uuid.UUID, featured bool) error {
	return s.db.WithContext(ctx).
		Model(&models.TemplateMarketplace{}).
		Where("id = ?", templateID).
		Update("is_featured", featured).Error
}

type UseMarketplaceTemplateInput struct {
	TemplateID   uuid.UUID
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	Name         string
	Variables    models.JSON
}

func (s *MarketplaceService) UseTemplate(ctx context.Context, input UseMarketplaceTemplateInput) (*models.Workflow, error) {
	template, err := s.Get(ctx, input.TemplateID)
	if err != nil {
		return nil, err
	}

	if !template.IsPublic {
		return nil, fmt.Errorf("template is not public")
	}

	workflow := &models.Workflow{
		WorkspaceID: input.WorkspaceID,
		CreatedBy:   input.UserID,
		Name:        input.Name,
		Description: &template.Description,
		Status:      models.WorkflowStatusDraft,
		Version:     1,
		Nodes:       template.Nodes,
		Connections: template.Connections,
		Settings:    template.Settings,
		Tags:        template.Tags,
	}

	if err := s.db.WithContext(ctx).Create(workflow).Error; err != nil {
		return nil, err
	}

	_ = s.IncrementUsage(ctx, template.ID)

	return workflow, nil
}
