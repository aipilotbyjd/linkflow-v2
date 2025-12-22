package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type WaitingExecutionRepository struct {
	*BaseRepository[models.WaitingExecution]
}

func NewWaitingExecutionRepository(db *gorm.DB) *WaitingExecutionRepository {
	return &WaitingExecutionRepository{
		BaseRepository: NewBaseRepository[models.WaitingExecution](db),
	}
}

func (r *WaitingExecutionRepository) FindByToken(ctx context.Context, token string) (*models.WaitingExecution, error) {
	var waiting models.WaitingExecution
	err := r.DB().WithContext(ctx).
		Where("resume_token = ?", token).
		First(&waiting).Error
	if err != nil {
		return nil, err
	}
	return &waiting, nil
}

func (r *WaitingExecutionRepository) FindByWebhookPath(ctx context.Context, path string) (*models.WaitingExecution, error) {
	var waiting models.WaitingExecution
	err := r.DB().WithContext(ctx).
		Where("webhook_path = ? AND status = ?", path, "waiting").
		First(&waiting).Error
	if err != nil {
		return nil, err
	}
	return &waiting, nil
}

func (r *WaitingExecutionRepository) FindByExecutionID(ctx context.Context, executionID uuid.UUID) ([]models.WaitingExecution, error) {
	var waitings []models.WaitingExecution
	err := r.DB().WithContext(ctx).
		Where("execution_id = ?", executionID).
		Order("created_at DESC").
		Find(&waitings).Error
	return waitings, err
}

func (r *WaitingExecutionRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]models.WaitingExecution, error) {
	var waitings []models.WaitingExecution
	err := r.DB().WithContext(ctx).
		Where("workflow_id = ? AND status = ?", workflowID, "waiting").
		Order("created_at DESC").
		Find(&waitings).Error
	return waitings, err
}

