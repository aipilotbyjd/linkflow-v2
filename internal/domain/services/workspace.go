package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"gorm.io/gorm"
)

var (
	ErrWorkspaceNotFound    = errors.New("workspace not found")
	ErrSlugExists           = errors.New("slug already exists")
	ErrNotWorkspaceMember   = errors.New("not a member of this workspace")
	ErrInsufficientRole     = errors.New("insufficient role for this action")
	ErrCannotRemoveOwner    = errors.New("cannot remove workspace owner")
	ErrInvitationNotFound   = errors.New("invitation not found")
	ErrInvitationExpired    = errors.New("invitation expired")
)

type WorkspaceService struct {
	workspaceRepo   *repositories.WorkspaceRepository
	memberRepo      *repositories.WorkspaceMemberRepository
	invitationRepo  *repositories.WorkspaceInvitationRepository
}

func NewWorkspaceService(
	workspaceRepo *repositories.WorkspaceRepository,
	memberRepo *repositories.WorkspaceMemberRepository,
	invitationRepo *repositories.WorkspaceInvitationRepository,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo:  workspaceRepo,
		memberRepo:     memberRepo,
		invitationRepo: invitationRepo,
	}
}

type CreateWorkspaceInput struct {
	OwnerID     uuid.UUID
	Name        string
	Slug        string
	Description *string
}

func (s *WorkspaceService) Create(ctx context.Context, input CreateWorkspaceInput) (*models.Workspace, error) {
	slug := strings.ToLower(strings.ReplaceAll(input.Slug, " ", "-"))

	exists, err := s.workspaceRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSlugExists
	}

	workspace := &models.Workspace{
		OwnerID:     input.OwnerID,
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		PlanID:      "free",
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	member := &models.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      input.OwnerID,
		Role:        models.RoleOwner,
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (s *WorkspaceService) GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	return s.workspaceRepo.FindByID(ctx, id)
}

func (s *WorkspaceService) GetBySlug(ctx context.Context, slug string) (*models.Workspace, error) {
	return s.workspaceRepo.FindBySlug(ctx, slug)
}

func (s *WorkspaceService) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]models.Workspace, error) {
	return s.workspaceRepo.FindByUserID(ctx, userID)
}

type UpdateWorkspaceInput struct {
	Name        *string
	Description *string
	LogoURL     *string
	Settings    models.JSON
}

func (s *WorkspaceService) Update(ctx context.Context, workspaceID uuid.UUID, input UpdateWorkspaceInput) (*models.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
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
	if input.Settings != nil {
		workspace.Settings = input.Settings
	}

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (s *WorkspaceService) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	return s.workspaceRepo.Delete(ctx, workspaceID)
}

func (s *WorkspaceService) GetMembers(ctx context.Context, workspaceID uuid.UUID) ([]models.WorkspaceMember, error) {
	return s.memberRepo.FindByWorkspaceID(ctx, workspaceID)
}

func (s *WorkspaceService) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	return s.memberRepo.IsMember(ctx, workspaceID, userID)
}

func (s *WorkspaceService) GetMemberRole(ctx context.Context, workspaceID, userID uuid.UUID) (string, error) {
	return s.memberRepo.GetRole(ctx, workspaceID, userID)
}

func (s *WorkspaceService) HasPermission(ctx context.Context, workspaceID, userID uuid.UUID, requiredRole string) (bool, error) {
	role, err := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	roleHierarchy := map[string]int{
		models.RoleViewer: 1,
		models.RoleMember: 2,
		models.RoleAdmin:  3,
		models.RoleOwner:  4,
	}

	return roleHierarchy[role] >= roleHierarchy[requiredRole], nil
}

type InviteMemberInput struct {
	WorkspaceID uuid.UUID
	Email       string
	Role        string
	InvitedBy   uuid.UUID
}

func (s *WorkspaceService) InviteMember(ctx context.Context, input InviteMemberInput) (*models.WorkspaceInvitation, error) {
	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	invitation := &models.WorkspaceInvitation{
		WorkspaceID: input.WorkspaceID,
		Email:       input.Email,
		Role:        input.Role,
		Token:       token,
		InvitedBy:   input.InvitedBy,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *WorkspaceService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvitationNotFound
		}
		return err
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
		return err
	}

	return s.invitationRepo.MarkAccepted(ctx, invitation.ID)
}

func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	if workspace.OwnerID == userID && role != models.RoleOwner {
		return ErrCannotRemoveOwner
	}

	return s.memberRepo.UpdateRole(ctx, workspaceID, userID, role)
}

func (s *WorkspaceService) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	if workspace.OwnerID == userID {
		return ErrCannotRemoveOwner
	}

	return s.memberRepo.RemoveMember(ctx, workspaceID, userID)
}
