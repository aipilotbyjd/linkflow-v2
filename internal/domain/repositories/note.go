package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type NoteRepository struct {
	*BaseRepository[models.Note]
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{
		BaseRepository: NewBaseRepository[models.Note](db),
	}
}

// FindByWorkspaceID returns all notes in a workspace
func (r *NoteRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]models.Note, int64, error) {
	var notes []models.Note
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	query.Model(&models.Note{}).Count(&total)

	if opts != nil {
		if opts.OrderBy != "" {
			query = query.Order(opts.OrderBy + " " + opts.Order)
		}
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	err := query.Find(&notes).Error
	return notes, total, err
}

// FindByResource returns all notes for a specific resource
func (r *NoteRepository) FindByResource(ctx context.Context, workspaceID, resourceID uuid.UUID, resourceName string, opts *ListOptions) ([]models.Note, int64, error) {
	var notes []models.Note
	var total int64

	query := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND resource_id = ? AND resource_name = ?", workspaceID, resourceID, resourceName)
	query.Model(&models.Note{}).Count(&total)

	if opts != nil {
		if opts.OrderBy != "" {
			query = query.Order(opts.OrderBy + " " + opts.Order)
		}
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	err := query.Find(&notes).Error
	return notes, total, err
}

// NoteFilter contains filter parameters for note queries
type NoteFilter struct {
	ResourceID   *uuid.UUID
	ResourceName *string
	Color        *string
	Search       *string
	SortBy       string
	Order        string
}

// FindWithFilters returns notes matching the given filters
func (r *NoteRepository) FindWithFilters(ctx context.Context, workspaceID uuid.UUID, filter *NoteFilter, opts *ListOptions) ([]models.Note, int64, error) {
	var notes []models.Note
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)

	// Apply filters
	if filter != nil {
		if filter.ResourceID != nil {
			query = query.Where("resource_id = ?", *filter.ResourceID)
		}
		if filter.ResourceName != nil && *filter.ResourceName != "" {
			query = query.Where("resource_name = ?", *filter.ResourceName)
		}
		if filter.Color != nil && *filter.Color != "" {
			query = query.Where("color = ?", *filter.Color)
		}
		if filter.Search != nil && *filter.Search != "" {
			searchPattern := "%" + *filter.Search + "%"
			query = query.Where("content ILIKE ?", searchPattern)
		}
	}

	// Count total before pagination
	query.Model(&models.Note{}).Count(&total)

	// Apply sorting
	sortBy := "created_at"
	order := "desc"
	if filter != nil {
		if filter.SortBy != "" {
			sortBy = filter.SortBy
		}
		if filter.Order != "" {
			order = filter.Order
		}
	}
	query = query.Order(sortBy + " " + order)

	// Apply pagination
	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	err := query.Find(&notes).Error
	return notes, total, err
}

// CountByResource returns the count of notes for a specific resource
func (r *NoteRepository) CountByResource(ctx context.Context, workspaceID, resourceID uuid.UUID, resourceName string) (int64, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.Note{}).
		Where("workspace_id = ? AND resource_id = ? AND resource_name = ?", workspaceID, resourceID, resourceName).
		Count(&count).Error
	return count, err
}

// DeleteByResource deletes all notes for a specific resource
func (r *NoteRepository) DeleteByResource(ctx context.Context, workspaceID, resourceID uuid.UUID, resourceName string) error {
	return r.DB().WithContext(ctx).
		Where("workspace_id = ? AND resource_id = ? AND resource_name = ?", workspaceID, resourceID, resourceName).
		Delete(&models.Note{}).Error
}
