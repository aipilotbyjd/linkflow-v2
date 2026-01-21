package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	billingapp "github.com/linkflow-ai/linkflow/internal/core/application/billing"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

// ActivateWorkflowCommand represents the command to activate a workflow
type ActivateWorkflowCommand struct {
	WorkflowID uuid.UUID
}

// ActivateWorkflowHandler handles workflow activation
type ActivateWorkflowHandler struct {
	workflowRepo workflow.Repository
	usageService *billingapp.UsageService
	eventBus     events.Bus
}

// NewActivateWorkflowHandler creates a new handler
func NewActivateWorkflowHandler(
	workflowRepo workflow.Repository,
	usageService *billingapp.UsageService,
	eventBus events.Bus,
) *ActivateWorkflowHandler {
	return &ActivateWorkflowHandler{
		workflowRepo: workflowRepo,
		usageService: usageService,
		eventBus:     eventBus,
	}
}

// Handle executes the activate workflow command
func (h *ActivateWorkflowHandler) Handle(ctx context.Context, cmd ActivateWorkflowCommand) error {
	// Get workflow
	wf, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return workflow.ErrWorkflowNotFound
	}

	// Check active scenarios limit before activation
	if h.usageService != nil && !wf.IsActive() {
		// Count current active workflows in workspace
		activeWorkflows, _, err := h.workflowRepo.FindByWorkspaceID(ctx, wf.WorkspaceID, &workflow.ListOptions{
			Status: func() *workflow.Status { s := workflow.StatusActive; return &s }(),
		})
		if err != nil {
			return fmt.Errorf("failed to count active workflows: %w", err)
		}

		if err := h.usageService.CheckActiveScenarios(ctx, wf.WorkspaceID, len(activeWorkflows)); err != nil {
			return err
		}
	}

	// Validate and activate
	if err := wf.Activate(); err != nil {
		return err
	}

	// Save
	if err := h.workflowRepo.Update(ctx, wf); err != nil {
		return fmt.Errorf("failed to activate workflow: %w", err)
	}

	// Publish event
	if h.eventBus != nil {
		event := events.WorkflowActivated{
			BaseEvent:   events.NewBaseEvent("workflow.activated", wf.ID, "workflow"),
			WorkflowID:  wf.ID,
			WorkspaceID: wf.WorkspaceID,
			Version:     wf.Version,
		}
		_ = h.eventBus.Publish(ctx, event)
	}

	return nil
}

// DeactivateWorkflowCommand represents the command to deactivate a workflow
type DeactivateWorkflowCommand struct {
	WorkflowID uuid.UUID
}

// DeactivateWorkflowHandler handles workflow deactivation
type DeactivateWorkflowHandler struct {
	workflowRepo workflow.Repository
	eventBus     events.Bus
}

// NewDeactivateWorkflowHandler creates a new handler
func NewDeactivateWorkflowHandler(
	workflowRepo workflow.Repository,
	eventBus events.Bus,
) *DeactivateWorkflowHandler {
	return &DeactivateWorkflowHandler{
		workflowRepo: workflowRepo,
		eventBus:     eventBus,
	}
}

// Handle executes the deactivate workflow command
func (h *DeactivateWorkflowHandler) Handle(ctx context.Context, cmd DeactivateWorkflowCommand) error {
	// Get workflow
	wf, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return workflow.ErrWorkflowNotFound
	}

	// Deactivate
	wf.Deactivate()

	// Save
	if err := h.workflowRepo.Update(ctx, wf); err != nil {
		return fmt.Errorf("failed to deactivate workflow: %w", err)
	}

	// Publish event
	if h.eventBus != nil {
		event := events.WorkflowDeactivated{
			BaseEvent:   events.NewBaseEvent("workflow.deactivated", wf.ID, "workflow"),
			WorkflowID:  wf.ID,
			WorkspaceID: wf.WorkspaceID,
		}
		_ = h.eventBus.Publish(ctx, event)
	}

	return nil
}
