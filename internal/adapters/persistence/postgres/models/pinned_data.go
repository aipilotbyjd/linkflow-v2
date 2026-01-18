package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PinnedData represents the database model for pinned data
type PinnedData struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey"`
	WorkflowID uuid.UUID       `gorm:"type:uuid;not null;index:idx_pinned_workflow_node,unique"`
	NodeID     string          `gorm:"type:varchar(255);not null;index:idx_pinned_workflow_node,unique"`
	Data       json.RawMessage `gorm:"type:jsonb"`
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime"`
}

// TableName returns the table name for PinnedData
func (PinnedData) TableName() string {
	return "pinned_data"
}
