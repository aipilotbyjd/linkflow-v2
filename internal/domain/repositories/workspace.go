package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type WorkspaceRepository struct {
	*BaseRepository[models.Workspace]
}

func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{
		BaseRepository: NewBaseRepository[models.Workspace](db),
	}
}

func (r *WorkspaceRepository) FindBySlug(ctx context.Context, slug string) (*models.Workspace, error) {
	var workspace models.Workspace
	err := r.DB().WithContext(ctx).Where("slug = ?", slug).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (r *WorkspaceRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]models.Workspace, error) {
	var workspaces []models.Workspace
	err := r.DB().WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&workspaces).Error
	return workspaces, err
}

func (r *WorkspaceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Workspace, error) {
	var workspaces []models.Workspace
	err := r.DB().WithContext(ctx).
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ?", userID).
		Order("workspaces.created_at DESC").
		Find(&workspaces).Error
	return workspaces, err
}

func (r *WorkspaceRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.Workspace{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *WorkspaceRepository) UpdatePlan(ctx context.Context, workspaceID uuid.UUID, planID string) error {
	return r.DB().WithContext(ctx).Model(&models.Workspace{}).
		Where("id = ?", workspaceID).
		Update("plan_id", planID).Error
}

func (r *WorkspaceRepository) UpdateStripeCustomerID(ctx context.Context, workspaceID uuid.UUID, customerID string) error {
	return r.DB().WithContext(ctx).Model(&models.Workspace{}).
		Where("id = ?", workspaceID).
		Update("stripe_customer_id", customerID).Error
}

// Member methods
type WorkspaceMemberRepository struct {
	*BaseRepository[models.WorkspaceMember]
}

func NewWorkspaceMemberRepository(db *gorm.DB) *WorkspaceMemberRepository {
	return &WorkspaceMemberRepository{
		BaseRepository: NewBaseRepository[models.WorkspaceMember](db),
	}
}

func (r *WorkspaceMemberRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]models.WorkspaceMember, error) {
	var members []models.WorkspaceMember
	err := r.DB().WithContext(ctx).
		Preload("User").
		Where("workspace_id = ?", workspaceID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

func (r *WorkspaceMemberRepository) FindByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error) {
	var member models.WorkspaceMember
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *WorkspaceMemberRepository) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *WorkspaceMemberRepository) GetRole(ctx context.Context, workspaceID, userID uuid.UUID) (string, error) {
	var member models.WorkspaceMember
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		First(&member).Error
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

func (r *WorkspaceMemberRepository) UpdateRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	return r.DB().WithContext(ctx).Model(&models.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Update("role", role).Error
}

func (r *WorkspaceMemberRepository) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return r.DB().WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Delete(&models.WorkspaceMember{}).Error
}

func (r *WorkspaceMemberRepository) CountMembers(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.WorkspaceMember{}).
		Where("workspace_id = ?", workspaceID).
		Count(&count).Error
	return count, err
}

// Invitation methods
type WorkspaceInvitationRepository struct {
	*BaseRepository[models.WorkspaceInvitation]
}

func NewWorkspaceInvitationRepository(db *gorm.DB) *WorkspaceInvitationRepository {
	return &WorkspaceInvitationRepository{
		BaseRepository: NewBaseRepository[models.WorkspaceInvitation](db),
	}
}

func (r *WorkspaceInvitationRepository) FindByToken(ctx context.Context, token string) (*models.WorkspaceInvitation, error) {
	var invitation models.WorkspaceInvitation
	err := r.DB().WithContext(ctx).
		Preload("Workspace").
		Preload("Inviter").
		Where("token = ? AND accepted_at IS NULL", token).
		First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *WorkspaceInvitationRepository) FindPendingByEmail(ctx context.Context, email string) ([]models.WorkspaceInvitation, error) {
	var invitations []models.WorkspaceInvitation
	err := r.DB().WithContext(ctx).
		Preload("Workspace").
		Where("email = ? AND accepted_at IS NULL", email).
		Find(&invitations).Error
	return invitations, err
}

func (r *WorkspaceInvitationRepository) MarkAccepted(ctx context.Context, invitationID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.WorkspaceInvitation{}).
		Where("id = ?", invitationID).
		Update("accepted_at", gorm.Expr("NOW()")).Error
}
