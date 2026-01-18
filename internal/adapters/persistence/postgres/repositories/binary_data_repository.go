package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
	"gorm.io/gorm"
)

// BinaryDataRepository implements binarydata.Repository
type BinaryDataRepository struct {
	db *gorm.DB
}

// NewBinaryDataRepository creates a new binary data repository
func NewBinaryDataRepository(db *gorm.DB) *BinaryDataRepository {
	return &BinaryDataRepository{db: db}
}

// Create creates binary data metadata
func (r *BinaryDataRepository) Create(ctx context.Context, data *binarydata.BinaryData) error {
	model := toBinaryDataModel(data)
	return r.db.WithContext(ctx).Create(&model).Error
}

// FindByID finds binary data by ID
func (r *BinaryDataRepository) FindByID(ctx context.Context, id uuid.UUID) (*binarydata.BinaryData, error) {
	var model models.BinaryData
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	result := toBinaryDataDomain(model)
	return &result, nil
}

// FindByWorkspace finds binary data by workspace with pagination
func (r *BinaryDataRepository) FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]binarydata.BinaryData, int64, error) {
	var modelList []models.BinaryData
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.BinaryData{}).Where("workspace_id = ?", workspaceID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	return toBinaryDataDomainList(modelList), total, nil
}

// FindByExecution finds binary data by execution
func (r *BinaryDataRepository) FindByExecution(ctx context.Context, executionID uuid.UUID) ([]binarydata.BinaryData, error) {
	var modelList []models.BinaryData
	if err := r.db.WithContext(ctx).Where("execution_id = ?", executionID).Find(&modelList).Error; err != nil {
		return nil, err
	}
	return toBinaryDataDomainList(modelList), nil
}

// Delete deletes binary data metadata
func (r *BinaryDataRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.BinaryData{}, "id = ?", id).Error
}

// DeleteOlderThan deletes binary data older than specified days
func (r *BinaryDataRepository) DeleteOlderThan(ctx context.Context, workspaceID uuid.UUID, days int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("workspace_id = ? AND created_at < NOW() - INTERVAL '1 day' * ?", workspaceID, days).
		Delete(&models.BinaryData{})
	return result.RowsAffected, result.Error
}

// GetStats gets storage statistics for a workspace
func (r *BinaryDataRepository) GetStats(ctx context.Context, workspaceID uuid.UUID) (*binarydata.StorageStats, error) {
	var stats struct {
		TotalFiles int64 `gorm:"column:total_files"`
		TotalSize  int64 `gorm:"column:total_size"`
	}

	if err := r.db.WithContext(ctx).Model(&models.BinaryData{}).
		Select("COUNT(*) as total_files, COALESCE(SUM(size), 0) as total_size").
		Where("workspace_id = ?", workspaceID).
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	maxStorage := int64(10 * 1024 * 1024 * 1024) // 10 GB default
	usagePercent := 0
	if maxStorage > 0 {
		usagePercent = int(float64(stats.TotalSize) / float64(maxStorage) * 100)
	}

	return &binarydata.StorageStats{
		TotalFiles:   stats.TotalFiles,
		TotalSize:    stats.TotalSize,
		UsedStorage:  stats.TotalSize,
		MaxStorage:   maxStorage,
		UsagePercent: usagePercent,
	}, nil
}

func toBinaryDataModel(data *binarydata.BinaryData) models.BinaryData {
	model := models.BinaryData{
		ID:          data.ID,
		WorkspaceID: data.WorkspaceID,
		NodeID:      data.NodeID,
		FileName:    data.FileName,
		MimeType:    data.MimeType,
		Size:        data.Size,
		StoragePath: data.StoragePath,
		CreatedAt:   data.CreatedAt,
	}
	if data.ExecutionID != uuid.Nil {
		model.ExecutionID = &data.ExecutionID
	}
	return model
}

func toBinaryDataDomain(m models.BinaryData) binarydata.BinaryData {
	data := binarydata.BinaryData{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		NodeID:      m.NodeID,
		FileName:    m.FileName,
		MimeType:    m.MimeType,
		Size:        m.Size,
		StoragePath: m.StoragePath,
		CreatedAt:   m.CreatedAt,
	}
	if m.ExecutionID != nil {
		data.ExecutionID = *m.ExecutionID
	}
	return data
}

func toBinaryDataDomainList(modelList []models.BinaryData) []binarydata.BinaryData {
	result := make([]binarydata.BinaryData, len(modelList))
	for i, m := range modelList {
		result[i] = toBinaryDataDomain(m)
	}
	return result
}
