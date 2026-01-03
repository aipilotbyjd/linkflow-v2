package dto

import "github.com/linkflow-ai/linkflow/internal/domain/models"

// Workspace requests

type CreateWorkspaceRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Slug        string  `json:"slug" validate:"required,slug"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Timezone    *string `json:"timezone,omitempty" validate:"omitempty,max=50"`
}

type UpdateWorkspaceRequest struct {
	Name        *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	LogoURL     *string     `json:"logo_url,omitempty" validate:"omitempty,url"`
	Timezone    *string     `json:"timezone,omitempty" validate:"omitempty,max=50"`
	Settings    models.JSON `json:"settings,omitempty"`
}

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member viewer"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member viewer"`
}

// Workspace responses

type WorkspaceResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	Timezone    string  `json:"timezone"`
	PlanID      string  `json:"plan_id"`
	CreatedAt   int64   `json:"created_at"`
}

type WorkspaceMemberResponse struct {
	ID        string       `json:"id"`
	User      UserResponse `json:"user"`
	Role      string       `json:"role"`
	JoinedAt  *int64       `json:"joined_at,omitempty"`
	InvitedAt *int64       `json:"invited_at,omitempty"`
}
