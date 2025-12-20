package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type WebhookEndpointRepository struct {
	*BaseRepository[models.WebhookEndpoint]
}

func NewWebhookEndpointRepository(db *gorm.DB) *WebhookEndpointRepository {
	return &WebhookEndpointRepository{
		BaseRepository: NewBaseRepository[models.WebhookEndpoint](db),
	}
}

func (r *WebhookEndpointRepository) FindByPath(ctx context.Context, path string) (*models.WebhookEndpoint, error) {
	var endpoint models.WebhookEndpoint
	err := r.DB().WithContext(ctx).
		Preload("Workflow").
		Where("path = ? AND is_active = ?", path, true).
		First(&endpoint).Error
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (r *WebhookEndpointRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]models.WebhookEndpoint, error) {
	var endpoints []models.WebhookEndpoint
	err := r.DB().WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Find(&endpoints).Error
	return endpoints, err
}

func (r *WebhookEndpointRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]models.WebhookEndpoint, error) {
	var endpoints []models.WebhookEndpoint
	err := r.DB().WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&endpoints).Error
	return endpoints, err
}

func (r *WebhookEndpointRepository) ExistsByPath(ctx context.Context, path string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&models.WebhookEndpoint{}).
		Where("path = ?", path).
		Count(&count).Error
	return count > 0, err
}

func (r *WebhookEndpointRepository) SetActive(ctx context.Context, endpointID uuid.UUID, isActive bool) error {
	return r.DB().WithContext(ctx).Model(&models.WebhookEndpoint{}).
		Where("id = ?", endpointID).
		Update("is_active", isActive).Error
}

func (r *WebhookEndpointRepository) RecordCall(ctx context.Context, endpointID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.WebhookEndpoint{}).
		Where("id = ?", endpointID).
		Updates(map[string]interface{}{
			"last_called_at": time.Now(),
			"call_count":     gorm.Expr("call_count + 1"),
		}).Error
}

func (r *WebhookEndpointRepository) DeactivateByWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	return r.DB().WithContext(ctx).Model(&models.WebhookEndpoint{}).
		Where("workflow_id = ?", workflowID).
		Update("is_active", false).Error
}

// Webhook Log methods
type WebhookLogRepository struct {
	*BaseRepository[models.WebhookLog]
}

func NewWebhookLogRepository(db *gorm.DB) *WebhookLogRepository {
	return &WebhookLogRepository{
		BaseRepository: NewBaseRepository[models.WebhookLog](db),
	}
}

func (r *WebhookLogRepository) FindByEndpointID(ctx context.Context, endpointID uuid.UUID, opts *ListOptions) ([]models.WebhookLog, int64, error) {
	var logs []models.WebhookLog
	var total int64

	query := r.DB().WithContext(ctx).Where("endpoint_id = ?", endpointID)
	query.Model(&models.WebhookLog{}).Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit).Order("created_at DESC")
	}

	err := query.Find(&logs).Error
	return logs, total, err
}

func (r *WebhookLogRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.DB().WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&models.WebhookLog{})
	return result.RowsAffected, result.Error
}
