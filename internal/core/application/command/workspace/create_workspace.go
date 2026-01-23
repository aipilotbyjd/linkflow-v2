package workspace

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

// CreateWorkspaceCommand represents the command to create a workspace
type CreateWorkspaceCommand struct {
	OwnerID     uuid.UUID
	Name        string
	Slug        string
	Description *string
	Timezone    string
	Language    string
}

// CreateWorkspaceHandler handles workspace creation
type CreateWorkspaceHandler struct {
	workspaceRepo workspace.Repository
	memberRepo    workspace.MemberRepository
	eventBus      events.Bus
}

// NewCreateWorkspaceHandler creates a new handler
func NewCreateWorkspaceHandler(
	workspaceRepo workspace.Repository,
	memberRepo workspace.MemberRepository,
	eventBus events.Bus,
) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
		eventBus:      eventBus,
	}
}

// Handle executes the create workspace command
func (h *CreateWorkspaceHandler) Handle(ctx context.Context, cmd CreateWorkspaceCommand) (*workspace.Workspace, error) {
	// Generate slug if not provided
	slug := cmd.Slug
	if slug == "" {
		slug = generateSlug(cmd.Name)
	}

	// Check slug uniqueness
	exists, err := h.workspaceRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug: %w", err)
	}
	if exists {
		return nil, workspace.ErrSlugAlreadyExists
	}

	// Create workspace (includes validation)
	ws, err := workspace.NewWorkspace(cmd.OwnerID, cmd.Name, slug)
	if err != nil {
		return nil, err
	}

	ws.Description = cmd.Description
	if cmd.Timezone != "" {
		ws.Timezone = cmd.Timezone
	}
	if cmd.Language != "" {
		ws.Language = cmd.Language
	}

	if err := h.workspaceRepo.Create(ctx, ws); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, workspace.ErrSlugAlreadyExists
		}
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Add owner as member with owner role
	member, err := workspace.NewMember(ws.ID, cmd.OwnerID, workspace.RoleOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to create member entity: %w", err)
	}

	if err := h.memberRepo.Create(ctx, member); err != nil {
		// Non-fatal but should be logged
	}

	// Publish event
	if h.eventBus != nil {
		event := events.WorkspaceCreated{
			BaseEvent:   events.NewBaseEvent("workspace.created", ws.ID, "workspace"),
			WorkspaceID: ws.ID,
			OwnerID:     ws.OwnerID,
			Name:        ws.Name,
			Slug:        ws.Slug,
		}
		_ = h.eventBus.Publish(ctx, event)
	}

	return ws, nil
}

func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	// Remove consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")
	return slug
}
