package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
	ErrWorkflowInactive = errors.New("workflow is not active")
)

type WorkflowService struct {
	workflowRepo *repositories.WorkflowRepository
	versionRepo  *repositories.WorkflowVersionRepository
}

func NewWorkflowService(
	workflowRepo *repositories.WorkflowRepository,
	versionRepo *repositories.WorkflowVersionRepository,
) *WorkflowService {
	return &WorkflowService{
		workflowRepo: workflowRepo,
		versionRepo:  versionRepo,
	}
}

type CreateWorkflowInput struct {
	WorkspaceID uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Description *string
	Nodes       models.JSON
	Connections models.JSON
	Settings    models.JSON
	Tags        []string
}

func (s *WorkflowService) Create(ctx context.Context, input CreateWorkflowInput) (*models.Workflow, error) {
	workflow := &models.Workflow{
		WorkspaceID: input.WorkspaceID,
		CreatedBy:   input.CreatedBy,
		Name:        input.Name,
		Description: input.Description,
		Status:      models.WorkflowStatusDraft,
		Version:     1,
		Nodes:       input.Nodes,
		Connections: input.Connections,
		Settings:    input.Settings,
		Tags:        input.Tags,
	}

	if err := s.workflowRepo.Create(ctx, workflow); err != nil {
		return nil, err
	}

	version := &models.WorkflowVersion{
		WorkflowID:  workflow.ID,
		Version:     1,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		CreatedBy:   &input.CreatedBy,
	}
	s.versionRepo.Create(ctx, version)

	return workflow, nil
}

func (s *WorkflowService) GetByID(ctx context.Context, id uuid.UUID) (*models.Workflow, error) {
	return s.workflowRepo.FindByID(ctx, id)
}

func (s *WorkflowService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Workflow, int64, error) {
	return s.workflowRepo.FindByWorkspaceID(ctx, workspaceID, opts)
}

func (s *WorkflowService) Search(ctx context.Context, workspaceID uuid.UUID, query string, opts *repositories.ListOptions) ([]models.Workflow, int64, error) {
	return s.workflowRepo.Search(ctx, workspaceID, query, opts)
}

type UpdateWorkflowInput struct {
	Name        *string
	Description *string
	Nodes       models.JSON
	Connections models.JSON
	Settings    models.JSON
	Tags        []string
}

func (s *WorkflowService) Update(ctx context.Context, workflowID uuid.UUID, input UpdateWorkflowInput, userID uuid.UUID) (*models.Workflow, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		workflow.Name = *input.Name
	}
	if input.Description != nil {
		workflow.Description = input.Description
	}
	if input.Nodes != nil {
		workflow.Nodes = input.Nodes
	}
	if input.Connections != nil {
		workflow.Connections = input.Connections
	}
	if input.Settings != nil {
		workflow.Settings = input.Settings
	}
	if input.Tags != nil {
		workflow.Tags = input.Tags
	}

	workflow.Version++

	if err := s.workflowRepo.Update(ctx, workflow); err != nil {
		return nil, err
	}

	version := &models.WorkflowVersion{
		WorkflowID:  workflow.ID,
		Version:     workflow.Version,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		CreatedBy:   &userID,
	}
	s.versionRepo.Create(ctx, version)

	return workflow, nil
}

func (s *WorkflowService) Delete(ctx context.Context, workflowID uuid.UUID) error {
	return s.workflowRepo.Delete(ctx, workflowID)
}

func (s *WorkflowService) Activate(ctx context.Context, workflowID uuid.UUID) error {
	return s.workflowRepo.UpdateStatus(ctx, workflowID, models.WorkflowStatusActive)
}

func (s *WorkflowService) Deactivate(ctx context.Context, workflowID uuid.UUID) error {
	return s.workflowRepo.UpdateStatus(ctx, workflowID, models.WorkflowStatusInactive)
}

func (s *WorkflowService) Archive(ctx context.Context, workflowID uuid.UUID) error {
	return s.workflowRepo.UpdateStatus(ctx, workflowID, models.WorkflowStatusArchived)
}

func (s *WorkflowService) Clone(ctx context.Context, workflowID uuid.UUID, userID uuid.UUID, newName string) (*models.Workflow, error) {
	original, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	cloned := &models.Workflow{
		WorkspaceID: original.WorkspaceID,
		CreatedBy:   userID,
		Name:        newName,
		Description: original.Description,
		Status:      models.WorkflowStatusDraft,
		Version:     1,
		Nodes:       original.Nodes,
		Connections: original.Connections,
		Settings:    original.Settings,
		Tags:        original.Tags,
	}

	if err := s.workflowRepo.Create(ctx, cloned); err != nil {
		return nil, err
	}

	return cloned, nil
}

func (s *WorkflowService) GetVersions(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowVersion, error) {
	return s.versionRepo.FindByWorkflowID(ctx, workflowID)
}

func (s *WorkflowService) GetVersion(ctx context.Context, workflowID uuid.UUID, version int) (*models.WorkflowVersion, error) {
	return s.versionRepo.FindByWorkflowAndVersion(ctx, workflowID, version)
}

func (s *WorkflowService) RestoreVersion(ctx context.Context, workflowID uuid.UUID, version int, userID uuid.UUID) (*models.Workflow, error) {
	v, err := s.versionRepo.FindByWorkflowAndVersion(ctx, workflowID, version)
	if err != nil {
		return nil, err
	}

	return s.Update(ctx, workflowID, UpdateWorkflowInput{
		Nodes:       v.Nodes,
		Connections: v.Connections,
		Settings:    v.Settings,
	}, userID)
}

// WebhookEndpoint represents a webhook endpoint for a workflow
type WebhookEndpoint struct {
	WorkflowID      uuid.UUID
	EndpointID      string
	Secret          string
	ResponseMode    string // "immediate" or "wait"
	ResponseTimeout int    // timeout in seconds for wait mode
}

func (s *WorkflowService) GetWebhookByEndpoint(ctx context.Context, endpointID string) (*WebhookEndpoint, error) {
	// Look for workflow with matching webhook endpoint in settings
	workflows, err := s.workflowRepo.FindActiveWithWebhook(ctx, endpointID)
	if err != nil {
		return nil, err
	}
	if len(workflows) == 0 {
		return nil, ErrWorkflowNotFound
	}

	workflow := workflows[0]
	
	// Extract webhook settings from workflow
	settings := workflow.Settings
	var secret, responseMode string
	
	if settings != nil {
		if webhookSettings, ok := settings["webhook"].(map[string]interface{}); ok {
			if s, ok := webhookSettings["secret"].(string); ok {
				secret = s
			}
			if rm, ok := webhookSettings["responseMode"].(string); ok {
				responseMode = rm
			}
		}
	}
	
	if responseMode == "" {
		responseMode = "immediate"
	}

	return &WebhookEndpoint{
		WorkflowID:   workflow.ID,
		EndpointID:   endpointID,
		Secret:       secret,
		ResponseMode: responseMode,
	}, nil
}
