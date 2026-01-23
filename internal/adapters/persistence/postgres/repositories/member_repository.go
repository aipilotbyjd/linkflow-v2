package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// MemberModel represents workspace member in database
type MemberModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null"`
	Role        string     `gorm:"size:20;not null;default:member"`
	InvitedBy   *uuid.UUID `gorm:"type:uuid"`
	InvitedAt   *time.Time
	JoinedAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (MemberModel) TableName() string {
	return "workspace_members"
}

type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(ctx context.Context, member *workspace.Member) error {
	now := time.Now()
	model := &MemberModel{
		ID:          member.ID,
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Role:        string(member.Role),
		JoinedAt:    member.JoinedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *MemberRepository) FindByID(ctx context.Context, id uuid.UUID) (*workspace.Member, error) {
	var model MemberModel
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workspace.ErrMemberNotFound
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *MemberRepository) FindByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*workspace.Member, error) {
	var model MemberModel
	if err := postgres.GetTx(ctx, r.db).First(&model, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workspace.ErrMemberNotFound
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *MemberRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]workspace.Member, int64, error) {
	var models []MemberModel
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&MemberModel{}).Where("workspace_id = ?", workspaceID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, 0, err
	}

	members := make([]workspace.Member, len(models))
	for i, m := range models {
		members[i] = *r.toDomain(&m)
	}
	return members, total, nil
}

func (r *MemberRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]workspace.Member, error) {
	var models []MemberModel
	if err := postgres.GetTx(ctx, r.db).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}
	members := make([]workspace.Member, len(models))
	for i, m := range models {
		members[i] = *r.toDomain(&m)
	}
	return members, nil
}

func (r *MemberRepository) Update(ctx context.Context, member *workspace.Member) error {
	return postgres.GetTx(ctx, r.db).Model(&MemberModel{}).
		Where("id = ?", member.ID).
		Update("role", string(member.Role)).Error
}

func (r *MemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&MemberModel{}, "id = ?", id).Error
}

func (r *MemberRepository) DeleteByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&MemberModel{}, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error
}

func (r *MemberRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&MemberModel{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
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
	err := postgres.GetTx(ctx, r.db).Model(&MemberModel{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *MemberRepository) toDomain(m *MemberModel) *workspace.Member {
	return &workspace.Member{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Role:        workspace.Role(m.Role),
	}
}
