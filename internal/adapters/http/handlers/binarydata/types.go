package binarydata

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
)

// BinaryDataResponse represents binary data in API response
type BinaryDataResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	ExecutionID string    `json:"executionId,omitempty"`
	NodeID      string    `json:"nodeId,omitempty"`
	FileName    string    `json:"fileName"`
	MimeType    string    `json:"mimeType"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
}

// StorageStatsResponse represents storage stats in API response
type StorageStatsResponse struct {
	TotalFiles   int64 `json:"totalFiles"`
	TotalSize    int64 `json:"totalSize"`
	UsedStorage  int64 `json:"usedStorage"`
	MaxStorage   int64 `json:"maxStorage"`
	UsagePercent int   `json:"usagePercent"`
}

// ToBinaryDataResponse converts domain to response
func ToBinaryDataResponse(b binarydata.BinaryData) BinaryDataResponse {
	resp := BinaryDataResponse{
		ID:          b.ID.String(),
		WorkspaceID: b.WorkspaceID.String(),
		NodeID:      b.NodeID,
		FileName:    b.FileName,
		MimeType:    b.MimeType,
		Size:        b.Size,
		CreatedAt:   b.CreatedAt,
	}
	if b.ExecutionID.String() != "00000000-0000-0000-0000-000000000000" {
		resp.ExecutionID = b.ExecutionID.String()
	}
	return resp
}

// ToBinaryDataResponseList converts domain list to response list
func ToBinaryDataResponseList(list []binarydata.BinaryData) []BinaryDataResponse {
	result := make([]BinaryDataResponse, len(list))
	for i, b := range list {
		result[i] = ToBinaryDataResponse(b)
	}
	return result
}

// ToStorageStatsResponse converts domain to response
func ToStorageStatsResponse(s *binarydata.StorageStats) StorageStatsResponse {
	return StorageStatsResponse{
		TotalFiles:   s.TotalFiles,
		TotalSize:    s.TotalSize,
		UsedStorage:  s.UsedStorage,
		MaxStorage:   s.MaxStorage,
		UsagePercent: s.UsagePercent,
	}
}
