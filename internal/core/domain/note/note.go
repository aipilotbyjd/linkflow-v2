package note

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

var (
	ErrInvalidWorkspaceID = errors.New("invalid workspace ID")
	ErrInvalidWorkflowID  = errors.New("invalid workflow ID")
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrContentRequired    = errors.New("content is required")
	ErrContentTooLong     = errors.New("content is too long (max 10000 characters)")
	ErrNotFound           = errors.New("note not found")
)

// Repository interface
type Repository interface {
	Create(ctx context.Context, note *Note) error
	Update(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
	FindByWorkflow(ctx context.Context, workflowID uuid.UUID, opts *ListOptions) ([]*Note, int64, error)
	FindByNode(ctx context.Context, workflowID uuid.UUID, nodeID string) ([]*Note, error)
	CountByWorkflow(ctx context.Context, workflowID uuid.UUID) (int64, error)
	CountUnresolvedByWorkflow(ctx context.Context, workflowID uuid.UUID) (int64, error)
}

// ListOptions for filtering notes
type ListOptions struct {
	ListOptions  *types.ListOptions
	NodeID       *string
	ResolvedOnly bool
	UnresolvedOnly bool
}

// Note represents a note on a workflow or a specific node
type Note struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	WorkflowID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workflow_id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	NodeID      *string        `gorm:"size:100" json:"node_id,omitempty"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Resolved    bool           `gorm:"default:false" json:"resolved"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	ResolvedBy  *uuid.UUID     `gorm:"type:uuid" json:"resolved_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Note) TableName() string {
	return "notes"
}

// NewNote creates a new note
func NewNote(workspaceID, workflowID, userID uuid.UUID, content string) (*Note, error) {
	if workspaceID == uuid.Nil {
		return nil, ErrInvalidWorkspaceID
	}
	if workflowID == uuid.Nil {
		return nil, ErrInvalidWorkflowID
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	if content == "" {
		return nil, ErrContentRequired
	}
	if len(content) > 10000 {
		return nil, ErrContentTooLong
	}

	return &Note{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		UserID:      userID,
		Content:     content,
		Resolved:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// WithNodeID attaches the note to a specific node
func (c *Note) WithNodeID(nodeID string) *Note {
	c.NodeID = &nodeID
	return c
}

// Update updates the note content
func (c *Note) Update(content string) error {
	if content == "" {
		return ErrContentRequired
	}
	if len(content) > 10000 {
		return ErrContentTooLong
	}
	c.Content = content
	c.UpdatedAt = time.Now()
	return nil
}

// Resolve marks the note as resolved
func (c *Note) Resolve(resolvedBy uuid.UUID) error {
	if resolvedBy == uuid.Nil {
		return ErrInvalidUserID
	}
	now := time.Now()
	c.Resolved = true
	c.ResolvedAt = &now
	c.ResolvedBy = &resolvedBy
	c.UpdatedAt = now
	return nil
}

// Unresolve marks the note as unresolved
func (c *Note) Unresolve() {
	c.Resolved = false
	c.ResolvedAt = nil
	c.ResolvedBy = nil
	c.UpdatedAt = time.Now()
}

// IsResolved returns true if the note is resolved
func (c *Note) IsResolved() bool {
	return c.Resolved
}
