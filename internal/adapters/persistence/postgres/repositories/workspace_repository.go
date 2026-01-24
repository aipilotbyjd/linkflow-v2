package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/mappers"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type WorkspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) DB() *gorm.DB {
	return r.db
}

func (r *WorkspaceRepository) Create(ctx context.Context, ws *workspace.Workspace) error {
	model := mappers.WorkspaceToModel(ws)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *WorkspaceRepository) Update(ctx context.Context, ws *workspace.Workspace) error {
	model := mappers.WorkspaceToModel(ws)
	return postgres.GetTx(ctx, r.db).Save(model).Error
}

func (r *WorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.Workspace{}, "id = ?", id).Error
}

func (r *WorkspaceRepository) FindByID(ctx context.Context, id uuid.UUID) (*workspace.Workspace, error) {
	var model models.Workspace
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workspace.ErrWorkspaceNotFound
		}
		return nil, err
	}
	return mappers.WorkspaceToDomain(&model), nil
}

func (r *WorkspaceRepository) FindBySlug(ctx context.Context, slug string) (*workspace.Workspace, error) {
	var model models.Workspace
	if err := postgres.GetTx(ctx, r.db).First(&model, "slug = ?", slug).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workspace.ErrWorkspaceNotFound
		}
		return nil, err
	}
	return mappers.WorkspaceToDomain(&model), nil
}

func (r *WorkspaceRepository) FindByUserID(ctx context.Context, userID uuid.UUID, opts *types.ListOptions) ([]workspace.Workspace, int64, error) {
	var modelList []models.Workspace
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Workspace{}).
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ? AND workspace_members.deleted_at IS NULL", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	workspaces := make([]workspace.Workspace, len(modelList))
	for i, m := range modelList {
		workspaces[i] = *mappers.WorkspaceToDomain(&m)
	}

	return workspaces, total, nil
}

func (r *WorkspaceRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Workspace{}).Where("slug = ? AND deleted_at IS NULL", slug).Count(&count).Error
	return count > 0, err
}

func (r *WorkspaceRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]workspace.Workspace, error) {
	var modelList []models.Workspace
	if err := postgres.GetTx(ctx, r.db).Where("owner_id = ?", ownerID).Find(&modelList).Error; err != nil {
		return nil, err
	}
	workspaces := make([]workspace.Workspace, len(modelList))
	for i, m := range modelList {
		workspaces[i] = *mappers.WorkspaceToDomain(&m)
	}
	return workspaces, nil
}

func (r *WorkspaceRepository) FindAll(ctx context.Context, opts *types.ListOptions) ([]workspace.Workspace, int64, error) {
	var modelList []models.Workspace
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Workspace{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	workspaces := make([]workspace.Workspace, len(modelList))
	for i, m := range modelList {
		workspaces[i] = *mappers.WorkspaceToDomain(&m)
	}
	return workspaces, total, nil
}
