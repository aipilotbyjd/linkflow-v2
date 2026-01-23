package models

import (
	"time"

	"github.com/google/uuid"
)

// Variable represents the database model for variables
type Variable struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null"`
	Key         string     `gorm:"size:100;not null"`
	Value       string     `gorm:"type:text;not null"`
	Description *string    `gorm:"type:text"`
	IsSecret    bool       `gorm:"default:false"`
	Scope       string     `gorm:"size:20;default:'workspace'"`
	FolderID    *uuid.UUID `gorm:"type:uuid"`
	WorkflowID  *uuid.UUID `gorm:"type:uuid"`
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

// TableName returns the table name for Variable
func (Variable) TableName() string {
	return "variables"
}

// Environment represents the database model for environments
type Environment struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null"`
	Name        string    `gorm:"size:50;not null"`
	DisplayName string    `gorm:"size:100"`
	Description *string   `gorm:"type:text"`
	IsDefault   bool      `gorm:"default:false"`
	Color       *string   `gorm:"size:20"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name for Environment
func (Environment) TableName() string {
	return "environments"
}

// EnvironmentVariable represents the database model for environment variables
type EnvironmentVariable struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	EnvironmentID uuid.UUID `gorm:"type:uuid;index;not null"`
	VariableID    uuid.UUID `gorm:"type:uuid;index;not null"`
	Value         string    `gorm:"type:text;not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name for EnvironmentVariable
func (EnvironmentVariable) TableName() string {
	return "environment_variables"
}
