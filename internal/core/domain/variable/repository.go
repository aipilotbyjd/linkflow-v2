package variable

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the variable repository interface
type Repository interface {
	// Variables
	CreateVariable(ctx context.Context, v *Variable) error
	UpdateVariable(ctx context.Context, v *Variable) error
	DeleteVariable(ctx context.Context, id uuid.UUID) error
	FindVariableByID(ctx context.Context, id uuid.UUID) (*Variable, error)
	FindVariableByKey(ctx context.Context, workspaceID uuid.UUID, key string) (*Variable, error)
	FindVariablesByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Variable, error)
	FindVariablesByScope(ctx context.Context, workspaceID uuid.UUID, scope Scope, scopeID *uuid.UUID) ([]Variable, error)

	// Environments
	CreateEnvironment(ctx context.Context, e *Environment) error
	UpdateEnvironment(ctx context.Context, e *Environment) error
	DeleteEnvironment(ctx context.Context, id uuid.UUID) error
	FindEnvironmentByID(ctx context.Context, id uuid.UUID) (*Environment, error)
	FindEnvironmentByName(ctx context.Context, workspaceID uuid.UUID, name string) (*Environment, error)
	FindEnvironmentsByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Environment, error)
	GetDefaultEnvironment(ctx context.Context, workspaceID uuid.UUID) (*Environment, error)
	SetDefaultEnvironment(ctx context.Context, workspaceID, environmentID uuid.UUID) error

	// Environment Variables
	SetEnvironmentVariable(ctx context.Context, ev *EnvironmentVar) error
	DeleteEnvironmentVariable(ctx context.Context, environmentID, variableID uuid.UUID) error
	FindEnvironmentVariables(ctx context.Context, environmentID uuid.UUID) ([]EnvironmentVar, error)

	// Resolution
	ResolveVariables(ctx context.Context, workspaceID uuid.UUID, environmentName string, workflowID *uuid.UUID) (*VariableSet, error)
}
