package aibuilder

import (
	"time"

	"github.com/google/uuid"
)

// GenerationRequest represents a request to generate a workflow from natural language
type GenerationRequest struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	UserID      uuid.UUID  `json:"user_id"`
	Prompt      string     `json:"prompt"`
	Context     *Context   `json:"context,omitempty"`
	Status      Status     `json:"status"`
	Result      *Result    `json:"result,omitempty"`
	Error       *string    `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Context provides additional context for workflow generation
type Context struct {
	ExistingWorkflowID *uuid.UUID        `json:"existing_workflow_id,omitempty"`
	AvailableCredTypes []string          `json:"available_cred_types,omitempty"`
	PreferredTrigger   *string           `json:"preferred_trigger,omitempty"`
	Constraints        map[string]string `json:"constraints,omitempty"`
}

// Result contains the generated workflow
type Result struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Nodes       []map[string]interface{} `json:"nodes"`
	Connections []map[string]interface{} `json:"connections"`
	Settings    map[string]interface{}   `json:"settings,omitempty"`
	Explanation string                   `json:"explanation"`
	Suggestions []string                 `json:"suggestions,omitempty"`
}

// Status represents the generation request status
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// NewGenerationRequest creates a new generation request
func NewGenerationRequest(workspaceID, userID uuid.UUID, prompt string, ctx *Context) *GenerationRequest {
	return &GenerationRequest{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Prompt:      prompt,
		Context:     ctx,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
	}
}

// MarkProcessing marks the request as processing
func (r *GenerationRequest) MarkProcessing() {
	r.Status = StatusProcessing
}

// MarkCompleted marks the request as completed with result
func (r *GenerationRequest) MarkCompleted(result *Result) {
	r.Status = StatusCompleted
	r.Result = result
	now := time.Now()
	r.CompletedAt = &now
}

// MarkFailed marks the request as failed
func (r *GenerationRequest) MarkFailed(err string) {
	r.Status = StatusFailed
	r.Error = &err
	now := time.Now()
	r.CompletedAt = &now
}

// NodeTemplate represents a template for generating nodes
type NodeTemplate struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Keywords    []string `json:"keywords"`
	Parameters  []ParameterTemplate `json:"parameters"`
}

// ParameterTemplate describes a node parameter
type ParameterTemplate struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}
