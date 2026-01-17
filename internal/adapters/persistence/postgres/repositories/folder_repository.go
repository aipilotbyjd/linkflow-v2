package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/folder"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type FolderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) *FolderRepository {
	return &FolderRepository{db: db}
}

func (r *FolderRepository) Create(ctx context.Context, f *folder.Folder) error {
	model := folderToModel(f)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *FolderRepository) Update(ctx context.Context, f *folder.Folder) error {
	model := folderToModel(f)
	return postgres.GetTx(ctx, r.db).Save(model).Error
}

func (r *FolderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.Folder{}, "id = ?", id).Error
}

func (r *FolderRepository) FindByID(ctx context.Context, id uuid.UUID) (*folder.Folder, error) {
	var model models.Folder
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, folder.ErrFolderNotFound
		}
		return nil, err
	}
	return folderToDomain(&model), nil
}

func (r *FolderRepository) FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]*folder.Folder, int64, error) {
	var modelList []models.Folder
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Folder{}).Where("workspace_id = ?", workspaceID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	query = query.Order("position ASC, name ASC")

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	folders := make([]*folder.Folder, len(modelList))
	for i, m := range modelList {
		folders[i] = folderToDomain(&m)
	}

	return folders, total, nil
}

func (r *FolderRepository) FindByParent(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID) ([]*folder.Folder, error) {
	var modelList []models.Folder

	query := postgres.GetTx(ctx, r.db).Model(&models.Folder{}).Where("workspace_id = ?", workspaceID)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	query = query.Order("position ASC, name ASC")

	if err := query.Find(&modelList).Error; err != nil {
		return nil, err
	}

	folders := make([]*folder.Folder, len(modelList))
	for i, m := range modelList {
		folders[i] = folderToDomain(&m)
	}

	return folders, nil
}

func (r *FolderRepository) FindRootFolders(ctx context.Context, workspaceID uuid.UUID) ([]*folder.Folder, error) {
	return r.FindByParent(ctx, workspaceID, nil)
}

func (r *FolderRepository) HasChildren(ctx context.Context, folderID uuid.UUID) (bool, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Folder{}).
		Where("parent_id = ?", folderID).
		Count(&count).Error
	return count > 0, err
}

func (r *FolderRepository) HasWorkflows(ctx context.Context, folderID uuid.UUID) (bool, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("folder_id = ?", folderID).
		Count(&count).Error
	return count > 0, err
}

// Mapper functions
func folderToModel(f *folder.Folder) *models.Folder {
	return &models.Folder{
		ID:          f.ID,
		WorkspaceID: f.WorkspaceID,
		ParentID:    f.ParentID,
		Name:        f.Name,
		Description: f.Description,
		Color:       f.Color,
		Position:    f.Position,
		CreatedBy:   f.CreatedBy,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

func folderToDomain(m *models.Folder) *folder.Folder {
	return &folder.Folder{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		ParentID:    m.ParentID,
		Name:        m.Name,
		Description: m.Description,
		Color:       m.Color,
		Position:    m.Position,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