func (r *WaitingExecutionRepository) FindPendingByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]models.WaitingExecution, int64, error) {
	var waitings []models.WaitingExecution
	var total int64

	query := r.DB().WithContext(ctx).Where("workspace_id = ? AND status = ?", workspaceID, "waiting")
	query.Model(&models.WaitingExecution{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order("created_at DESC")
	}

	err := query.Find(&waitings).Error
	return waitings, total, err
}

func (r *WaitingExecutionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	updates := map[string]interface{}{"status": status}
	if status == "resumed" {
		now := time.Now()
		updates["resumed_at"] = now
	}
	return r.DB().WithContext(ctx).Model(&models.WaitingExecution{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *WaitingExecutionRepository) ExpireOld(ctx context.Context) (int64, error) {
	result := r.DB().WithContext(ctx).Model(&models.WaitingExecution{}).
		Where("status = ? AND timeout_at < ?", "waiting", time.Now()).
		Update("status", "expired")
	return result.RowsAffected, result.Error
}

func (r *WaitingExecutionRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.DB().WithContext(ctx).
		Where("created_at < ? AND status IN ?", cutoff, []string{"resumed", "expired"}).
		Delete(&models.WaitingExecution{})
	return result.RowsAffected, result.Error
}

// TemplateRepository handles template operations
type TemplateRepository struct {
	*BaseRepository[models.Template]
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{
		BaseRepository: NewBaseRepository[models.Template](db),
	}
}

func (r *TemplateRepository) FindPublic(ctx context.Context, opts *ListOptions) ([]models.Template, int64, error) {
	var templates []models.Template
	var total int64

	query := r.DB().WithContext(ctx).Where("is_public = ?", true)
	query.Model(&models.Template{}).Count(&total)

	if opts != nil {
		if opts.OrderBy != "" {
			query = query.Order(opts.OrderBy + " " + opts.Order)
		}
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	err := query.Find(&templates).Error
	return templates, total, err
}

func (r *TemplateRepository) FindByCategory(ctx context.Context, category string, opts *ListOptions) ([]models.Template, int64, error) {
	var templates []models.Template
	var total int64

	query := r.DB().WithContext(ctx).Where("category = ? AND is_public = ?", category, true)
	query.Model(&models.Template{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order("use_count DESC")
	}

	err := query.Find(&templates).Error
	return templates, total, err
}

func (r *TemplateRepository) FindFeatured(ctx context.Context, limit int) ([]models.Template, error) {
	var templates []models.Template
	err := r.DB().WithContext(ctx).
		Where("is_featured = ? AND is_public = ?", true, true).
		Order("use_count DESC").
		Limit(limit).
		Find(&templates).Error
	return templates, err
}

func (r *TemplateRepository) Search(ctx context.Context, query string, opts *ListOptions) ([]models.Template, int64, error) {
	var templates []models.Template
	var total int64

	dbQuery := r.DB().WithContext(ctx).
		Where("is_public = ? AND (name ILIKE ? OR description ILIKE ?)", true, "%"+query+"%", "%"+query+"%")
	dbQuery.Model(&models.Template{}).Count(&total)

	if opts != nil {
		dbQuery = dbQuery.Offset(opts.Offset).Limit(opts.Limit).Order("use_count DESC")
	}

	err := dbQuery.Find(&templates).Error
	return templates, total, err
}

func (r *TemplateRepository) IncrementUseCount(ctx context.Context, templateID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.Template{}).
		Where("id = ?", templateID).
		Update("use_count", gorm.Expr("use_count + 1")).Error
}

// PinnedDataRepository handles pinned data operations
type PinnedDataRepository struct {
	*BaseRepository[models.PinnedData]
}

func NewPinnedDataRepository(db *gorm.DB) *PinnedDataRepository {
	return &PinnedDataRepository{
		BaseRepository: NewBaseRepository[models.PinnedData](db),
	}
}

func (r *PinnedDataRepository) FindByWorkflowAndNode(ctx context.Context, workflowID uuid.UUID, nodeID string) (*models.PinnedData, error) {
	var pinned models.PinnedData
	err := r.DB().WithContext(ctx).
		Where("workflow_id = ? AND node_id = ?", workflowID, nodeID).
		First(&pinned).Error
	if err != nil {
		return nil, err
	}
	return &pinned, nil
}

func (r *PinnedDataRepository) FindByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.PinnedData, error) {
	var pinnedList []models.PinnedData
	err := r.DB().WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Find(&pinnedList).Error
	return pinnedList, err
}

func (r *PinnedDataRepository) Upsert(ctx context.Context, pinned *models.PinnedData) error {
	return r.DB().WithContext(ctx).
		Where("workflow_id = ? AND node_id = ?", pinned.WorkflowID, pinned.NodeID).
		Assign(models.PinnedData{
			Data:      pinned.Data,
			Name:      pinned.Name,
			CreatedBy: pinned.CreatedBy,
		}).
		FirstOrCreate(pinned).Error
}

func (r *PinnedDataRepository) DeleteByWorkflowAndNode(ctx context.Context, workflowID uuid.UUID, nodeID string) error {
	return r.DB().WithContext(ctx).
		Where("workflow_id = ? AND node_id = ?", workflowID, nodeID).
		Delete(&models.PinnedData{}).Error
}

// BinaryDataRepository handles binary data operations
type BinaryDataRepository struct {
	*BaseRepository[models.BinaryData]
}

func NewBinaryDataRepository(db *gorm.DB) *BinaryDataRepository {
	return &BinaryDataRepository{
		BaseRepository: NewBaseRepository[models.BinaryData](db),
	}
}

func (r *BinaryDataRepository) FindByExecution(ctx context.Context, executionID uuid.UUID) ([]models.BinaryData, error) {
	var data []models.BinaryData
	err := r.DB().WithContext(ctx).
		Where("execution_id = ?", executionID).
		Find(&data).Error
	return data, err
}

func (r *BinaryDataRepository) FindByExecutionAndNode(ctx context.Context, executionID uuid.UUID, nodeID string) ([]models.BinaryData, error) {
	var data []models.BinaryData
	err := r.DB().WithContext(ctx).
		Where("execution_id = ? AND node_id = ?", executionID, nodeID).
		Find(&data).Error
	return data, err
}

func (r *BinaryDataRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.DB().WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.BinaryData{})
	return result.RowsAffected, result.Error
}

// OAuthStateRepository handles OAuth state operations
type OAuthStateRepository struct {
	*BaseRepository[models.OAuthState]
}

func NewOAuthStateRepository(db *gorm.DB) *OAuthStateRepository {
	return &OAuthStateRepository{
		BaseRepository: NewBaseRepository[models.OAuthState](db),
	}
}

func (r *OAuthStateRepository) FindByState(ctx context.Context, state string) (*models.OAuthState, error) {
	var oauthState models.OAuthState
	err := r.DB().WithContext(ctx).
		Where("state = ? AND expires_at > ?", state, time.Now()).
		First(&oauthState).Error
	if err != nil {
		return nil, err
	}
	return &oauthState, nil
}

func (r *OAuthStateRepository) DeleteByState(ctx context.Context, state string) error {
	return r.DB().WithContext(ctx).
		Where("state = ?", state).
		Delete(&models.OAuthState{}).Error
}

func (r *OAuthStateRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.DB().WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.OAuthState{})
	return result.RowsAffected, result.Error
}
