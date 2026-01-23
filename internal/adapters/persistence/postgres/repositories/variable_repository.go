package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/variable"
	"gorm.io/gorm"
)

// VariableRepository implements variable.Repository
type VariableRepository struct {
	db *gorm.DB
}

// NewVariableRepository creates a new variable repository
func NewVariableRepository(db *gorm.DB) *VariableRepository {
	return &VariableRepository{db: db}
}

// Variables

func (r *VariableRepository) CreateVariable(ctx context.Context, v *variable.Variable) error {
	model := toVariableModel(v)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *VariableRepository) UpdateVariable(ctx context.Context, v *variable.Variable) error {
	model := toVariableModel(v)
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *VariableRepository) DeleteVariable(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Variable{}, "id = ?", id).Error
}

func (r *VariableRepository) FindVariableByID(ctx context.Context, id uuid.UUID) (*variable.Variable, error) {
	var model models.Variable
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toVariableDomain(model), nil
}

func (r *VariableRepository) FindVariableByKey(ctx context.Context, workspaceID uuid.UUID, key string) (*variable.Variable, error) {
	var model models.Variable
	if err := r.db.WithContext(ctx).Where("workspace_id = ? AND key = ?", workspaceID, key).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toVariableDomain(model), nil
}

func (r *VariableRepository) FindVariablesByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]variable.Variable, error) {
	var models []models.Variable
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&models).Error; err != nil {
		return nil, err
	}
	return toVariableDomainList(models), nil
}

func (r *VariableRepository) FindVariablesByScope(ctx context.Context, workspaceID uuid.UUID, scope variable.Scope, scopeID *uuid.UUID) ([]variable.Variable, error) {
	query := r.db.WithContext(ctx).Where("workspace_id = ? AND scope = ?", workspaceID, string(scope))
	if scopeID != nil {
		if scope == variable.ScopeFolder {
			query = query.Where("folder_id = ?", scopeID)
		} else if scope == variable.ScopeWorkflow {
			query = query.Where("workflow_id = ?", scopeID)
		}
	}
	var models []models.Variable
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	return toVariableDomainList(models), nil
}

// Environments

func (r *VariableRepository) CreateEnvironment(ctx context.Context, e *variable.Environment) error {
	model := toEnvironmentModel(e)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *VariableRepository) UpdateEnvironment(ctx context.Context, e *variable.Environment) error {
	model := toEnvironmentModel(e)
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *VariableRepository) DeleteEnvironment(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete environment variables first
		if err := tx.Delete(&models.EnvironmentVariable{}, "environment_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Environment{}, "id = ?", id).Error
	})
}

func (r *VariableRepository) FindEnvironmentByID(ctx context.Context, id uuid.UUID) (*variable.Environment, error) {
	var model models.Environment
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toEnvironmentDomain(model), nil
}

func (r *VariableRepository) FindEnvironmentByName(ctx context.Context, workspaceID uuid.UUID, name string) (*variable.Environment, error) {
	var model models.Environment
	if err := r.db.WithContext(ctx).Where("workspace_id = ? AND name = ?", workspaceID, name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toEnvironmentDomain(model), nil
}

func (r *VariableRepository) FindEnvironmentsByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]variable.Environment, error) {
	var models []models.Environment
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	return toEnvironmentDomainList(models), nil
}

func (r *VariableRepository) GetDefaultEnvironment(ctx context.Context, workspaceID uuid.UUID) (*variable.Environment, error) {
	var model models.Environment
	if err := r.db.WithContext(ctx).Where("workspace_id = ? AND is_default = ?", workspaceID, true).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toEnvironmentDomain(model), nil
}

func (r *VariableRepository) SetDefaultEnvironment(ctx context.Context, workspaceID, environmentID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset current default
		if err := tx.Model(&models.Environment{}).Where("workspace_id = ? AND is_default = ?", workspaceID, true).Update("is_default", false).Error; err != nil {
			return err
		}
		// Set new default
		return tx.Model(&models.Environment{}).Where("id = ?", environmentID).Update("is_default", true).Error
	})
}

// Environment Variables

func (r *VariableRepository) SetEnvironmentVariable(ctx context.Context, ev *variable.EnvironmentVar) error {
	model := toEnvironmentVarModel(ev)
	// Upsert
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *VariableRepository) DeleteEnvironmentVariable(ctx context.Context, environmentID, variableID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.EnvironmentVariable{}, "environment_id = ? AND variable_id = ?", environmentID, variableID).Error
}

