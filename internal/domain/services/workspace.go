package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/rs/zerolog/log"
)

// Workspace errors
var (
	ErrWorkspaceNotFound  = errors.New("workspace not found")
	ErrSlugExists         = errors.New("slug already exists")
	ErrNotWorkspaceMember = errors.New("not a member of this workspace")
	ErrInsufficientRole   = errors.New("insufficient role for this action")
	ErrCannotRemoveOwner  = errors.New("cannot remove workspace owner")
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationExpired  = errors.New("invitation expired")
	ErrWorkspaceNameRequired = errors.New("workspace name is required")
)

// Invitation expiry duration
const InvitationExpiryDuration = 7 * 24 * time.Hour

// WorkspaceService handles workspace management operations.
type WorkspaceService struct {
	workspaceRepo  *repositories.WorkspaceRepository
	memberRepo     *repositories.WorkspaceMemberRepository
	invitationRepo *repositories.WorkspaceInvitationRepository
}

// NewWorkspaceService creates a new WorkspaceService with required dependencies.
func NewWorkspaceService(
	workspaceRepo *repositories.WorkspaceRepository,
	memberRepo *repositories.WorkspaceMemberRepository,
	invitationRepo *repositories.WorkspaceInvitationRepository,
) *WorkspaceService {
	if workspaceRepo == nil || memberRepo == nil || invitationRepo == nil {
		panic("workspace service: all repositories are required")
	}
	return &WorkspaceService{
		workspaceRepo:  workspaceRepo,
		memberRepo:     memberRepo,
		invitationRepo: invitationRepo,
	}
}

// CreateWorkspaceInput holds the input for creating a workspace.
type CreateWorkspaceInput struct {
	OwnerID      uuid.UUID
	Name         string
	Slug         string
	Description  *string
	Timezone     *string
	Language     *string
	Currency     *string
	Country      *string
	Industry     *string
	CompanySize  *string
	Website      *string
	BillingEmail *string
}

// Create creates a new workspace with the owner as the first member.
func (s *WorkspaceService) Create(ctx context.Context, input CreateWorkspaceInput) (*models.Workspace, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, ErrWorkspaceNameRequired
	}

	slug := strings.ToLower(strings.ReplaceAll(input.Slug, " ", "-"))

	exists, err := s.workspaceRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return nil, ErrSlugExists
	}

	timezone := "UTC"
	if input.Timezone != nil && *input.Timezone != "" {
		timezone = *input.Timezone
	}
	language := "en"
	if input.Language != nil && *input.Language != "" {
		language = *input.Language
	}
	currency := "USD"
	if input.Currency != nil && *input.Currency != "" {
		currency = *input.Currency
	}

	workspace := &models.Workspace{
		OwnerID:      input.OwnerID,
		Name:         input.Name,
		Slug:         slug,
		Description:  input.Description,
		Timezone:     timezone,
		Language:     language,
		Currency:     currency,
		Country:      input.Country,
		Industry:     input.Industry,
		CompanySize:  input.CompanySize,
		Website:      input.Website,
		BillingEmail: input.BillingEmail,
		PlanID:       models.PlanFree,
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	member := &models.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      input.OwnerID,
		Role:        models.RoleOwner,
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	log.Info().
		Str("workspace_id", workspace.ID.String()).
		Str("owner_id", input.OwnerID.String()).
		Str("slug", slug).
		Msg("Workspace created")

	return workspace, nil
}

// GetByID returns a workspace by its ID.
func (s *WorkspaceService) GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, id)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	return workspace, nil
}

// GetBySlug returns a workspace by its slug.
func (s *WorkspaceService) GetBySlug(ctx context.Context, slug string) (*models.Workspace, error) {
	workspace, err := s.workspaceRepo.FindBySlug(ctx, slug)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	return workspace, nil
}

// GetUserWorkspaces returns all workspaces a user is a member of.
func (s *WorkspaceService) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]models.Workspace, error) {
	workspaces, err := s.workspaceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", err)
	}
	return workspaces, nil
}

// UpdateWorkspaceInput holds the input for updating a workspace.
type UpdateWorkspaceInput struct {
	Name         *string
	Description  *string
	LogoURL      *string
	Timezone     *string
	Language     *string
	Currency     *string
	Country      *string
	Industry     *string
	CompanySize  *string
	Website      *string
	BillingEmail *string
	Settings     models.JSON
}

