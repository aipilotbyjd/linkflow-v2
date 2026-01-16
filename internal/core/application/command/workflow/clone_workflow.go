package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type CloneWorkflowCommand struct {
	SourceWorkflowID uuid.UUID
	WorkspaceID      uuid.UUID
	UserID           uuid.UUID
	NewName          string
}

type CloneWorkflowHandler struct {
	workflowRepo workflow.Repository
}

func NewCloneWorkflowHandler(workflowRepo workflow.Repository) *CloneWorkflowHandler {
	return &CloneWorkflowHandler{workflowRepo: workflowRepo}
}

func (h *CloneWorkflowHandler) Handle(ctx context.Context, cmd CloneWorkflowCommand) (*workflow.Workflow, error) {
	source, err := h.workflowRepo.FindByID(ctx, cmd.SourceWorkflowID)
	if err != nil {
		return nil, err
	}

	// Generate name if not provided
	name := cmd.NewName
	if name == "" {
		name = fmt.Sprintf("%s (Copy)", source.Name)
	}

	// Create clone
	clone := &workflow.Workflow{
		ID:          uuid.New(),
		WorkspaceID: cmd.WorkspaceID,
		Name:        name,
		Description: source.Description,
		Status:      workflow.StatusDraft,
		Version:     1,
		Nodes:       source.Nodes,
		Connections: source.Connections,
		Settings:    source.Settings,
		CreatedBy:   cmd.UserID,
	}

	if err := h.workflowRepo.Create(ctx, clone); err != nil {
		return nil, err
	}

	return clone, nil
}
