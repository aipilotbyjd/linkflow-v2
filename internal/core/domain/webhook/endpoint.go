package webhook

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Endpoint entity (aggregate root)
type Endpoint struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID   uuid.UUID      `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	NodeID       string         `gorm:"size:100;not null" json:"node_id"`
	Path         string         `gorm:"size:255;uniqueIndex;not null" json:"path"`
	Method       string         `gorm:"size:10;default:POST" json:"method"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Secret       *string        `gorm:"size:255" json:"-"`
	LastCalledAt *time.Time     `json:"last_called_at,omitempty"`
	CallCount    int            `gorm:"default:0" json:"call_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Endpoint) TableName() string {
	return "webhook_endpoints"
}

// NewEndpoint creates a new webhook endpoint
func NewEndpoint(workflowID, workspaceID uuid.UUID, nodeID, path string) *Endpoint {
	return &Endpoint{
		ID:          uuid.New(),
		WorkflowID:  workflowID,
		WorkspaceID: workspaceID,
		NodeID:      nodeID,
		Path:        path,
		Method:      "POST",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// GetWorkspaceID implements the WorkspaceOwned interface
func (e *Endpoint) GetWorkspaceID() uuid.UUID {
	return e.WorkspaceID
}

// WithMethod sets the HTTP method
func (e *Endpoint) WithMethod(method string) *Endpoint {
	e.Method = method
	return e
}

// WithSecret sets the webhook secret
func (e *Endpoint) WithSecret(secret string) *Endpoint {
	e.Secret = &secret
	return e
}

// Activate activates the endpoint
func (e *Endpoint) Activate() {
	e.IsActive = true
	e.UpdatedAt = time.Now()
}

// Deactivate deactivates the endpoint
func (e *Endpoint) Deactivate() {
	e.IsActive = false
	e.UpdatedAt = time.Now()
}

// RegenerateSecret sets a new secret
func (e *Endpoint) RegenerateSecret(secret string) {
	e.Secret = &secret
	e.UpdatedAt = time.Now()
}

// RecordCall records a webhook call
func (e *Endpoint) RecordCall() {
	now := time.Now()
	e.LastCalledAt = &now
	e.CallCount++
	e.UpdatedAt = now
}

// UpdatePath updates the webhook path
func (e *Endpoint) UpdatePath(path string) {
	e.Path = path
	e.UpdatedAt = time.Now()
}

// HasSecret checks if the endpoint has a secret configured
func (e *Endpoint) HasSecret() bool {
	return e.Secret != nil && *e.Secret != ""
}

// GetURL returns the full webhook URL
func (e *Endpoint) GetURL(baseURL string) string {
	return baseURL + "/webhooks/" + e.Path
}
