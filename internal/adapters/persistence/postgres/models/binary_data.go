package models

import (
	"time"

	"github.com/google/uuid"
)

// BinaryData represents the database model for binary file metadata
type BinaryData struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ExecutionID *uuid.UUID `gorm:"type:uuid;index"`
	NodeID      string     `gorm:"type:varchar(255)"`
	FileName    string     `gorm:"type:varchar(255);not null"`
	MimeType    string     `gorm:"type:varchar(100);not null"`
	Size        int64      `gorm:"not null"`
	StoragePath string     `gorm:"type:varchar(500);not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
}

// TableName returns the table name for BinaryData
func (BinaryData) TableName() string {
	return "binary_data"
}
