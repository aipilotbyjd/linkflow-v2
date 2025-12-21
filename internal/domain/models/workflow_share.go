package models

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowShare represents a workflow shared between workspaces
type WorkflowShare struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID          uuid.UUID  `gorm:"type:uuid;index;not null" json:"workflow_id"`
	SourceWorkspaceID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"source_workspace_id"`
	TargetWorkspaceID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"target_workspace_id"`
	SharedBy            uuid.UUID  `gorm:"type:uuid;not null" json:"shared_by"`
	Permission          string     `gorm:"size:20;not null;default:read" json:"permission"` // read, execute, edit
	AcceptedAt          *time.Time `json:"accepted_at,omitempty"`
	AcceptedBy          *uuid.UUID `gorm:"type:uuid" json:"accepted_by,omitempty"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`

	Workflow        Workflow  `gorm:"foreignKey:WorkflowID" json:"-"`
	SourceWorkspace Workspace `gorm:"foreignKey:SourceWorkspaceID" json:"-"`
	TargetWorkspace Workspace `gorm:"foreignKey:TargetWorkspaceID" json:"-"`
	Sharer          User      `gorm:"foreignKey:SharedBy" json:"-"`
}

func (WorkflowShare) TableName() string {
	return "workflow_shares"
}

// TemplateMarketplace represents published workflow templates
type TemplateMarketplace struct {
	ID            uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID    uuid.UUID   `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID   uuid.UUID   `gorm:"type:uuid;index;not null" json:"workspace_id"`
	PublishedBy   uuid.UUID   `gorm:"type:uuid;not null" json:"published_by"`
	Name          string      `gorm:"size:255;not null" json:"name"`
	Description   string      `gorm:"type:text" json:"description"`
	Category      string      `gorm:"size:50;not null;index" json:"category"`
	Tags          StringArray `gorm:"type:text[]" json:"tags"`
	Icon          string      `gorm:"size:100" json:"icon"`
	Nodes         JSONArray   `gorm:"type:jsonb;not null" json:"nodes"`
	Connections   JSONArray   `gorm:"type:jsonb;not null" json:"connections"`
	Settings      JSON        `gorm:"type:jsonb" json:"settings"`
	Variables     JSON        `gorm:"type:jsonb" json:"variables"`       // Template variables for customization
	IsPublic      bool        `gorm:"default:false;index" json:"is_public"`
	IsFeatured    bool        `gorm:"default:false;index" json:"is_featured"`
	UsageCount    int         `gorm:"default:0" json:"usage_count"`
	Rating        float64     `gorm:"default:0" json:"rating"`
	RatingCount   int         `gorm:"default:0" json:"rating_count"`
	Version       string      `gorm:"size:20;not null;default:1.0.0" json:"version"`
	PublishedAt   time.Time   `json:"published_at"`
	UpdatedAt     time.Time   `json:"updated_at"`

	Workflow  Workflow  `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Publisher User      `gorm:"foreignKey:PublishedBy" json:"-"`
}

func (TemplateMarketplace) TableName() string {
	return "template_marketplace"
}

// TemplateRating represents user ratings for marketplace templates
type TemplateRating struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TemplateID uuid.UUID `gorm:"type:uuid;index;not null" json:"template_id"`
	UserID     uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Rating     int       `gorm:"not null" json:"rating"` // 1-5
	Review     *string   `gorm:"type:text" json:"review,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Template TemplateMarketplace `gorm:"foreignKey:TemplateID" json:"-"`
	User     User                `gorm:"foreignKey:UserID" json:"-"`
}

func (TemplateRating) TableName() string {
	return "template_ratings"
}

// WorkflowVariable represents configurable variables in workflows
type WorkflowVariable struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  uuid.UUID `gorm:"type:uuid;index;not null" json:"workflow_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Key         string    `gorm:"size:100;not null" json:"key"`
	Type        string    `gorm:"size:20;not null" json:"type"` // string, number, boolean, json, secret
	Value       string    `gorm:"type:text" json:"value"`
	Default     string    `gorm:"type:text" json:"default"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	Required    bool      `gorm:"default:false" json:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Workflow Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
}

func (WorkflowVariable) TableName() string {
	return "workflow_variables"
}
