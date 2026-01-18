package workflow

import (
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// CreateWorkflowRequest represents workflow creation request
type CreateWorkflowRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description *string         `json:"description,omitempty"`
	Nodes       types.JSONArray `json:"nodes"`
	Connections types.JSONArray `json:"connections"`
	Settings    types.JSON      `json:"settings"`
	Tags        []string        `json:"tags"`
	Color       *string         `json:"color,omitempty"`
	Icon        *string         `json:"icon,omitempty"`
	Category    *string         `json:"category,omitempty"`
}

// WorkflowResponse represents workflow in responses
type WorkflowResponse struct {
	ID             string          `json:"id"`
	WorkspaceID    string          `json:"workspace_id"`
	Name           string          `json:"name"`
	Description    *string         `json:"description,omitempty"`
	Status         string          `json:"status"`
	Version        int             `json:"version"`
	Nodes          types.JSONArray `json:"nodes"`
	Connections    types.JSONArray `json:"connections"`
	Settings       types.JSON      `json:"settings"`
	Tags           []string        `json:"tags"`
	Color          *string         `json:"color,omitempty"`
	Icon           *string         `json:"icon,omitempty"`
	Category       *string         `json:"category,omitempty"`
	IsFavorite     bool            `json:"is_favorite"`
	ExecutionCount int             `json:"execution_count"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

// UpdateRequest represents workflow update request
type UpdateRequest struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Nodes       types.JSONArray `json:"nodes,omitempty"`
	Connections types.JSONArray `json:"connections,omitempty"`
	Settings    types.JSON      `json:"settings,omitempty"`
}

// CloneRequest represents workflow clone request
type CloneRequest struct {
	Name string `json:"name" validate:"required"`
}

// DuplicateRequest represents workflow duplicate request
type DuplicateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	FolderID    *string `json:"folderId,omitempty"`
}

// ValidateRequest represents workflow validation request
type ValidateRequest struct {
	Nodes       []NodeDef       `json:"nodes"`
	Connections []ConnectionDef `json:"connections"`
}

// NodeDef represents a node definition for validation
type NodeDef struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Position   map[string]float64     `json:"position"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ConnectionDef represents a connection definition for validation
type ConnectionDef struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// ValidationResult represents workflow validation result
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// ValidationError represents a validation error or warning
type ValidationError struct {
	NodeID  string `json:"node_id,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// TestNodeRequest represents test node request
type TestNodeRequest struct {
	NodeType   string                 `json:"node_type" validate:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Input      map[string]interface{} `json:"input,omitempty"`
}

// TestNodeResponse represents test node response
type TestNodeResponse struct {
	Success    bool                   `json:"success"`
	Output     map[string]interface{} `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
}

// VersionResponse represents workflow version in responses
type VersionResponse struct {
	ID          string          `json:"id"`
	WorkflowID  string          `json:"workflow_id"`
	Version     int             `json:"version"`
	Nodes       types.JSONArray `json:"nodes"`
	Connections types.JSONArray `json:"connections"`
	Settings    types.JSON      `json:"settings,omitempty"`
	ChangeNote  *string         `json:"change_note,omitempty"`
	CreatedBy   string          `json:"created_by"`
	CreatedAt   string          `json:"created_at"`
}

// CompareVersionsResponse represents version comparison result
type CompareVersionsResponse struct {
	Version1    int                 `json:"version1"`
	Version2    int                 `json:"version2"`
	Differences []VersionDifference `json:"differences"`
	Summary     DiffSummary         `json:"summary"`
}

// VersionDifference represents a difference between versions
type VersionDifference struct {
	Type        string      `json:"type"`
	Path        string      `json:"path"`
	OldValue    interface{} `json:"oldValue,omitempty"`
	NewValue    interface{} `json:"newValue,omitempty"`
	Description string      `json:"description"`
}

// DiffSummary represents a summary of differences
type DiffSummary struct {
	NodesAdded         int  `json:"nodesAdded"`
	NodesRemoved       int  `json:"nodesRemoved"`
	NodesModified      int  `json:"nodesModified"`
	ConnectionsAdded   int  `json:"connectionsAdded"`
	ConnectionsRemoved int  `json:"connectionsRemoved"`
	SettingsChanged    bool `json:"settingsChanged"`
}

// ToWorkflowResponse converts domain workflow to response
func ToWorkflowResponse(wf *workflow.Workflow) WorkflowResponse {
	return WorkflowResponse{
		ID:             wf.ID.String(),
		WorkspaceID:    wf.WorkspaceID.String(),
		Name:           wf.Name,
		Description:    wf.Description,
		Status:         string(wf.Status),
		Version:        wf.Version,
		Nodes:          wf.Nodes,
		Connections:    wf.Connections,
		Settings:       wf.Settings,
		Tags:           wf.Tags,
		Color:          wf.Color,
		Icon:           wf.Icon,
		Category:       wf.Category,
		IsFavorite:     wf.IsFavorite,
		ExecutionCount: wf.ExecutionCount,
		CreatedAt:      wf.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      wf.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToVersionResponse converts domain version to response
func ToVersionResponse(v *workflow.Version) VersionResponse {
	createdBy := ""
	if v.CreatedBy != nil {
		createdBy = v.CreatedBy.String()
	}
	return VersionResponse{
		ID:          v.ID.String(),
		WorkflowID:  v.WorkflowID.String(),
		Version:     v.Version,
		Nodes:       v.Nodes,
		Connections: v.Connections,
		Settings:    v.Settings,
		ChangeNote:  v.ChangeMessage,
		CreatedBy:   createdBy,
		CreatedAt:   v.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
