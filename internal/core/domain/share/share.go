package share

import (
	"time"

	"github.com/google/uuid"
)

// ShareStatus represents the status of a share
type ShareStatus string

const (
	StatusPending  ShareStatus = "pending"
	StatusAccepted ShareStatus = "accepted"
	StatusRejected ShareStatus = "rejected"
)

// SharePermission represents what can be done with shared resource
type SharePermission string

const (
	PermissionView SharePermission = "view"
	PermissionUse  SharePermission = "use"
	PermissionEdit SharePermission = "edit"
)

// Share represents a shared credential or resource
type Share struct {
	ID              uuid.UUID       `json:"id"`
	ResourceType    string          `json:"resourceType"` // "credential", "workflow"
	ResourceID      uuid.UUID       `json:"resourceId"`
	ResourceName    string          `json:"resourceName"`
	SharedByID      uuid.UUID       `json:"sharedById"`
	SharedByEmail   string          `json:"sharedByEmail"`
	SharedWithID    uuid.UUID       `json:"sharedWithId"`
	SharedWithEmail string          `json:"sharedWithEmail"`
	Permission      SharePermission `json:"permission"`
	Status          ShareStatus     `json:"status"`
	Message         string          `json:"message,omitempty"`
	ExpiresAt       *time.Time      `json:"expiresAt,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// NewShare creates a new share
func NewShare(resourceType string, resourceID uuid.UUID, resourceName string, sharedByID uuid.UUID, sharedByEmail string, sharedWithID uuid.UUID, sharedWithEmail string, permission SharePermission) (*Share, error) {
	if resourceType == "" {
		return nil, ErrResourceTypeRequired
	}
	if resourceID == uuid.Nil {
		return nil, ErrResourceIDRequired
	}
	if resourceName == "" {
		return nil, ErrResourceNameRequired
	}
	if sharedByID == uuid.Nil {
		return nil, ErrSharedByRequired
	}
	if sharedByEmail == "" {
		return nil, ErrSharedByEmailRequired
	}
	if sharedWithID == uuid.Nil {
		return nil, ErrSharedWithRequired
	}
	if sharedWithEmail == "" {
		return nil, ErrSharedWithEmailRequired
	}
	if permission == "" {
		permission = PermissionView
	}

	now := time.Now()
	return &Share{
		ID:              uuid.New(),
		ResourceType:    resourceType,
		ResourceID:      resourceID,
		ResourceName:    resourceName,
		SharedByID:      sharedByID,
		SharedByEmail:   sharedByEmail,
		SharedWithID:    sharedWithID,
		SharedWithEmail: sharedWithEmail,
		Permission:      permission,
		Status:          StatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Accept marks the share as accepted
func (s *Share) Accept() {
	s.Status = StatusAccepted
	s.UpdatedAt = time.Now()
}

// Reject marks the share as rejected
func (s *Share) Reject() {
	s.Status = StatusRejected
	s.UpdatedAt = time.Now()
}
