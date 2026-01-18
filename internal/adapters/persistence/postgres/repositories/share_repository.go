package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/share"
	"gorm.io/gorm"
)

// ShareRepository implements share.Repository
type ShareRepository struct {
	db *gorm.DB
}

// NewShareRepository creates a new share repository
func NewShareRepository(db *gorm.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

// Create creates a new share
func (r *ShareRepository) Create(ctx context.Context, s *share.Share) error {
	model := toShareModel(s)
	return r.db.WithContext(ctx).Create(&model).Error
}

// FindByID finds a share by ID
func (r *ShareRepository) FindByID(ctx context.Context, id uuid.UUID) (*share.Share, error) {
	var model models.Share
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	result := toShareDomain(model)
	return &result, nil
}

// Update updates a share
func (r *ShareRepository) Update(ctx context.Context, s *share.Share) error {
	model := toShareModel(s)
	return r.db.WithContext(ctx).Save(&model).Error
}

// Delete deletes a share
func (r *ShareRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Share{}, "id = ?", id).Error
}

// FindSharedByUser finds all shares created by a user
func (r *ShareRepository) FindSharedByUser(ctx context.Context, userID uuid.UUID) ([]share.Share, error) {
	var modelList []models.Share
	if err := r.db.WithContext(ctx).Where("shared_by_id = ?", userID).Order("created_at DESC").Find(&modelList).Error; err != nil {
		return nil, err
	}
	return toShareDomainList(modelList), nil
}

// FindSharedWithUser finds all shares shared with a user
func (r *ShareRepository) FindSharedWithUser(ctx context.Context, userID uuid.UUID) ([]share.Share, error) {
	var modelList []models.Share
	if err := r.db.WithContext(ctx).Where("shared_with_id = ? AND status = ?", userID, "accepted").Order("created_at DESC").Find(&modelList).Error; err != nil {
		return nil, err
	}
	return toShareDomainList(modelList), nil
}

// FindPendingForUser finds all pending shares for a user
func (r *ShareRepository) FindPendingForUser(ctx context.Context, userID uuid.UUID) ([]share.Share, error) {
	var modelList []models.Share
	if err := r.db.WithContext(ctx).Where("shared_with_id = ? AND status = ?", userID, "pending").Order("created_at DESC").Find(&modelList).Error; err != nil {
		return nil, err
	}
	return toShareDomainList(modelList), nil
}

func toShareModel(s *share.Share) models.Share {
	return models.Share{
		ID:              s.ID,
		ResourceType:    s.ResourceType,
		ResourceID:      s.ResourceID,
		ResourceName:    s.ResourceName,
		SharedByID:      s.SharedByID,
		SharedByEmail:   s.SharedByEmail,
		SharedWithID:    s.SharedWithID,
		SharedWithEmail: s.SharedWithEmail,
		Permission:      string(s.Permission),
		Status:          string(s.Status),
		Message:         s.Message,
		ExpiresAt:       s.ExpiresAt,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

func toShareDomain(m models.Share) share.Share {
	return share.Share{
		ID:              m.ID,
		ResourceType:    m.ResourceType,
		ResourceID:      m.ResourceID,
		ResourceName:    m.ResourceName,
		SharedByID:      m.SharedByID,
		SharedByEmail:   m.SharedByEmail,
		SharedWithID:    m.SharedWithID,
		SharedWithEmail: m.SharedWithEmail,
		Permission:      share.SharePermission(m.Permission),
		Status:          share.ShareStatus(m.Status),
		Message:         m.Message,
		ExpiresAt:       m.ExpiresAt,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func toShareDomainList(modelList []models.Share) []share.Share {
	result := make([]share.Share, len(modelList))
	for i, m := range modelList {
		result[i] = toShareDomain(m)
	}
	return result
}
