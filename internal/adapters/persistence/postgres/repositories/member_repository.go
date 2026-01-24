package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/mappers"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// MemberModel definition removed in favor of models.WorkspaceMember
type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(ctx context.Context, member *workspace.Member) error {
	now := time.Now()
	model := &models.WorkspaceMember{
		ID:          member.ID,
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Role:        string(member.Role),
		RoleID:      member.RoleID,
		JoinedAt:    member.JoinedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *MemberRepository) FindByID(ctx context.Context, id uuid.UUID) (*workspace.Member, error) {
	var model models.WorkspaceMember
	if err := postgres.GetTx(ctx, r.db).Preload("RoleRef.Permissions").First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workspace.ErrMemberNotFound
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *MemberRepository) FindByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*workspace.Member, error) {
	var model models.WorkspaceMember
	if err := postgres.GetTx(ctx, r.db).Preload("RoleRef.Permissions").First(&model, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workspace.ErrMemberNotFound
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *MemberRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]workspace.Member, int64, error) {
	var models []models.WorkspaceMember
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models).Where("workspace_id = ?", workspaceID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	// Preload RBAC Role and Permissions
	if err := query.Preload("RoleRef.Permissions").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	members := make([]workspace.Member, len(models))
	for i, m := range models {
		members[i] = *r.toDomain(&m)
	}
	return members, total, nil
}

func (r *MemberRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]workspace.Member, error) {
	var models []models.WorkspaceMember
	if err := postgres.GetTx(ctx, r.db).Preload("RoleRef.Permissions").Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}
	members := make([]workspace.Member, len(models))
	for i, m := range models {
		members[i] = *r.toDomain(&m)
	}
	return members, nil
}

func (r *MemberRepository) Update(ctx context.Context, member *workspace.Member) error {
	updates := map[string]interface{}{
		"role": string(member.Role),
	}
	if member.RoleID != nil {
		updates["role_id"] = member.RoleID
	}

	return postgres.GetTx(ctx, r.db).Model(&models.WorkspaceMember{}).
		Where("id = ?", member.ID).
		Updates(updates).Error
}

func (r *MemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.WorkspaceMember{}, "id = ?", id).Error
}

func (r *MemberRepository) DeleteByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.WorkspaceMember{}, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error
}

func (r *MemberRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.WorkspaceMember{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

func (r *MemberRepository) FindWorkspacesByUserID(ctx context.Context, userID uuid.UUID) ([]workspace.Workspace, error) {
	var results []struct {
		ID          uuid.UUID
		Name        string
		Slug        string
		Description *string
		OwnerID     uuid.UUID
		PlanID      string
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	err := postgres.GetTx(ctx, r.db).
		Table("workspaces").
		Select("workspaces.id, workspaces.name, workspaces.slug, workspaces.description, workspaces.owner_id, workspaces.plan_id, workspaces.created_at, workspaces.updated_at").
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ? AND workspace_members.deleted_at IS NULL AND workspaces.deleted_at IS NULL", userID).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	workspaces := make([]workspace.Workspace, len(results))
	for i, r := range results {
		workspaces[i] = workspace.Workspace{
			ID:          r.ID,
			Name:        r.Name,
			Slug:        r.Slug,
			Description: r.Description,
			OwnerID:     r.OwnerID,
			PlanID:      r.PlanID,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		}
	}
	return workspaces, nil
}

func (r *MemberRepository) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *MemberRepository) toDomain(m *models.WorkspaceMember) *workspace.Member {
	member := &workspace.Member{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Role:        workspace.Role(m.Role),
		RoleID:      m.RoleID,
		InvitedBy:   m.InvitedBy,
		InvitedAt:   m.InvitedAt,
		JoinedAt:    m.JoinedAt,
		CreatedAt:   m.CreatedAt,
	}
	if m.RoleRef != nil {
		member.RBACRole = mappers.ToDomainRole(m.RoleRef)
	}
	return member
}
