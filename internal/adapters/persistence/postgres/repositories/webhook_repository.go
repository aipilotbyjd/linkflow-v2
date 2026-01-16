package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	"github.com/linkflow-ai/linkflow/internal/shared/errors"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// WebhookRepository implements webhook.Repository
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Create creates a new webhook endpoint
func (r *WebhookRepository) Create(ctx context.Context, e *webhook.Endpoint) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		if errors.IsUniqueConstraintError(err) {
			return webhook.ErrPathAlreadyExists
		}
		return errors.Wrap("create webhook", err)
	}
	return nil
}

// Update updates a webhook endpoint
func (r *WebhookRepository) Update(ctx context.Context, e *webhook.Endpoint) error {
	return r.db.WithContext(ctx).Save(e).Error
}

// Delete soft-deletes a webhook endpoint
func (r *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&webhook.Endpoint{}, "id = ?", id).Error
}

// FindByID finds a webhook endpoint by ID
func (r *WebhookRepository) FindByID(ctx context.Context, id uuid.UUID) (*webhook.Endpoint, error) {
	var e webhook.Endpoint
	if err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error; err != nil {
		if errors.IsNotFoundError(err) {
			return nil, webhook.ErrEndpointNotFound
		}
		return nil, err
	}
	return &e, nil
}

// FindByPath finds a webhook endpoint by path
func (r *WebhookRepository) FindByPath(ctx context.Context, path string) (*webhook.Endpoint, error) {
	var e webhook.Endpoint
	if err := r.db.WithContext(ctx).Where("path = ?", path).First(&e).Error; err != nil {
		if errors.IsNotFoundError(err) {
			return nil, webhook.ErrEndpointNotFound
		}
		return nil, err
	}
	return &e, nil
}

// FindByWorkflowID finds webhook endpoints by workflow ID
func (r *WebhookRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]webhook.Endpoint, error) {
	var endpoints []webhook.Endpoint
	err := r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Find(&endpoints).Error
	return endpoints, err
}

// FindByWorkspaceID finds webhook endpoints by workspace ID
func (r *WebhookRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]webhook.Endpoint, int64, error) {
	var endpoints []webhook.Endpoint
	var total int64

	query := r.db.WithContext(ctx).Model(&webhook.Endpoint{}).Where("workspace_id = ?", workspaceID)
	query.Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}
	query = query.Order("created_at DESC")

	if err := query.Find(&endpoints).Error; err != nil {
		return nil, 0, err
	}

	return endpoints, total, nil
}

// FindByWorkflowAndNode finds a webhook endpoint by workflow and node ID
func (r *WebhookRepository) FindByWorkflowAndNode(ctx context.Context, workflowID uuid.UUID, nodeID string) (*webhook.Endpoint, error) {
	var e webhook.Endpoint
	if err := r.db.WithContext(ctx).
		Where("workflow_id = ? AND node_id = ?", workflowID, nodeID).
		First(&e).Error; err != nil {
		if errors.IsNotFoundError(err) {
			return nil, webhook.ErrEndpointNotFound
		}
		return nil, err
	}
	return &e, nil
}

// ExistsByPath checks if a webhook endpoint exists by path
func (r *WebhookRepository) ExistsByPath(ctx context.Context, path string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&webhook.Endpoint{}).
		Where("path = ?", path).
		Count(&count).Error
	return count > 0, err
}

// SetActive sets the active status
func (r *WebhookRepository) SetActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	return r.db.WithContext(ctx).Model(&webhook.Endpoint{}).
		Where("id = ?", id).
		Update("is_active", isActive).Error
}

// RecordCall records a webhook call
func (r *WebhookRepository) RecordCall(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&webhook.Endpoint{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_called_at": gorm.Expr("NOW()"),
			"call_count":     gorm.Expr("call_count + 1"),
		}).Error
}

// CountByWorkspaceID counts webhook endpoints by workspace
func (r *WebhookRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&webhook.Endpoint{}).
		Where("workspace_id = ?", workspaceID).
		Count(&count).Error
	return count, err
}

// DeleteByWorkflowID deletes all webhook endpoints for a workflow
func (r *WebhookRepository) DeleteByWorkflowID(ctx context.Context, workflowID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Delete(&webhook.Endpoint{}).Error
}

// EventRepository implements webhook.EventRepository
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create creates a new webhook event
func (r *EventRepository) Create(ctx context.Context, e *webhook.Event) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// FindByID finds a webhook event by ID
func (r *EventRepository) FindByID(ctx context.Context, id uuid.UUID) (*webhook.Event, error) {
	var e webhook.Event
	if err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

// FindByEndpointID finds webhook events by endpoint ID
func (r *EventRepository) FindByEndpointID(ctx context.Context, endpointID uuid.UUID, opts *types.ListOptions) ([]webhook.Event, int64, error) {
	var events []webhook.Event
	var total int64

	query := r.db.WithContext(ctx).Model(&webhook.Event{}).Where("endpoint_id = ?", endpointID)
	query.Count(&total)

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}
	query = query.Order("created_at DESC")

	if err := query.Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// DeleteByEndpointID deletes all events for an endpoint
func (r *EventRepository) DeleteByEndpointID(ctx context.Context, endpointID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("endpoint_id = ?", endpointID).
		Delete(&webhook.Event{}).Error
}

// DeleteOlderThan deletes events older than the given time
func (r *EventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&webhook.Event{})
	return result.RowsAffected, result.Error
}

// CountByEndpointID counts events by endpoint
func (r *EventRepository) CountByEndpointID(ctx context.Context, endpointID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&webhook.Event{}).
		Where("endpoint_id = ?", endpointID).
		Count(&count).Error
	return count, err
}
