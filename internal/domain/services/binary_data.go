package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type BinaryDataService struct {
	db          *gorm.DB
	storagePath string
	storageType string
}

func NewBinaryDataService(db *gorm.DB, storagePath string) *BinaryDataService {
	if storagePath == "" {
		storagePath = "/tmp/linkflow/binary"
	}
	return &BinaryDataService{
		db:          db,
		storagePath: storagePath,
		storageType: "local",
	}
}

type StoreBinaryInput struct {
	ExecutionID uuid.UUID
	WorkspaceID uuid.UUID
	NodeID      string
	FileName    string
	MimeType    string
	Data        io.Reader
	Size        int64
	Metadata    models.JSON
	ExpiresAt   *time.Time
}

func (s *BinaryDataService) Store(ctx context.Context, input StoreBinaryInput) (*models.BinaryData, error) {
	dir := filepath.Join(s.storagePath, input.WorkspaceID.String(), input.ExecutionID.String())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	id := uuid.New()
	storagePath := filepath.Join(dir, id.String())

	file, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)

	written, err := io.Copy(writer, input.Data)
	if err != nil {
		os.Remove(storagePath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	binaryData := &models.BinaryData{
		ID:          id,
		ExecutionID: input.ExecutionID,
		WorkspaceID: input.WorkspaceID,
		NodeID:      input.NodeID,
		FileName:    input.FileName,
		MimeType:    input.MimeType,
		Size:        written,
		StoragePath: storagePath,
		StorageType: s.storageType,
		Checksum:    checksum,
		Metadata:    input.Metadata,
		ExpiresAt:   input.ExpiresAt,
	}

	if err := s.db.WithContext(ctx).Create(binaryData).Error; err != nil {
		os.Remove(storagePath)
		return nil, err
	}

	return binaryData, nil
}

func (s *BinaryDataService) Get(ctx context.Context, id, workspaceID uuid.UUID) (*models.BinaryData, error) {
	var data models.BinaryData
	if err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", id, workspaceID).First(&data).Error; err != nil {
		return nil, ErrNotFound
	}
	return &data, nil
}

func (s *BinaryDataService) GetContent(ctx context.Context, id, workspaceID uuid.UUID) (*models.BinaryData, io.ReadCloser, error) {
	data, err := s.Get(ctx, id, workspaceID)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(data.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return data, file, nil
}

func (s *BinaryDataService) Delete(ctx context.Context, id, workspaceID uuid.UUID) error {
	data, err := s.Get(ctx, id, workspaceID)
	if err != nil {
		return err
	}

	if err := os.Remove(data.StoragePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return s.db.WithContext(ctx).Delete(&models.BinaryData{}, "id = ?", id).Error
}

func (s *BinaryDataService) ListByExecution(ctx context.Context, executionID, workspaceID uuid.UUID) ([]models.BinaryData, error) {
	var data []models.BinaryData
	err := s.db.WithContext(ctx).
		Where("execution_id = ? AND workspace_id = ?", executionID, workspaceID).
		Order("created_at DESC").
		Find(&data).Error
	return data, err
}

func (s *BinaryDataService) ListByNode(ctx context.Context, executionID uuid.UUID, nodeID string, workspaceID uuid.UUID) ([]models.BinaryData, error) {
	var data []models.BinaryData
	err := s.db.WithContext(ctx).
		Where("execution_id = ? AND node_id = ? AND workspace_id = ?", executionID, nodeID, workspaceID).
		Order("created_at DESC").
		Find(&data).Error
	return data, err
}

func (s *BinaryDataService) CleanupExpired(ctx context.Context) (int64, error) {
	var expired []models.BinaryData
	if err := s.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Find(&expired).Error; err != nil {
		return 0, err
	}

	var deleted int64
	for _, data := range expired {
		if err := os.Remove(data.StoragePath); err != nil && !os.IsNotExist(err) {
			continue
		}
		if err := s.db.WithContext(ctx).Delete(&data).Error; err == nil {
			deleted++
		}
	}

	return deleted, nil
}

func (s *BinaryDataService) CleanupByExecution(ctx context.Context, executionID, workspaceID uuid.UUID) error {
	data, err := s.ListByExecution(ctx, executionID, workspaceID)
	if err != nil {
		return err
	}

	for _, d := range data {
		_ = os.Remove(d.StoragePath)
	}

	return s.db.WithContext(ctx).
		Where("execution_id = ? AND workspace_id = ?", executionID, workspaceID).
		Delete(&models.BinaryData{}).Error
}

func (s *BinaryDataService) GetStorageStats(ctx context.Context, workspaceID uuid.UUID) (map[string]interface{}, error) {
	var stats struct {
		TotalFiles int64 `gorm:"column:total_files"`
		TotalSize  int64 `gorm:"column:total_size"`
	}

	err := s.db.WithContext(ctx).
		Model(&models.BinaryData{}).
		Select("COUNT(*) as total_files, COALESCE(SUM(size), 0) as total_size").
		Where("workspace_id = ?", workspaceID).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_files": stats.TotalFiles,
		"total_size":  stats.TotalSize,
		"storage_type": s.storageType,
	}, nil
}
