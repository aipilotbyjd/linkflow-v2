package variable

import (
	"time"

	"github.com/google/uuid"
)

// Variable represents a workspace-level variable
type Variable struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	WorkspaceID uuid.UUID  `json:"workspace_id" gorm:"type:uuid;index;not null"`
	Key         string     `json:"key" gorm:"size:100;not null"`
	Value       string     `json:"value" gorm:"type:text;not null"`
	Description *string    `json:"description,omitempty" gorm:"type:text"`
	IsSecret    bool       `json:"is_secret" gorm:"default:false"`
	Scope       Scope      `json:"scope" gorm:"size:20;default:workspace"`
	FolderID    *uuid.UUID `json:"folder_id,omitempty" gorm:"type:uuid"`
	WorkflowID  *uuid.UUID `json:"workflow_id,omitempty" gorm:"type:uuid"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Scope defines the variable scope
type Scope string

const (
	ScopeWorkspace Scope = "workspace"
	ScopeFolder    Scope = "folder"
	ScopeWorkflow  Scope = "workflow"
)

func (Variable) TableName() string {
	return "variables"
}

// NewVariable creates a new workspace variable
func NewVariable(workspaceID, createdBy uuid.UUID, key, value string) (*Variable, error) {
	if workspaceID == uuid.Nil {
		return nil, ErrInvalidWorkspaceID
	}
	if createdBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}
	if key == "" {
		return nil, ErrKeyRequired
	}
	if len(key) > 100 {
		return nil, ErrKeyTooLong
	}
	if value == "" {
		return nil, ErrValueRequired
	}

	return &Variable{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Key:         key,
		Value:       value,
		Scope:       ScopeWorkspace,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// Update updates the variable
func (v *Variable) Update(value string, description *string) {
	v.Value = value
	v.Description = description
	v.UpdatedAt = time.Now()
}

// MarkAsSecret marks the variable as secret
func (v *Variable) MarkAsSecret() {
	v.IsSecret = true
	v.UpdatedAt = time.Now()
}

// Environment represents an environment configuration
type Environment struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey"`
	WorkspaceID uuid.UUID        `json:"workspace_id" gorm:"type:uuid;index;not null"`
	Name        string           `json:"name" gorm:"size:50;not null"` // development, staging, production
	DisplayName string           `json:"display_name" gorm:"size:100"`
	Description *string          `json:"description,omitempty" gorm:"type:text"`
	IsDefault   bool             `json:"is_default" gorm:"default:false"`
	Color       *string          `json:"color,omitempty" gorm:"size:20"`
	Variables   []EnvironmentVar `json:"variables" gorm:"-"`
	CreatedBy   uuid.UUID        `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (Environment) TableName() string {
	return "environments"
}

// EnvironmentVar represents a variable value for a specific environment
type EnvironmentVar struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	EnvironmentID uuid.UUID `json:"environment_id" gorm:"type:uuid;index;not null"`
	VariableID    uuid.UUID `json:"variable_id" gorm:"type:uuid;index;not null"`
	Value         string    `json:"value" gorm:"type:text;not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (EnvironmentVar) TableName() string {
	return "environment_variables"
}

// NewEnvironment creates a new environment
func NewEnvironment(workspaceID, createdBy uuid.UUID, name, displayName string) *Environment {
	return &Environment{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        name,
		DisplayName: displayName,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// DefaultEnvironments returns the default environments for a workspace
func DefaultEnvironments(workspaceID, createdBy uuid.UUID) []*Environment {
	return []*Environment{
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			Name:        "development",
			DisplayName: "Development",
			IsDefault:   true,
			Color:       strPtr("#22c55e"),
			CreatedBy:   createdBy,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			Name:        "staging",
			DisplayName: "Staging",
			Color:       strPtr("#f59e0b"),
			CreatedBy:   createdBy,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			Name:        "production",
			DisplayName: "Production",
			Color:       strPtr("#ef4444"),
			CreatedBy:   createdBy,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

func strPtr(s string) *string {
	return &s
}

// VariableSet represents a resolved set of variables for execution
type VariableSet struct {
	Variables   map[string]string `json:"variables"`
	Environment string            `json:"environment"`
	ResolvedAt  time.Time         `json:"resolved_at"`
}

// NewVariableSet creates a new variable set
func NewVariableSet(env string) *VariableSet {
	return &VariableSet{
		Variables:   make(map[string]string),
		Environment: env,
		ResolvedAt:  time.Now(),
	}
}

// Get returns a variable value
func (vs *VariableSet) Get(key string) (string, bool) {
	val, ok := vs.Variables[key]
	return val, ok
}

// Set sets a variable value
func (vs *VariableSet) Set(key, value string) {
	vs.Variables[key] = value
}

// Merge merges another variable set (other takes precedence)
func (vs *VariableSet) Merge(other *VariableSet) {
	for k, v := range other.Variables {
		vs.Variables[k] = v
	}
}
