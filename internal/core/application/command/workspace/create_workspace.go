package workspace

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
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
	rbacRepo      rbac.Repository
	eventBus      events.Bus
}

// NewCreateWorkspaceHandler creates a new handler
func NewCreateWorkspaceHandler(
	workspaceRepo workspace.Repository,
	memberRepo workspace.MemberRepository,
	rbacRepo rbac.Repository,
	eventBus events.Bus,
) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
		rbacRepo:      rbacRepo,
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

	var ws *workspace.Workspace
	err = postgres.WithTransaction(ctx, h.workspaceRepo.(*repositories.WorkspaceRepository).DB(), func(txCtx context.Context) error {
		// Create workspace (includes validation)
		var innerErr error
		ws, innerErr = workspace.NewWorkspace(cmd.OwnerID, cmd.Name, slug)
		if innerErr != nil {
			return innerErr
		}

		ws.Description = cmd.Description
		if cmd.Timezone != "" {
			ws.Timezone = cmd.Timezone
		}
		if cmd.Language != "" {
			ws.Language = cmd.Language
		}

		if innerErr := h.workspaceRepo.Create(txCtx, ws); innerErr != nil {
			if strings.Contains(innerErr.Error(), "duplicate key") || strings.Contains(innerErr.Error(), "unique constraint") {
				return workspace.ErrSlugAlreadyExists
			}
			return fmt.Errorf("failed to create workspace: %w", innerErr)
		}

		// Add owner as member with owner role
		member, innerErr := workspace.NewMember(ws.ID, cmd.OwnerID, workspace.RoleOwner)
		if innerErr != nil {
			return fmt.Errorf("failed to create member entity: %w", innerErr)
		}

		// Fetch RBAC Owner role
		ownerRole, innerErr := h.rbacRepo.GetRoleByName(txCtx, nil, rbac.RoleOwner)
		if innerErr == nil && ownerRole != nil {
			member.RoleID = &ownerRole.ID
		}

		if innerErr := h.memberRepo.Create(txCtx, member); innerErr != nil {
			return fmt.Errorf("failed to create workspace member: %w", innerErr)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event (after transaction)
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
