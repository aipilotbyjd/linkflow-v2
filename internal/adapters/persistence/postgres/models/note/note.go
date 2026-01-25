package note

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
)

type Note struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null"`
	WorkflowID  uuid.UUID  `gorm:"type:uuid;index;not null"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null"`
	NodeID      *string    `gorm:"size:100;index"`
	Content     string     `gorm:"type:text;not null"`
	Resolved    bool       `gorm:"default:false;index"`
	ResolvedAt  *time.Time
	ResolvedBy  *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Workspace models.Workspace `gorm:"foreignKey:WorkspaceID"`
	Workflow  models.Workflow  `gorm:"foreignKey:WorkflowID"`
	User      models.User      `gorm:"foreignKey:UserID"`
	Resolver  *models.User     `gorm:"foreignKey:ResolvedBy"`
}

func (Note) TableName() string {
	return "notes"
}
