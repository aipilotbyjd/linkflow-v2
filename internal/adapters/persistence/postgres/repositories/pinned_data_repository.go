package repositories

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/pinneddata"
	"gorm.io/gorm"
)

// PinnedDataRepository implements pinneddata.Repository
type PinnedDataRepository struct {
	db *gorm.DB
}

// NewPinnedDataRepository creates a new pinned data repository
func NewPinnedDataRepository(db *gorm.DB) *PinnedDataRepository {
	return &PinnedDataRepository{db: db}
}

// GetByWorkflow gets all pinned data for a workflow
func (r *PinnedDataRepository) GetByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]pinneddata.PinnedData, error) {
	var modelList []models.PinnedData
	if err := r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Find(&modelList).Error; err != nil {
		return nil, err
	}

	result := make([]pinneddata.PinnedData, len(modelList))
	for i, m := range modelList {
		result[i] = toPinnedDataDomain(m)
	}
	return result, nil
}

// GetByNode gets pinned data for a specific node
func (r *PinnedDataRepository) GetByNode(ctx context.Context, workflowID uuid.UUID, nodeID string) (*pinneddata.PinnedData, error) {
	var model models.PinnedData
	if err := r.db.WithContext(ctx).Where("workflow_id = ? AND node_id = ?", workflowID, nodeID).First(&model).Error; err != nil {
		return nil, err
	}
	result := toPinnedDataDomain(model)
	return &result, nil
}

// Set creates or updates pinned data for a node
func (r *PinnedDataRepository) Set(ctx context.Context, workflowID uuid.UUID, nodeID string, data json.RawMessage) (*pinneddata.PinnedData, error) {
	model := models.PinnedData{
		ID:         uuid.New(),
		WorkflowID: workflowID,
		NodeID:     nodeID,
		Data:       data,
	}

	// Upsert
	if err := r.db.WithContext(ctx).Where("workflow_id = ? AND node_id = ?", workflowID, nodeID).
		Assign(models.PinnedData{Data: data}).
		FirstOrCreate(&model).Error; err != nil {
		return nil, err
	}

	result := toPinnedDataDomain(model)
	return &result, nil
}

// Delete removes pinned data for a node
func (r *PinnedDataRepository) Delete(ctx context.Context, workflowID uuid.UUID, nodeID string) error {
	return r.db.WithContext(ctx).Where("workflow_id = ? AND node_id = ?", workflowID, nodeID).Delete(&models.PinnedData{}).Error
}

func toPinnedDataDomain(m models.PinnedData) pinneddata.PinnedData {
	return pinneddata.PinnedData{
		ID:         m.ID,
		WorkflowID: m.WorkflowID,
		NodeID:     m.NodeID,
		Data:       m.Data,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
