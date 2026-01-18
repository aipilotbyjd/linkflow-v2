package binarydata

import (
	"time"

	"github.com/google/uuid"
)

// BinaryData represents stored binary file metadata
type BinaryData struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	ExecutionID uuid.UUID `json:"executionId,omitempty"`
	NodeID      string    `json:"nodeId,omitempty"`
	FileName    string    `json:"fileName"`
	MimeType    string    `json:"mimeType"`
	Size        int64     `json:"size"`
	StoragePath string    `json:"-"` // Internal storage path
	CreatedAt   time.Time `json:"createdAt"`
}

// NewBinaryData creates new binary data metadata
func NewBinaryData(workspaceID uuid.UUID, fileName, mimeType string, size int64, storagePath string) *BinaryData {
	return &BinaryData{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		FileName:    fileName,
		MimeType:    mimeType,
		Size:        size,
		StoragePath: storagePath,
		CreatedAt:   time.Now(),
	}
}

// StorageStats represents storage usage statistics
type StorageStats struct {
	TotalFiles   int64 `json:"totalFiles"`
	TotalSize    int64 `json:"totalSize"`
	UsedStorage  int64 `json:"usedStorage"`
	MaxStorage   int64 `json:"maxStorage"`
	UsagePercent int   `json:"usagePercent"`
}
