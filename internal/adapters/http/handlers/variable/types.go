package variable

import (
	"context"

	"github.com/google/uuid"
	domainvar "github.com/linkflow-ai/linkflow/internal/core/domain/variable"
)

// VariableRequest represents a variable create/update request
type VariableRequest struct {
	Key         string  `json:"key" validate:"required,max=100"`
	Value       string  `json:"value" validate:"required"`
	Description *string `json:"description,omitempty"`
	IsSecret    bool    `json:"is_secret,omitempty"`
}

// VariableResponse represents a variable in responses
type VariableResponse struct {
	ID          string  `json:"id"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Description *string `json:"description,omitempty"`
	IsSecret    bool    `json:"is_secret"`
	Scope       string  `json:"scope"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// EnvironmentRequest represents an environment create/update request
type EnvironmentRequest struct {
	Name        string  `json:"name" validate:"required,max=50"`
	DisplayName string  `json:"display_name" validate:"required,max=100"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	IsDefault   bool    `json:"is_default,omitempty"`
}

// EnvironmentResponse represents an environment in responses
type EnvironmentResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	DisplayName string                   `json:"display_name"`
	Description *string                  `json:"description,omitempty"`
	Color       *string                  `json:"color,omitempty"`
	IsDefault   bool                     `json:"is_default"`
	Variables   []EnvironmentVarResponse `json:"variables,omitempty"`
	CreatedAt   string                   `json:"created_at"`
}

// EnvironmentVarResponse represents an environment variable override
type EnvironmentVarResponse struct {
	VariableID  string `json:"variable_id"`
	VariableKey string `json:"variable_key"`
	Value       string `json:"value"`
}

// SetEnvVarRequest represents a request to set an environment variable
type SetEnvVarRequest struct {
	VariableID string `json:"variable_id" validate:"required"`
	Value      string `json:"value" validate:"required"`
}

// ResolveRequest represents a request to resolve variables
type ResolveRequest struct {
	Environment string  `json:"environment"`
	WorkflowID  *string `json:"workflow_id,omitempty"`
}

// ResolveResponse represents resolved variables
type ResolveResponse struct {
	Variables   map[string]string `json:"variables"`
	Environment string            `json:"environment"`
}

// Service interface for dependency injection
type Service interface {
	// Variables
	CreateVariable(ctx context.Context, workspaceID, userID uuid.UUID, req VariableRequest) (*domainvar.Variable, error)
	UpdateVariable(ctx context.Context, id uuid.UUID, req VariableRequest) (*domainvar.Variable, error)
	DeleteVariable(ctx context.Context, id uuid.UUID) error
	GetVariable(ctx context.Context, id uuid.UUID) (*domainvar.Variable, error)
	ListVariables(ctx context.Context, workspaceID uuid.UUID) ([]domainvar.Variable, error)

	// Environments
	CreateEnvironment(ctx context.Context, workspaceID, userID uuid.UUID, req EnvironmentRequest) (*domainvar.Environment, error)
	UpdateEnvironment(ctx context.Context, id uuid.UUID, req EnvironmentRequest) (*domainvar.Environment, error)
	DeleteEnvironment(ctx context.Context, id uuid.UUID) error
	GetEnvironment(ctx context.Context, id uuid.UUID) (*domainvar.Environment, error)
	ListEnvironments(ctx context.Context, workspaceID uuid.UUID) ([]domainvar.Environment, error)
	SetEnvironmentVariable(ctx context.Context, envID, varID uuid.UUID, value string) error

	// Resolution
	ResolveVariables(ctx context.Context, workspaceID uuid.UUID, env string, workflowID *uuid.UUID) (*domainvar.VariableSet, error)
}

// ToVariableResponse converts domain variable to response
func ToVariableResponse(v *domainvar.Variable) VariableResponse {
	value := v.Value
	if v.IsSecret {
		value = "********"
	}
	return VariableResponse{
		ID:          v.ID.String(),
		Key:         v.Key,
		Value:       value,
		Description: v.Description,
		IsSecret:    v.IsSecret,
		Scope:       string(v.Scope),
		CreatedAt:   v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToEnvironmentResponse converts domain environment to response
func ToEnvironmentResponse(e *domainvar.Environment) EnvironmentResponse {
	return EnvironmentResponse{
		ID:          e.ID.String(),
		Name:        e.Name,
		DisplayName: e.DisplayName,
		Description: e.Description,
		Color:       e.Color,
		IsDefault:   e.IsDefault,
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
