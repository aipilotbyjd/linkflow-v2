package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// UpdateWorkflowCommand represents the command to update a workflow
type UpdateWorkflowCommand struct {
	WorkflowID    uuid.UUID
	UpdatedBy     uuid.UUID
	Name          *string
	Description   *string
	Nodes         types.JSONArray
	Connections   types.JSONArray
	Settings      types.JSON
	Tags          []string
	Color         *string
	Icon          *string
	Category      *string
	IsFavorite    *bool
	FolderID      *uuid.UUID
	ClearFolder   bool
	ChangeMessage *string
}

// UpdateWorkflowHandler handles workflow updates
type UpdateWorkflowHandler struct {
	workflowRepo workflow.Repository
	versionRepo  workflow.VersionRepository
}

// NewUpdateWorkflowHandler creates a new handler
func NewUpdateWorkflowHandler(
	workflowRepo workflow.Repository,
	versionRepo workflow.VersionRepository,
) *UpdateWorkflowHandler {
	return &UpdateWorkflowHandler{
		workflowRepo: workflowRepo,
		versionRepo:  versionRepo,
	}
}

// Handle executes the update workflow command
func (h *UpdateWorkflowHandler) Handle(ctx context.Context, cmd UpdateWorkflowCommand) (*workflow.Workflow, error) {
	// Get workflow
	wf, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, workflow.ErrWorkflowNotFound
	}

	// Track if content changed (for versioning)
	contentChanged := false

	// Update fields
	if cmd.Name != nil && *cmd.Name != "" {
		wf.Name = *cmd.Name
	}
	if cmd.Description != nil {
		wf.Description = cmd.Description
	}
	if cmd.Nodes != nil {
		wf.Nodes = cmd.Nodes
		contentChanged = true
	}
	if cmd.Connections != nil {
		wf.Connections = cmd.Connections
		contentChanged = true
	}
	if cmd.Settings != nil {
		wf.Settings = cmd.Settings
		contentChanged = true
	}
	if cmd.Tags != nil {
		wf.Tags = cmd.Tags
	}
	if cmd.Color != nil {
		wf.Color = cmd.Color
	}
	if cmd.Icon != nil {
		wf.Icon = cmd.Icon
	}
	if cmd.Category != nil {
		wf.Category = cmd.Category
	}
	if cmd.IsFavorite != nil {
		wf.IsFavorite = *cmd.IsFavorite
	}
	if cmd.FolderID != nil {
		wf.FolderID = cmd.FolderID
	}
	if cmd.ClearFolder {
		wf.FolderID = nil
	}

	// Increment version if content changed
	if contentChanged {
		wf.Version++
	}

	// Save
	if err := h.workflowRepo.Update(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	// Create new version if content changed
	if contentChanged {
		version := workflow.NewVersion(wf, cmd.ChangeMessage)
		version.CreatedBy = &cmd.UpdatedBy
		if err := h.versionRepo.Create(ctx, version); err != nil {
			// Non-fatal
		}
	}

	return wf, nil
}
