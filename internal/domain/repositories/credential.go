package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type CredentialRepository struct {
	*BaseRepository[models.Credential]
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{
		BaseRepository: NewBaseRepository[models.Credential](db),
	}
}

func (r *CredentialRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]models.Credential, int64, error) {
	var credentials []models.Credential
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	query.Model(&models.Credential{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order(opts.OrderBy + " " + opts.Order)
	}

	err := query.Find(&credentials).Error
	return credentials, total, err
}

func (r *CredentialRepository) FindByType(ctx context.Context, workspaceID uuid.UUID, credType string) ([]models.Credential, error) {
	var credentials []models.Credential
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND type = ?", workspaceID, credType).
		Order("name ASC").
		Find(&credentials).Error
	return credentials, err
}

func (r *CredentialRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("workspace_id = ?", workspaceID).
		Count(&count).Error
	return count, err
}

func (r *CredentialRepository) UpdateLastUsed(ctx context.Context, credentialID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("id = ?", credentialID).
		Update("last_used_at", time.Now()).Error
}

func (r *CredentialRepository) UpdateData(ctx context.Context, credentialID uuid.UUID, encryptedData string) error {
	return r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("id = ?", credentialID).
		Update("data", encryptedData).Error
}

// CredentialFilter contains filter parameters for credential queries
type CredentialFilter struct {
	Type   *string
	Search *string
	SortBy string
	Order  string
}

// FindWithFilters returns credentials matching the given filters
func (r *CredentialRepository) FindWithFilters(ctx context.Context, workspaceID uuid.UUID, filter *CredentialFilter, opts *ListOptions) ([]models.Credential, int64, error) {
	var credentials []models.Credential
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)

	// Apply filters
	if filter != nil {
		if filter.Type != nil && *filter.Type != "" {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.Search != nil && *filter.Search != "" {
			searchPattern := "%" + *filter.Search + "%"
			query = query.Where("(name ILIKE ? OR description ILIKE ?)", searchPattern, searchPattern)
		}
	}

	// Count total before pagination
	query.Model(&models.Credential{}).Count(&total)

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

	err := query.Find(&credentials).Error
	return credentials, total, err
}

// FindExpiringTokens returns OAuth credentials with tokens expiring within the given duration
func (r *CredentialRepository) FindExpiringTokens(ctx context.Context, within time.Duration) ([]models.Credential, error) {
	var credentials []models.Credential
	expiryThreshold := time.Now().Add(within)

	err := r.DB().WithContext(ctx).
		Where("type = ? AND token_expires_at IS NOT NULL AND token_expires_at <= ?",
			models.CredentialTypeOAuth2, expiryThreshold).
		Find(&credentials).Error

	return credentials, err
}

// FindByProvider returns all credentials for a specific OAuth provider
func (r *CredentialRepository) FindByProvider(ctx context.Context, workspaceID uuid.UUID, provider string) ([]models.Credential, error) {
	var credentials []models.Credential
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ? AND provider = ?", workspaceID, provider).
		Order("name ASC").
		Find(&credentials).Error
	return credentials, err
}

// UpdateTokenExpiry updates the token expiry time for a credential
func (r *CredentialRepository) UpdateTokenExpiry(ctx context.Context, credentialID uuid.UUID, expiresAt *time.Time) error {
	return r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("id = ?", credentialID).
		Update("token_expires_at", expiresAt).Error
}

// FindAccessibleByUser returns credentials the user can access in a workspace
// This includes: owned credentials, workspace-scoped, and specifically shared
func (r *CredentialRepository) FindAccessibleByUser(ctx context.Context, workspaceID, userID uuid.UUID, filter *CredentialFilter, opts *ListOptions) ([]models.Credential, int64, error) {
	var credentials []models.Credential
	var total int64

	// Build query for credentials user can access:
	// 1. User owns it (created_by = userID)
	// 2. Workspace scope AND user is workspace member (handled by caller)
	// 3. Specific scope AND user has a share record
	query := r.DB().WithContext(ctx).
		Preload("Shares", "deleted_at IS NULL").
		Where("workspace_id = ?", workspaceID).
		Where(`(
			created_by = ? 
			OR sharing_scope = 'workspace' 
			OR (sharing_scope = 'specific' AND id IN (
				SELECT credential_id FROM credential_shares 
				WHERE user_id = ? AND deleted_at IS NULL
			))
		)`, userID, userID)

	// Apply filters
	if filter != nil {
		if filter.Type != nil && *filter.Type != "" {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.Search != nil && *filter.Search != "" {
			searchPattern := "%" + *filter.Search + "%"
			query = query.Where("(name ILIKE ? OR description ILIKE ?)", searchPattern, searchPattern)
		}
	}

	// Count total before pagination
	countQuery := r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("workspace_id = ?", workspaceID).
		Where(`(
			created_by = ? 
			OR sharing_scope = 'workspace' 
			OR (sharing_scope = 'specific' AND id IN (
				SELECT credential_id FROM credential_shares 
				WHERE user_id = ? AND deleted_at IS NULL
			))
		)`, userID, userID)

	if filter != nil {
		if filter.Type != nil && *filter.Type != "" {
			countQuery = countQuery.Where("type = ?", *filter.Type)
		}
		if filter.Search != nil && *filter.Search != "" {
			searchPattern := "%" + *filter.Search + "%"
			countQuery = countQuery.Where("(name ILIKE ? OR description ILIKE ?)", searchPattern, searchPattern)
		}
	}
	countQuery.Count(&total)

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

	err := query.Find(&credentials).Error
	return credentials, total, err
}

// FindByIDWithShares returns a credential with its shares preloaded
func (r *CredentialRepository) FindByIDWithShares(ctx context.Context, id uuid.UUID) (*models.Credential, error) {
	var credential models.Credential
	err := r.DB().WithContext(ctx).
		Preload("Shares", "deleted_at IS NULL").
		Preload("Shares.User").
		Where("id = ?", id).
		First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// UpdateSharingScope updates the sharing scope of a credential
func (r *CredentialRepository) UpdateSharingScope(ctx context.Context, credentialID uuid.UUID, scope models.SharingScope) error {
	return r.DB().WithContext(ctx).Model(&models.Credential{}).
		Where("id = ?", credentialID).
		Update("sharing_scope", scope).Error
}

// CredentialShareRepository handles credential share database operations
type CredentialShareRepository struct {
	*BaseRepository[models.CredentialShare]
}

func NewCredentialShareRepository(db *gorm.DB) *CredentialShareRepository {
	return &CredentialShareRepository{
		BaseRepository: NewBaseRepository[models.CredentialShare](db),
	}
}

// FindByCredentialID returns all shares for a credential
func (r *CredentialShareRepository) FindByCredentialID(ctx context.Context, credentialID uuid.UUID) ([]models.CredentialShare, error) {
	var shares []models.CredentialShare
	err := r.DB().WithContext(ctx).
		Preload("User").
		Where("credential_id = ?", credentialID).
		Find(&shares).Error
	return shares, err
}

// FindByUserID returns all credentials shared with a user
func (r *CredentialShareRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.CredentialShare, error) {
	var shares []models.CredentialShare
	err := r.DB().WithContext(ctx).
		Preload("Credential").
		Where("user_id = ?", userID).
		Find(&shares).Error
	return shares, err
}

// FindShare returns a specific share record
func (r *CredentialShareRepository) FindShare(ctx context.Context, credentialID, userID uuid.UUID) (*models.CredentialShare, error) {
	var share models.CredentialShare
	err := r.DB().WithContext(ctx).
		Where("credential_id = ? AND user_id = ?", credentialID, userID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// DeleteByCredentialID removes all shares for a credential
func (r *CredentialShareRepository) DeleteByCredentialID(ctx context.Context, credentialID uuid.UUID) error {
	return r.DB().WithContext(ctx).
		Where("credential_id = ?", credentialID).
		Delete(&models.CredentialShare{}).Error
}

// DeleteShare removes a specific share
func (r *CredentialShareRepository) DeleteShare(ctx context.Context, credentialID, userID uuid.UUID) error {
	return r.DB().WithContext(ctx).
		Where("credential_id = ? AND user_id = ?", credentialID, userID).
		Delete(&models.CredentialShare{}).Error
}
