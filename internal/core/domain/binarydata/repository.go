package binarydata

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// Repository defines the binary data metadata repository interface
type Repository interface {
	Create(ctx context.Context, data *BinaryData) error
	FindByID(ctx context.Context, id uuid.UUID) (*BinaryData, error)
	FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]BinaryData, int64, error)
	FindByExecution(ctx context.Context, executionID uuid.UUID) ([]BinaryData, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOlderThan(ctx context.Context, workspaceID uuid.UUID, days int) (int64, error)
	GetStats(ctx context.Context, workspaceID uuid.UUID) (*StorageStats, error)
}

// StorageService defines the file storage interface
type StorageService interface {
	Upload(ctx context.Context, workspaceID uuid.UUID, fileName string, reader io.Reader, size int64) (string, error)
	Download(ctx context.Context, storagePath string) (io.ReadCloser, error)
	Delete(ctx context.Context, storagePath string) error
	GetURL(ctx context.Context, storagePath string) (string, error)
}