func (r *VariableRepository) FindEnvironmentVariables(ctx context.Context, environmentID uuid.UUID) ([]variable.EnvironmentVar, error) {
	var models []models.EnvironmentVariable
	if err := r.db.WithContext(ctx).Where("environment_id = ?", environmentID).Find(&models).Error; err != nil {
		return nil, err
	}
	return toEnvironmentVarDomainList(models), nil
}

// Resolution

func (r *VariableRepository) ResolveVariables(ctx context.Context, workspaceID uuid.UUID, environmentName string, workflowID *uuid.UUID) (*variable.VariableSet, error) {
	// 1. Get base workspace variables
	vars, err := r.FindVariablesByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	result := variable.NewVariableSet(environmentName)
	varMap := make(map[uuid.UUID]variable.Variable)

	// Add workspace variables (base)
	for _, v := range vars {
		result.Set(v.Key, v.Value)
		varMap[v.ID] = v
	}

	// 2. Get environment if specified
	if environmentName != "" {
		env, err := r.FindEnvironmentByName(ctx, workspaceID, environmentName)
		if err != nil {
			return nil, err
		}
		if env != nil {
			// Get overrides
			envVars, err := r.FindEnvironmentVariables(ctx, env.ID)
			if err != nil {
				return nil, err
			}
			for _, ev := range envVars {
				if v, ok := varMap[ev.VariableID]; ok {
					result.Set(v.Key, ev.Value)
				}
			}
		}
	}

	return result, nil
}

// Helpers

func toVariableModel(v *variable.Variable) models.Variable {
	return models.Variable{
		ID:          v.ID,
		WorkspaceID: v.WorkspaceID,
		Key:         v.Key,
		Value:       v.Value,
		Description: v.Description,
		IsSecret:    v.IsSecret,
		Scope:       string(v.Scope),
		FolderID:    v.FolderID,
		WorkflowID:  v.WorkflowID,
		CreatedBy:   v.CreatedBy,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func toVariableDomain(m models.Variable) *variable.Variable {
	return &variable.Variable{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		Key:         m.Key,
		Value:       m.Value,
		Description: m.Description,
		IsSecret:    m.IsSecret,
		Scope:       variable.Scope(m.Scope),
		FolderID:    m.FolderID,
		WorkflowID:  m.WorkflowID,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toVariableDomainList(list []models.Variable) []variable.Variable {
	result := make([]variable.Variable, len(list))
	for i, m := range list {
		result[i] = *toVariableDomain(m)
	}
	return result
}

func toEnvironmentModel(e *variable.Environment) models.Environment {
	return models.Environment{
		ID:          e.ID,
		WorkspaceID: e.WorkspaceID,
		Name:        e.Name,
		DisplayName: e.DisplayName,
		Description: e.Description,
		IsDefault:   e.IsDefault,
		Color:       e.Color,
		CreatedBy:   e.CreatedBy,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func toEnvironmentDomain(m models.Environment) *variable.Environment {
	return &variable.Environment{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Description: m.Description,
		IsDefault:   m.IsDefault,
		Color:       m.Color,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toEnvironmentDomainList(list []models.Environment) []variable.Environment {
	result := make([]variable.Environment, len(list))
	for i, m := range list {
		result[i] = *toEnvironmentDomain(m)
	}
	return result
}

func toEnvironmentVarModel(e *variable.EnvironmentVar) models.EnvironmentVariable {
	return models.EnvironmentVariable{
		ID:            e.ID,
		EnvironmentID: e.EnvironmentID,
		VariableID:    e.VariableID,
		Value:         e.Value,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func toEnvironmentVarDomain(m models.EnvironmentVariable) *variable.EnvironmentVar {
	return &variable.EnvironmentVar{
		ID:            m.ID,
		EnvironmentID: m.EnvironmentID,
		VariableID:    m.VariableID,
		Value:         m.Value,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func toEnvironmentVarDomainList(list []models.EnvironmentVariable) []variable.EnvironmentVar {
	result := make([]variable.EnvironmentVar, len(list))
	for i, m := range list {
		result[i] = *toEnvironmentVarDomain(m)
	}
	return result
}
