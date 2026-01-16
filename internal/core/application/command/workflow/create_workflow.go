package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// CreateWorkflowCommand represents the command to create a workflow
type CreateWorkflowCommand struct {
	WorkspaceID uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Description *string
	Nodes       types.JSONArray
	Connections types.JSONArray
	Settings    types.JSON
	Tags        []string
	Color       *string
	Icon        *string
	Category    *string
}

// CreateWorkflowHandler handles workflow creation
type CreateWorkflowHandler struct {
	workflowRepo workflow.Repository
	versionRepo  workflow.VersionRepository
	eventBus     events.Bus
}

// NewCreateWorkflowHandler creates a new handler
func NewCreateWorkflowHandler(
	workflowRepo workflow.Repository,
	versionRepo workflow.VersionRepository,
	eventBus events.Bus,
) *CreateWorkflowHandler {
	return &CreateWorkflowHandler{
		workflowRepo: workflowRepo,
		versionRepo:  versionRepo,
		eventBus:     eventBus,
	}
}

// Handle executes the create workflow command
func (h *CreateWorkflowHandler) Handle(ctx context.Context, cmd CreateWorkflowCommand) (*workflow.Workflow, error) {
	// Validate
	if cmd.Name == "" {
		return nil, workflow.ErrWorkflowNameExists // Should be validation error
	}

	// Check name uniqueness
	exists, err := h.workflowRepo.ExistsByName(ctx, cmd.WorkspaceID, cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check workflow name: %w", err)
	}
	if exists {
		return nil, workflow.ErrWorkflowNameExists
	}

	// Create workflow
	wf := workflow.NewWorkflow(cmd.WorkspaceID, cmd.CreatedBy, cmd.Name)
	wf.Description = cmd.Description
	wf.Nodes = cmd.Nodes
	wf.Connections = cmd.Connections
	wf.Settings = cmd.Settings
	wf.Tags = cmd.Tags
	wf.Color = cmd.Color
	wf.Icon = cmd.Icon
	wf.Category = cmd.Category

	if err := h.workflowRepo.Create(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Create initial version
	version := workflow.NewVersion(wf, nil)
	if err := h.versionRepo.Create(ctx, version); err != nil {
		// Non-fatal
	}

	// Publish event
	if h.eventBus != nil {
		event := events.WorkflowCreated{
			BaseEvent:   events.NewBaseEvent("workflow.created", wf.ID, "workflow"),
			WorkflowID:  wf.ID,
			WorkspaceID: wf.WorkspaceID,
			Name:        wf.Name,
			CreatedBy:   wf.CreatedBy,
		}
		_ = h.eventBus.Publish(ctx, event)
	}

	return wf, nil
}
