package note

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/note"
)

// NoteResponse represents a note in API responses
type NoteResponse struct {
	ID          uuid.UUID        `json:"id"`
	WorkspaceID uuid.UUID        `json:"workspace_id"`
	WorkflowID  uuid.UUID        `json:"workflow_id"`
	UserID      uuid.UUID        `json:"user_id"`
	NodeID      *string          `json:"node_id,omitempty"`
	Content     string           `json:"content"`
	Resolved    bool             `json:"resolved"`
	ResolvedAt  *time.Time       `json:"resolved_at,omitempty"`
	ResolvedBy  *uuid.UUID       `json:"resolved_by,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	User        *UserBrief       `json:"user,omitempty"`
	ResolvedByUser *UserBrief    `json:"resolved_by_user,omitempty"`
}

// UserBrief is a minimal user representation for embedding
type UserBrief struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

// CreateNoteRequest is the request body for creating a note
type CreateNoteRequest struct {
	Content string  `json:"content" validate:"required,min=1,max=10000"`
	NodeID  *string `json:"node_id,omitempty" validate:"omitempty,max=100"`
}

// UpdateNoteRequest is the request body for updating a note
type UpdateNoteRequest struct {
	Content string `json:"content" validate:"required,min=1,max=10000"`
}

// ToResponse converts a domain Note to API response
func ToResponse(c *note.Note) *NoteResponse {
	return &NoteResponse{
		ID:          c.ID,
		WorkspaceID: c.WorkspaceID,
		WorkflowID:  c.WorkflowID,
		UserID:      c.UserID,
		NodeID:      c.NodeID,
		Content:     c.Content,
		Resolved:    c.Resolved,
		ResolvedAt:  c.ResolvedAt,
		ResolvedBy:  c.ResolvedBy,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// ToResponseList converts a slice of domain Notes to API responses
func ToResponseList(notes []*note.Note) []*NoteResponse {
	result := make([]*NoteResponse, len(notes))
	for i, c := range notes {
		result[i] = ToResponse(c)
	}
	return result
}
