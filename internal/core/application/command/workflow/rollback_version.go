package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type RollbackVersionCommand struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Version     int
}

type RollbackVersionHandler struct {
	workflowRepo workflow.Repository
	versionRepo  workflow.VersionRepository
}

func NewRollbackVersionHandler(
	workflowRepo workflow.Repository,
	versionRepo workflow.VersionRepository,
) *RollbackVersionHandler {
	return &RollbackVersionHandler{
		workflowRepo: workflowRepo,
		versionRepo:  versionRepo,
	}
}

func (h *RollbackVersionHandler) Handle(ctx context.Context, cmd RollbackVersionCommand) (*workflow.Workflow, error) {
	wf, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, err
	}

	// Verify workspace
	if wf.WorkspaceID != cmd.WorkspaceID {
		return nil, workflow.ErrWorkflowNotFound
	}

	// Get target version
	version, err := h.versionRepo.FindByWorkflowAndVersion(ctx, cmd.WorkflowID, cmd.Version)
	if err != nil {
		return nil, workflow.ErrVersionNotFound
	}

	// Restore workflow to version state
	wf.Nodes = version.Nodes
	wf.Connections = version.Connections
	wf.Settings = version.Settings
	wf.Version = wf.Version + 1
	wf.UpdatedAt = time.Now()

	// Save updated workflow
	if err := h.workflowRepo.Update(ctx, wf); err != nil {
		return nil, err
	}

	// Create new version record
	newVersion := &workflow.Version{
		ID:            uuid.New(),
		WorkflowID:    wf.ID,
		Version:       wf.Version,
		Nodes:         wf.Nodes,
		Connections:   wf.Connections,
		Settings:      wf.Settings,
		ChangeMessage: strPtr(fmt.Sprintf("Rolled back to version %d", cmd.Version)),
		CreatedBy:     &cmd.UserID,
		CreatedAt:     time.Now(),
	}

	if err := h.versionRepo.Create(ctx, newVersion); err != nil {
		return nil, err
	}

	return wf, nil
}

func strPtr(s string) *string {
	return &s
}
