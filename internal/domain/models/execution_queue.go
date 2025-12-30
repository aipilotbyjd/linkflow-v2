package models

import (
	"time"

	"github.com/google/uuid"
)

// ExecutionQueue represents priority-based execution queues
type ExecutionQueue struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Priority    int       `gorm:"default:5;index" json:"priority"` // 1-10, higher = more priority
	Concurrency int       `gorm:"default:1" json:"concurrency"`    // Max concurrent executions
	RateLimit   int       `gorm:"default:0" json:"rate_limit"`     // Executions per minute, 0 = unlimited
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (ExecutionQueue) TableName() string {
	return "execution_queues"
}
