package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/rs/zerolog/log"
)

// Workflow errors
var (
	ErrWorkflowNotFound     = errors.New("workflow not found")
	ErrWorkflowInactive     = errors.New("workflow is not active")
	ErrWorkflowNameRequired = errors.New("workflow name is required")
	ErrVersionNotFound      = errors.New("workflow version not found")
)

// Webhook response mode constants
const (
	WebhookResponseModeImmediate = "immediate"
	WebhookResponseModeWait      = "wait"
)

type WorkflowService struct {
	workflowRepo        *repositories.WorkflowRepository
	versionRepo         *repositories.WorkflowVersionRepository
	webhookEndpointRepo *repositories.WebhookEndpointRepository
}

// NewWorkflowService creates a new WorkflowService with required repositories.
func NewWorkflowService(
	workflowRepo *repositories.WorkflowRepository,
	versionRepo *repositories.WorkflowVersionRepository,
) *WorkflowService {
	if workflowRepo == nil || versionRepo == nil {
		panic("workflow service: workflowRepo and versionRepo are required")
	}
	return &WorkflowService{
		workflowRepo: workflowRepo,
		versionRepo:  versionRepo,
	}
}

// SetWebhookEndpointRepo sets the webhook endpoint repository (optional dependency)
func (s *WorkflowService) SetWebhookEndpointRepo(repo *repositories.WebhookEndpointRepository) {
	s.webhookEndpointRepo = repo
}

type CreateWorkflowInput struct {
	WorkspaceID uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Description *string
	Nodes       models.JSONArray
	Connections models.JSONArray
	Settings    models.JSON
	Tags        []string
}

// Create creates a new workflow with an initial version.
func (s *WorkflowService) Create(ctx context.Context, input CreateWorkflowInput) (*models.Workflow, error) {
	// Validate input
	if input.Name == "" {
		return nil, ErrWorkflowNameRequired
	}

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
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Create initial version
	version := &models.WorkflowVersion{
		WorkflowID:  workflow.ID,
		Version:     1,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		CreatedBy:   &input.CreatedBy,
	}
	if err := s.versionRepo.Create(ctx, version); err != nil {
		log.Error().
			Err(err).
			Str("workflow_id", workflow.ID.String()).
			Msg("Failed to create initial workflow version")
	}

	log.Info().
		Str("workflow_id", workflow.ID.String()).
		Str("workspace_id", input.WorkspaceID.String()).
		Str("name", input.Name).
		Msg("Workflow created")

	return workflow, nil
}

// GetByID returns a workflow by its ID.
func (s *WorkflowService) GetByID(ctx context.Context, id uuid.UUID) (*models.Workflow, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrWorkflowNotFound, id)
	}
	return workflow, nil
}

// GetByWorkspace returns paginated workflows for a workspace.
func (s *WorkflowService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Workflow, int64, error) {
	workflows, total, err := s.workflowRepo.FindByWorkspaceID(ctx, workspaceID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get workflows: %w", err)
	}
	return workflows, total, nil
}

// Search searches workflows by name/description in a workspace.
func (s *WorkflowService) Search(ctx context.Context, workspaceID uuid.UUID, query string, opts *repositories.ListOptions) ([]models.Workflow, int64, error) {
	workflows, total, err := s.workflowRepo.Search(ctx, workspaceID, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search workflows: %w", err)
	}
	return workflows, total, nil
}

type UpdateWorkflowInput struct {
	Name        *string
	Description *string
	Nodes       models.JSONArray
	Connections models.JSONArray
	Settings    models.JSON
	Tags        []string
}

// Update updates a workflow and creates a new version.
func (s *WorkflowService) Update(ctx context.Context, workflowID uuid.UUID, input UpdateWorkflowInput, userID uuid.UUID) (*models.Workflow, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrWorkflowNotFound, workflowID)
	}

	// Validate name if provided
	if input.Name != nil && *input.Name == "" {
		return nil, ErrWorkflowNameRequired
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
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	// Create new version
	version := &models.WorkflowVersion{
		WorkflowID:  workflow.ID,
		Version:     workflow.Version,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		CreatedBy:   &userID,
	}
	if err := s.versionRepo.Create(ctx, version); err != nil {
		log.Error().
			Err(err).
			Str("workflow_id", workflow.ID.String()).
			Int("version", workflow.Version).
			Msg("Failed to create workflow version")
	}

	log.Info().
		Str("workflow_id", workflow.ID.String()).
		Int("version", workflow.Version).
		Msg("Workflow updated")

	return workflow, nil
}

// Delete deletes a workflow and all its versions.
func (s *WorkflowService) Delete(ctx context.Context, workflowID uuid.UUID) error {
	if err := s.workflowRepo.Delete(ctx, workflowID); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}
	log.Info().Str("workflow_id", workflowID.String()).Msg("Workflow deleted")
	return nil
}

