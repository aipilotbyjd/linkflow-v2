package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Workspace struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"owner_id"`
	Name             string         `gorm:"size:100;not null" json:"name"`
	Slug             string         `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description      *string        `gorm:"type:text" json:"description,omitempty"`
	LogoURL          *string        `gorm:"size:500" json:"logo_url,omitempty"`
	Settings         JSON           `gorm:"type:jsonb;default:'{}'" json:"settings"`
	PlanID           string         `gorm:"size:50;default:free" json:"plan_id"`
	StripeCustomerID *string        `gorm:"size:255" json:"-"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Owner       User              `gorm:"foreignKey:OwnerID" json:"-"`
	Members     []WorkspaceMember `gorm:"foreignKey:WorkspaceID" json:"-"`
	Workflows   []Workflow        `gorm:"foreignKey:WorkspaceID" json:"-"`
	Credentials []Credential      `gorm:"foreignKey:WorkspaceID" json:"-"`
	Schedules   []Schedule        `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

type WorkspaceMember struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Role        string     `gorm:"size:20;not null;default:member" json:"role"`
	InvitedBy   *uuid.UUID `gorm:"type:uuid" json:"invited_by,omitempty"`
	InvitedAt   *time.Time `json:"invited_at,omitempty"`
	JoinedAt    *time.Time `gorm:"default:now()" json:"joined_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	// Relations
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	Inviter   *User     `gorm:"foreignKey:InvitedBy" json:"-"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

type WorkspaceInvitation struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Email       string     `gorm:"size:255;not null" json:"email"`
	Role        string     `gorm:"size:20;not null;default:member" json:"role"`
	Token       string     `gorm:"size:255;uniqueIndex;not null" json:"-"`
	InvitedBy   uuid.UUID  `gorm:"type:uuid;not null" json:"invited_by"`
	ExpiresAt   time.Time  `gorm:"not null" json:"expires_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Inviter   User      `gorm:"foreignKey:InvitedBy" json:"-"`
}

func (WorkspaceInvitation) TableName() string {
	return "workspace_invitations"
}
