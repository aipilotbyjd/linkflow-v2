package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// MemberModel represents workspace member in database
type MemberModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null"`
	Role        string    `gorm:"size:20;not null;default:member"`
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
	model := &MemberModel{
		ID:          member.ID,
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Role:        string(member.Role),
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
	// This requires joining with workspaces table - simplified for now
	return nil, nil
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
