package workspace

import (
	"time"

	"github.com/google/uuid"
)

// Member represents a workspace member
type Member struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Role        Role       `gorm:"size:20;not null;default:member" json:"role"`
	InvitedBy   *uuid.UUID `gorm:"type:uuid" json:"invited_by,omitempty"`
	InvitedAt   *time.Time `json:"invited_at,omitempty"`
	JoinedAt    *time.Time `gorm:"default:now()" json:"joined_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Member) TableName() string {
	return "workspace_members"
}

// NewMember creates a new workspace member
func NewMember(workspaceID, userID uuid.UUID, role Role) *Member {
	now := time.Now()
	return &Member{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
		JoinedAt:    &now,
		CreatedAt:   now,
	}
}

// WithInviter sets the inviter information
func (m *Member) WithInviter(inviterID uuid.UUID) *Member {
	now := time.Now()
	m.InvitedBy = &inviterID
	m.InvitedAt = &now
	return m
}

// IsOwner checks if the member is the owner
func (m *Member) IsOwner() bool {
	return m.Role == RoleOwner
}

// IsAdmin checks if the member is an admin or owner
func (m *Member) IsAdmin() bool {
	return m.Role == RoleOwner || m.Role == RoleAdmin
}

// CanManageMembers checks if the member can manage other members
func (m *Member) CanManageMembers() bool {
	return m.IsAdmin()
}

// CanManageWorkflows checks if the member can manage workflows
func (m *Member) CanManageWorkflows() bool {
	return m.Role != RoleViewer
}

// CanViewWorkflows checks if the member can view workflows
func (m *Member) CanViewWorkflows() bool {
	return true // All members can view
}

// CanExecuteWorkflows checks if the member can execute workflows
func (m *Member) CanExecuteWorkflows() bool {
	return m.Role != RoleViewer
}

// UpdateRole updates the member's role
func (m *Member) UpdateRole(role Role) {
	m.Role = role
}

// Invitation represents a workspace invitation
type Invitation struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Email       string     `gorm:"size:255;not null" json:"email"`
	Role        Role       `gorm:"size:20;not null;default:member" json:"role"`
	Token       string     `gorm:"size:255;uniqueIndex;not null" json:"-"`
	InvitedBy   uuid.UUID  `gorm:"type:uuid;not null" json:"invited_by"`
	ExpiresAt   time.Time  `gorm:"not null" json:"expires_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Invitation) TableName() string {
	return "workspace_invitations"
}

// NewInvitation creates a new invitation
func NewInvitation(workspaceID uuid.UUID, email string, role Role, invitedBy uuid.UUID, token string, expiresIn time.Duration) *Invitation {
	return &Invitation{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Email:       email,
		Role:        role,
		Token:       token,
		InvitedBy:   invitedBy,
		ExpiresAt:   time.Now().Add(expiresIn),
		CreatedAt:   time.Now(),
	}
}

// IsExpired checks if the invitation is expired
func (i *Invitation) IsExpired() bool {
	return i.ExpiresAt.Before(time.Now())
}

// IsAccepted checks if the invitation has been accepted
func (i *Invitation) IsAccepted() bool {
	return i.AcceptedAt != nil
}

// IsValid checks if the invitation is still valid
func (i *Invitation) IsValid() bool {
	return !i.IsExpired() && !i.IsAccepted()
}

// Accept marks the invitation as accepted
func (i *Invitation) Accept() {
	now := time.Now()
	i.AcceptedAt = &now
}