// Activate activates a workflow, making it available for execution.
func (s *WorkflowService) Activate(ctx context.Context, workflowID uuid.UUID) error {
	if err := s.workflowRepo.UpdateStatus(ctx, workflowID, models.WorkflowStatusActive); err != nil {
		return fmt.Errorf("failed to activate workflow: %w", err)
	}
	log.Info().Str("workflow_id", workflowID.String()).Msg("Workflow activated")
	return nil
}

// Deactivate deactivates a workflow, preventing it from being executed.
func (s *WorkflowService) Deactivate(ctx context.Context, workflowID uuid.UUID) error {
	if err := s.workflowRepo.UpdateStatus(ctx, workflowID, models.WorkflowStatusInactive); err != nil {
		return fmt.Errorf("failed to deactivate workflow: %w", err)
	}
	log.Info().Str("workflow_id", workflowID.String()).Msg("Workflow deactivated")
	return nil
}

// Archive archives a workflow.
func (s *WorkflowService) Archive(ctx context.Context, workflowID uuid.UUID) error {
	if err := s.workflowRepo.UpdateStatus(ctx, workflowID, models.WorkflowStatusArchived); err != nil {
		return fmt.Errorf("failed to archive workflow: %w", err)
	}
	log.Info().Str("workflow_id", workflowID.String()).Msg("Workflow archived")
	return nil
}

// Clone creates a copy of an existing workflow with a new name.
func (s *WorkflowService) Clone(ctx context.Context, workflowID uuid.UUID, userID uuid.UUID, newName string) (*models.Workflow, error) {
	if newName == "" {
		return nil, ErrWorkflowNameRequired
	}

	original, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrWorkflowNotFound, workflowID)
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
		return nil, fmt.Errorf("failed to clone workflow: %w", err)
	}

	log.Info().
		Str("original_id", workflowID.String()).
		Str("cloned_id", cloned.ID.String()).
		Str("name", newName).
		Msg("Workflow cloned")

	return cloned, nil
}

// GetVersions returns all versions of a workflow.
func (s *WorkflowService) GetVersions(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowVersion, error) {
	versions, err := s.versionRepo.FindByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow versions: %w", err)
	}
	return versions, nil
}

// GetVersion returns a specific version of a workflow.
func (s *WorkflowService) GetVersion(ctx context.Context, workflowID uuid.UUID, version int) (*models.WorkflowVersion, error) {
	v, err := s.versionRepo.FindByWorkflowAndVersion(ctx, workflowID, version)
	if err != nil {
		return nil, fmt.Errorf("%w: workflow %s version %d", ErrVersionNotFound, workflowID, version)
	}
	return v, nil
}

// RestoreVersion restores a workflow to a previous version.
func (s *WorkflowService) RestoreVersion(ctx context.Context, workflowID uuid.UUID, version int, userID uuid.UUID) (*models.Workflow, error) {
	v, err := s.versionRepo.FindByWorkflowAndVersion(ctx, workflowID, version)
	if err != nil {
		return nil, fmt.Errorf("%w: workflow %s version %d", ErrVersionNotFound, workflowID, version)
	}

	workflow, err := s.Update(ctx, workflowID, UpdateWorkflowInput{
		Nodes:       v.Nodes,
		Connections: v.Connections,
		Settings:    v.Settings,
	}, userID)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("workflow_id", workflowID.String()).
		Int("restored_version", version).
		Int("new_version", workflow.Version).
		Msg("Workflow version restored")

	return workflow, nil
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
	// First, check the webhook_endpoints table (new webhook management system)
	if s.webhookEndpointRepo != nil {
		endpoint, err := s.webhookEndpointRepo.FindByPath(ctx, endpointID)
		if err == nil && endpoint != nil && endpoint.IsActive {
			// Verify the workflow is active
			workflow, err := s.workflowRepo.FindByID(ctx, endpoint.WorkflowID)
			if err == nil && workflow.Status == models.WorkflowStatusActive {
				secret := ""
				if endpoint.Secret != nil {
					secret = *endpoint.Secret
				}
				return &WebhookEndpoint{
					WorkflowID:   endpoint.WorkflowID,
					EndpointID:   endpointID,
					Secret:       secret,
					ResponseMode: WebhookResponseModeImmediate,
				}, nil
			}
		}
	}

	// Fallback: Look for workflow with matching webhook endpoint in settings
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
		responseMode = WebhookResponseModeImmediate
	}

	return &WebhookEndpoint{
		WorkflowID:   workflow.ID,
		EndpointID:   endpointID,
		Secret:       secret,
		ResponseMode: responseMode,
	}, nil
}
