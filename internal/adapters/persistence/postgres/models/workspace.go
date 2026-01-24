package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type Workspace struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OwnerID          uuid.UUID  `gorm:"type:uuid;index;not null"`
	Name             string     `gorm:"size:100;not null"`
	Slug             string     `gorm:"size:100;uniqueIndex;not null"`
	Description      *string    `gorm:"type:text"`
	LogoURL          *string    `gorm:"size:500"`
	Website          *string    `gorm:"size:255"`
	Timezone         string     `gorm:"size:50;default:UTC"`
	Language         string     `gorm:"size:10;default:en"`
	Currency         string     `gorm:"size:3;default:USD"`
	Country          *string    `gorm:"size:2"`
	Industry         *string    `gorm:"size:50"`
	CompanySize      *string    `gorm:"size:20"`
	BillingEmail     *string    `gorm:"size:255"`
	Settings         types.JSON `gorm:"type:jsonb;default:'{}'"`
	PlanID           string     `gorm:"size:50;default:free"`
	StripeCustomerID *string    `gorm:"size:255"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	Owner   User              `gorm:"foreignKey:OwnerID"`
	Members []WorkspaceMember `gorm:"foreignKey:WorkspaceID"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

type WorkspaceMember struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null"`
	Role        string     `gorm:"size:20;not null;default:member"`
	InvitedBy   *uuid.UUID `gorm:"type:uuid"`
	InvitedAt   *time.Time
	JoinedAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID"`
	User      User      `gorm:"foreignKey:UserID"`

	// RBAC
	RoleID  *uuid.UUID `gorm:"type:uuid;index"`
	RoleRef *Role      `gorm:"foreignKey:RoleID"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}
