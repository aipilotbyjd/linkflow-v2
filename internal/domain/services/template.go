package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type TemplateService struct {
	templateRepo *repositories.TemplateRepository
	workflowRepo *repositories.WorkflowRepository
}

func NewTemplateService(
	templateRepo *repositories.TemplateRepository,
	workflowRepo *repositories.WorkflowRepository,
) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		workflowRepo: workflowRepo,
	}
}

type CreateTemplateInput struct {
	Name        string
	Description *string
	Category    string
	Tags        []string
	Nodes       models.JSONArray
	Connections models.JSONArray
	Settings    models.JSON
	Variables   models.JSON
	IconURL     *string
	IsPublic    bool
	CreatedBy   uuid.UUID
}

func (s *TemplateService) Create(ctx context.Context, input CreateTemplateInput) (*models.Template, error) {
	template := &models.Template{
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Tags:        input.Tags,
		Nodes:       input.Nodes,
		Connections: input.Connections,
		Settings:    input.Settings,
		Variables:   input.Variables,
		IconURL:     input.IconURL,
		IsPublic:    input.IsPublic,
		CreatedBy:   &input.CreatedBy,
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *TemplateService) GetByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	return s.templateRepo.FindByID(ctx, id)
}

func (s *TemplateService) GetPublicTemplates(ctx context.Context, opts *repositories.ListOptions) ([]models.Template, int64, error) {
	return s.templateRepo.FindPublic(ctx, opts)
}

func (s *TemplateService) GetByCategory(ctx context.Context, category string, opts *repositories.ListOptions) ([]models.Template, int64, error) {
	return s.templateRepo.FindByCategory(ctx, category, opts)
}

func (s *TemplateService) GetFeatured(ctx context.Context, limit int) ([]models.Template, error) {
	return s.templateRepo.FindFeatured(ctx, limit)
}

func (s *TemplateService) Search(ctx context.Context, query string, opts *repositories.ListOptions) ([]models.Template, int64, error) {
	return s.templateRepo.Search(ctx, query, opts)
}

type CreateFromTemplateInput struct {
	TemplateID  uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Name        string
	Variables   models.JSON // Variable values to substitute
}

func (s *TemplateService) CreateWorkflowFromTemplate(ctx context.Context, input CreateFromTemplateInput) (*models.Workflow, error) {
	template, err := s.templateRepo.FindByID(ctx, input.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Process nodes with variable substitution
	nodes := s.substituteVariables(template.Nodes, input.Variables)
	settings := s.substituteSettingsVariables(template.Settings, input.Variables)

	workflow := &models.Workflow{
		WorkspaceID: input.WorkspaceID,
		CreatedBy:   input.UserID,
		Name:        input.Name,
		Description: template.Description,
		Status:      models.WorkflowStatusDraft,
		Version:     1,
		Nodes:       nodes,
		Connections: template.Connections,
		Settings:    settings,
		Tags:        template.Tags,
	}

	if err := s.workflowRepo.Create(ctx, workflow); err != nil {
		return nil, err
	}

	// Increment template use count
	_ = s.templateRepo.IncrementUseCount(ctx, template.ID)

	return workflow, nil
}

func (s *TemplateService) substituteVariables(nodes models.JSONArray, variables models.JSON) models.JSONArray {
	if variables == nil {
		return nodes
	}

	result := make(models.JSONArray, len(nodes))
	for i, node := range nodes {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			result[i] = node
			continue
		}

		// Deep copy and substitute
		newNode := make(map[string]interface{})
		for k, v := range nodeMap {
			newNode[k] = s.substituteValue(v, variables)
		}
		result[i] = newNode
	}
	return result
}

func (s *TemplateService) substituteSettingsVariables(settings models.JSON, variables models.JSON) models.JSON {
	if settings == nil || variables == nil {
		return settings
	}

	result := make(models.JSON)
	for k, v := range settings {
		result[k] = s.substituteValue(v, variables)
	}
	return result
}

func (s *TemplateService) substituteValue(value interface{}, variables models.JSON) interface{} {
	switch v := value.(type) {
	case string:
		// Check if it's a variable reference like {{var.name}}
		if len(v) > 4 && v[:2] == "{{" && v[len(v)-2:] == "}}" {
			varName := v[2 : len(v)-2]
			if varName[:4] == "var." {
				varKey := varName[4:]
				if val, exists := variables[varKey]; exists {
					return val
				}
			}
		}
		return v
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = s.substituteValue(val, variables)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = s.substituteValue(val, variables)
		}
		return result
	default:
		return value
	}
}

func (s *TemplateService) Delete(ctx context.Context, templateID uuid.UUID) error {
	return s.templateRepo.Delete(ctx, templateID)
}

func (s *TemplateService) Update(ctx context.Context, templateID uuid.UUID, input CreateTemplateInput) (*models.Template, error) {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	template.Name = input.Name
	template.Description = input.Description
	template.Category = input.Category
	template.Tags = input.Tags
	template.Nodes = input.Nodes
	template.Connections = input.Connections
	template.Settings = input.Settings
	template.Variables = input.Variables
	template.IconURL = input.IconURL
	template.IsPublic = input.IsPublic

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetCategories returns all unique template categories
func (s *TemplateService) GetCategories(ctx context.Context) ([]string, error) {
	templates, _, err := s.templateRepo.FindPublic(ctx, &repositories.ListOptions{Limit: 1000})
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[string]bool)
	for _, t := range templates {
		if t.Category != "" {
			categoryMap[t.Category] = true
		}
	}

	categories := make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	return categories, nil
}
