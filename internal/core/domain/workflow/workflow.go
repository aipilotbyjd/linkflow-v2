package workflow

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// Workflow entity (aggregate root)
type Workflow struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy       uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name            string         `gorm:"size:255;not null" json:"name"`
	Description     *string        `gorm:"type:text" json:"description,omitempty"`
	Status          Status         `gorm:"size:20;not null;default:draft;index" json:"status"`
	Version         int            `gorm:"not null;default:1" json:"version"`
	Nodes           types.JSONArray `gorm:"type:jsonb;not null;default:'[]'" json:"nodes"`
	Connections     types.JSONArray `gorm:"type:jsonb;not null;default:'[]'" json:"connections"`
	Settings        types.JSON     `gorm:"type:jsonb;default:'{}'" json:"settings"`
	Tags            types.StringArray `gorm:"type:text[]" json:"tags"`
	Color           *string        `gorm:"size:20" json:"color,omitempty"`
	Icon            *string        `gorm:"size:50" json:"icon,omitempty"`
	Category        *string        `gorm:"size:50" json:"category,omitempty"`
	IsFavorite      bool           `gorm:"default:false" json:"is_favorite"`
	FolderID        *uuid.UUID     `gorm:"type:uuid;column:project_id" json:"folder_id,omitempty"`
	ErrorWorkflowID *uuid.UUID     `gorm:"type:uuid" json:"error_workflow_id,omitempty"`
	ErrorTrigger    *string        `gorm:"size:50" json:"error_trigger,omitempty"`
	ExecutionCount  int            `gorm:"default:0" json:"execution_count"`
	LastExecutedAt  *time.Time     `json:"last_executed_at,omitempty"`
	ActivatedAt     *time.Time     `json:"activated_at,omitempty"`
	ArchivedAt      *time.Time     `json:"archived_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Workflow) TableName() string {
	return "workflows"
}

// NewWorkflow creates a new workflow
func NewWorkflow(workspaceID, createdBy uuid.UUID, name string) *Workflow {
	return &Workflow{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		CreatedBy:   createdBy,
		Name:        name,
		Status:      StatusDraft,
		Version:     1,
		Nodes:       make(types.JSONArray, 0),
		Connections: make(types.JSONArray, 0),
		Settings:    make(types.JSON),
		Tags:        make(types.StringArray, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// GetWorkspaceID implements the WorkspaceOwned interface
func (w *Workflow) GetWorkspaceID() uuid.UUID {
	return w.WorkspaceID
}

// IsActive checks if workflow is active
func (w *Workflow) IsActive() bool {
	return w.Status == StatusActive
}

// IsDraft checks if workflow is in draft status
func (w *Workflow) IsDraft() bool {
	return w.Status == StatusDraft
}

// IsArchived checks if workflow is archived
func (w *Workflow) IsArchived() bool {
	return w.Status == StatusArchived
}

// CanActivate checks if workflow can be activated
func (w *Workflow) CanActivate() error {
	if len(w.Nodes) == 0 {
		return ErrEmptyWorkflow
	}
	if !w.HasTriggerNode() {
		return ErrNoTriggerNode
	}
	return nil
}

// Activate activates the workflow
func (w *Workflow) Activate() error {
	if err := w.CanActivate(); err != nil {
		return err
	}
	w.Status = StatusActive
	now := time.Now()
	w.ActivatedAt = &now
	w.UpdatedAt = now
	return nil
}

// Deactivate deactivates the workflow
func (w *Workflow) Deactivate() {
	w.Status = StatusInactive
	w.UpdatedAt = time.Now()
}

// Archive archives the workflow
func (w *Workflow) Archive() {
	w.Status = StatusArchived
	now := time.Now()
	w.ArchivedAt = &now
	w.UpdatedAt = now
}

// Unarchive unarchives the workflow
func (w *Workflow) Unarchive() {
	w.Status = StatusDraft
	w.ArchivedAt = nil
	w.UpdatedAt = time.Now()
}

// Update updates workflow content
func (w *Workflow) Update(name string, description *string, nodes types.JSONArray, connections types.JSONArray, settings types.JSON) {
	w.Name = name
	w.Description = description
	w.Nodes = nodes
	w.Connections = connections
	w.Settings = settings
	w.Version++
	w.UpdatedAt = time.Now()
}

// UpdateMetadata updates workflow metadata
func (w *Workflow) UpdateMetadata(color, icon, category *string, tags []string, isFavorite bool) {
	w.Color = color
	w.Icon = icon
	w.Category = category
	w.Tags = tags
	w.IsFavorite = isFavorite
	w.UpdatedAt = time.Now()
}

// SetErrorWorkflow sets the error workflow
func (w *Workflow) SetErrorWorkflow(errorWorkflowID *uuid.UUID, trigger *string) {
	w.ErrorWorkflowID = errorWorkflowID
	w.ErrorTrigger = trigger
	w.UpdatedAt = time.Now()
}

// MoveToFolder moves the workflow to a folder
func (w *Workflow) MoveToFolder(folderID *uuid.UUID) {
	w.FolderID = folderID
	w.UpdatedAt = time.Now()
}

// IncrementExecutionCount increments the execution counter
func (w *Workflow) IncrementExecutionCount() {
	w.ExecutionCount++
	now := time.Now()
	w.LastExecutedAt = &now
	w.UpdatedAt = now
}

// HasTriggerNode checks if workflow has at least one trigger node
func (w *Workflow) HasTriggerNode() bool {
	for _, n := range w.Nodes {
		if node, ok := n.(map[string]interface{}); ok {
			if nodeType, ok := node["type"].(string); ok {
				if isTriggerType(nodeType) {
					return true
				}
			}
		}
	}
	return false
}

// GetSetting retrieves a setting value
func (w *Workflow) GetSetting(key string) interface{} {
	if w.Settings == nil {
		return nil
	}
	return w.Settings[key]
}

// GetTimeoutSeconds returns the timeout in seconds
func (w *Workflow) GetTimeoutSeconds() int {
	if timeout := w.Settings.GetInt("timeout_seconds"); timeout > 0 {
		return timeout
	}
	return 3600 // Default 1 hour
}

// GetMaxRetries returns the maximum retry count
func (w *Workflow) GetMaxRetries() int {
	if retries := w.Settings.GetInt("max_retries"); retries > 0 {
		return retries
	}
	return 3
}

func isTriggerType(nodeType string) bool {
	triggers := []string{
		"trigger.manual",
		"trigger.webhook",
		"trigger.schedule",
		"trigger.api",
		"n8n-nodes-base.manualTrigger",
		"n8n-nodes-base.webhook",
		"n8n-nodes-base.scheduleTrigger",
	}
	for _, t := range triggers {
		if nodeType == t {
			return true
		}
	}
	return false
}