// Update updates a workspace's details.
func (s *WorkspaceService) Update(ctx context.Context, workspaceID uuid.UUID, input UpdateWorkspaceInput) (*models.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	if input.Name != nil {
		workspace.Name = *input.Name
	}
	if input.Description != nil {
		workspace.Description = input.Description
	}
	if input.LogoURL != nil {
		workspace.LogoURL = input.LogoURL
	}
	if input.Timezone != nil {
		workspace.Timezone = *input.Timezone
	}
	if input.Language != nil {
		workspace.Language = *input.Language
	}
	if input.Currency != nil {
		workspace.Currency = *input.Currency
	}
	if input.Country != nil {
		workspace.Country = input.Country
	}
	if input.Industry != nil {
		workspace.Industry = input.Industry
	}
	if input.CompanySize != nil {
		workspace.CompanySize = input.CompanySize
	}
	if input.Website != nil {
		workspace.Website = input.Website
	}
	if input.BillingEmail != nil {
		workspace.BillingEmail = input.BillingEmail
	}
	if input.Settings != nil {
		workspace.Settings = input.Settings
	}

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	log.Info().Str("workspace_id", workspaceID.String()).Msg("Workspace updated")

	return workspace, nil
}

// Delete deletes a workspace and all associated data.
func (s *WorkspaceService) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	if err := s.workspaceRepo.Delete(ctx, workspaceID); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	log.Info().Str("workspace_id", workspaceID.String()).Msg("Workspace deleted")
	return nil
}

// GetMembers returns all members of a workspace.
func (s *WorkspaceService) GetMembers(ctx context.Context, workspaceID uuid.UUID) ([]models.WorkspaceMember, error) {
	members, err := s.memberRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace members: %w", err)
	}
	return members, nil
}

// IsMember checks if a user is a member of a workspace.
func (s *WorkspaceService) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	isMember, err := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	return isMember, nil
}

// GetMemberRole returns the role of a user in a workspace.
func (s *WorkspaceService) GetMemberRole(ctx context.Context, workspaceID, userID uuid.UUID) (string, error) {
	role, err := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get member role: %w", err)
	}
	return role, nil
}

// HasPermission checks if a user has at least the required role in a workspace.
func (s *WorkspaceService) HasPermission(ctx context.Context, workspaceID, userID uuid.UUID, requiredRole string) (bool, error) {
	role, err := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if err != nil {
		if IsNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get member role: %w", err)
	}

	roleHierarchy := map[string]int{
		models.RoleViewer: 1,
		models.RoleMember: 2,
		models.RoleAdmin:  3,
		models.RoleOwner:  4,
	}

	return roleHierarchy[role] >= roleHierarchy[requiredRole], nil
}

// InviteMemberInput holds the input for inviting a member.
type InviteMemberInput struct {
	WorkspaceID uuid.UUID
	Email       string
	Role        string
	InvitedBy   uuid.UUID
}

// InviteMember creates an invitation for a user to join a workspace.
func (s *WorkspaceService) InviteMember(ctx context.Context, input InviteMemberInput) (*models.WorkspaceInvitation, error) {
	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	invitation := &models.WorkspaceInvitation{
		WorkspaceID: input.WorkspaceID,
		Email:       input.Email,
		Role:        input.Role,
		Token:       token,
		InvitedBy:   input.InvitedBy,
		ExpiresAt:   time.Now().Add(InvitationExpiryDuration),
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	log.Info().
		Str("workspace_id", input.WorkspaceID.String()).
		Str("email", input.Email).
		Str("role", input.Role).
		Msg("Member invited")

	return invitation, nil
}

// AcceptInvitation accepts a workspace invitation and adds the user as a member.
func (s *WorkspaceService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		if IsNotFoundError(err) {
			return ErrInvitationNotFound
		}
		return fmt.Errorf("failed to find invitation: %w", err)
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return ErrInvitationExpired
	}

	member := &models.WorkspaceMember{
		WorkspaceID: invitation.WorkspaceID,
		UserID:      userID,
		Role:        invitation.Role,
		InvitedBy:   &invitation.InvitedBy,
		InvitedAt:   &invitation.CreatedAt,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	if err := s.invitationRepo.MarkAccepted(ctx, invitation.ID); err != nil {
		log.Warn().Err(err).Str("invitation_id", invitation.ID.String()).Msg("Failed to mark invitation as accepted")
	}

	log.Info().
		Str("workspace_id", invitation.WorkspaceID.String()).
		Str("user_id", userID.String()).
		Msg("Invitation accepted")

	return nil
}

// UpdateMemberRole updates a member's role in a workspace.
func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		if IsNotFoundError(err) {
			return ErrWorkspaceNotFound
		}
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if workspace.OwnerID == userID && role != models.RoleOwner {
		return ErrCannotRemoveOwner
	}

	if err := s.memberRepo.UpdateRole(ctx, workspaceID, userID, role); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	log.Info().
		Str("workspace_id", workspaceID.String()).
		Str("user_id", userID.String()).
		Str("role", role).
		Msg("Member role updated")

	return nil
}

// RemoveMember removes a member from a workspace.
func (s *WorkspaceService) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		if IsNotFoundError(err) {
			return ErrWorkspaceNotFound
		}
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if workspace.OwnerID == userID {
		return ErrCannotRemoveOwner
	}

	if err := s.memberRepo.RemoveMember(ctx, workspaceID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	log.Info().
		Str("workspace_id", workspaceID.String()).
		Str("user_id", userID.String()).
		Msg("Member removed from workspace")

	return nil
}
