package share

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/share"
)

// ShareResponse represents a share in API response
type ShareResponse struct {
	ID              string     `json:"id"`
	ResourceType    string     `json:"resourceType"`
	ResourceID      string     `json:"resourceId"`
	ResourceName    string     `json:"resourceName"`
	SharedByID      string     `json:"sharedById"`
	SharedByEmail   string     `json:"sharedByEmail"`
	SharedWithID    string     `json:"sharedWithId"`
	SharedWithEmail string     `json:"sharedWithEmail"`
	Permission      string     `json:"permission"`
	Status          string     `json:"status"`
	Message         string     `json:"message,omitempty"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// CreateShareRequest represents request to create a share
type CreateShareRequest struct {
	ResourceType    string `json:"resourceType" validate:"required,oneof=workflow credential folder"`
	ResourceID      string `json:"resourceId" validate:"required,uuid"`
	SharedWithEmail string `json:"sharedWithEmail" validate:"required,email"`
	Permission      string `json:"permission" validate:"required,oneof=view edit admin"`
	Message         string `json:"message,omitempty"`
}

// UpdateShareRequest represents request to update a share
type UpdateShareRequest struct {
	Permission string `json:"permission"`
}

// ToShareResponse converts domain to response
func ToShareResponse(s share.Share) ShareResponse {
	return ShareResponse{
		ID:              s.ID.String(),
		ResourceType:    s.ResourceType,
		ResourceID:      s.ResourceID.String(),
		ResourceName:    s.ResourceName,
		SharedByID:      s.SharedByID.String(),
		SharedByEmail:   s.SharedByEmail,
		SharedWithID:    s.SharedWithID.String(),
		SharedWithEmail: s.SharedWithEmail,
		Permission:      string(s.Permission),
		Status:          string(s.Status),
		Message:         s.Message,
		ExpiresAt:       s.ExpiresAt,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

// ToShareResponseList converts domain list to response list
func ToShareResponseList(list []share.Share) []ShareResponse {
	result := make([]ShareResponse, len(list))
	for i, s := range list {
		result[i] = ToShareResponse(s)
	}
	return result
}
