package share

import "time"

// WorkflowShare represents a workflow share
type WorkflowShare struct {
	ID              string     `json:"id"`
	WorkflowID      string     `json:"workflowId"`
	WorkflowName    string     `json:"workflowName"`
	SharedBy        string     `json:"sharedBy"`
	SharedByName    string     `json:"sharedByName"`
	SharedWith      string     `json:"sharedWith"`
	SharedWithName  string     `json:"sharedWithName,omitempty"`
	SharedWithEmail string     `json:"sharedWithEmail,omitempty"`
	Permission      string     `json:"permission"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"createdAt"`
	AcceptedAt      *time.Time `json:"acceptedAt,omitempty"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
}

// ShareRepository defines the share repository interface
type ShareRepository interface {
	Create(share *WorkflowShare) error
	GetByID(id string) (*WorkflowShare, error)
	GetSharedByMe(userID string) ([]WorkflowShare, error)
	GetSharedWithMe(userID string) ([]WorkflowShare, error)
	GetPending(userID string) ([]WorkflowShare, error)
	Accept(id string) error
	Update(id string, permission string) error
	Delete(id string) error
}
