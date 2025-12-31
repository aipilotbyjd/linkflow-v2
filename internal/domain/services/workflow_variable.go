package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type WorkflowVariableService struct {
	db *gorm.DB
}

func NewWorkflowVariableService(db *gorm.DB) *WorkflowVariableService {
	return &WorkflowVariableService{db: db}
}

type CreateVariableInput struct {
	WorkflowID  uuid.UUID
	Name        string
	Key         string
	Type        string
	Value       string
	Default     string
	Description *string
	Required    bool
}

type UpdateVariableInput struct {
	Name        *string
	Value       *string
	Default     *string
	Description *string
	Required    *bool
}

func (s *WorkflowVariableService) Create(ctx context.Context, input CreateVariableInput) (*models.WorkflowVariable, error) {
	if input.Type == "" {
		input.Type = "string"
	}

	validTypes := map[string]bool{"string": true, "number": true, "boolean": true, "json": true, "secret": true}
	if !validTypes[input.Type] {
		return nil, fmt.Errorf("invalid variable type: %s", input.Type)
	}

	var existing models.WorkflowVariable
	err := s.db.WithContext(ctx).
		Where("workflow_id = ? AND key = ?", input.WorkflowID, input.Key).
		First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("variable with key '%s' already exists", input.Key)
	}

	variable := &models.WorkflowVariable{
		WorkflowID:  input.WorkflowID,
		Name:        input.Name,
		Key:         input.Key,
		Type:        input.Type,
		Value:       input.Value,
		Default:     input.Default,
		Description: input.Description,
		Required:    input.Required,
	}

	if err := s.db.WithContext(ctx).Create(variable).Error; err != nil {
		return nil, err
	}

	return variable, nil
}

func (s *WorkflowVariableService) List(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowVariable, error) {
	var variables []models.WorkflowVariable
	err := s.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("name ASC").
		Find(&variables).Error
	return variables, err
}

func (s *WorkflowVariableService) Get(ctx context.Context, id uuid.UUID) (*models.WorkflowVariable, error) {
	var variable models.WorkflowVariable
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&variable).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &variable, nil
}

func (s *WorkflowVariableService) Update(ctx context.Context, id uuid.UUID, input UpdateVariableInput) (*models.WorkflowVariable, error) {
	variable, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Value != nil {
		updates["value"] = *input.Value
	}
	if input.Default != nil {
		updates["default"] = *input.Default
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Required != nil {
		updates["required"] = *input.Required
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(variable).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.Get(ctx, id)
}

func (s *WorkflowVariableService) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&models.WorkflowVariable{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

func (s *WorkflowVariableService) GetByWorkflowAndKey(ctx context.Context, workflowID uuid.UUID, key string) (*models.WorkflowVariable, error) {
	var variable models.WorkflowVariable
	err := s.db.WithContext(ctx).
		Where("workflow_id = ? AND key = ?", workflowID, key).
		First(&variable).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &variable, nil
}

func (s *WorkflowVariableService) GetResolvedVariables(ctx context.Context, workflowID uuid.UUID, overrides map[string]string) (map[string]interface{}, error) {
	variables, err := s.List(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, v := range variables {
		value := v.Value
		if value == "" {
			value = v.Default
		}
		if override, ok := overrides[v.Key]; ok {
			value = override
		}
		if v.Required && value == "" {
			return nil, fmt.Errorf("required variable '%s' has no value", v.Key)
		}
		result[v.Key] = value
	}

	return result, nil
}
