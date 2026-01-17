package binarydata

import (
	"io"
	"time"
)

// BinaryData represents binary data metadata
type BinaryData struct {
	ID          string     `json:"id"`
	ExecutionID string     `json:"executionId"`
	NodeID      string     `json:"nodeId"`
	FileName    string     `json:"fileName"`
	MimeType    string     `json:"mimeType"`
	Size        int64      `json:"size"`
	StoragePath string     `json:"storagePath,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

// BinaryDataStats represents binary data statistics
type BinaryDataStats struct {
	TotalFiles     int                      `json:"totalFiles"`
	TotalSizeBytes int64                    `json:"totalSizeBytes"`
	ByMimeType     map[string]MimeTypeStats `json:"byMimeType"`
}

// MimeTypeStats represents statistics by mime type
type MimeTypeStats struct {
	Count int   `json:"count"`
	Size  int64 `json:"size"`
}

// StorageService defines the storage service interface
type StorageService interface {
	Upload(data io.Reader, fileName, mimeType string) (string, error)
	Download(path string) (io.ReadCloser, error)
	Delete(path string) error
	GetInfo(path string) (*BinaryData, error)
}
