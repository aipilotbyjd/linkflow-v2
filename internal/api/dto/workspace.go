package dto

import "github.com/linkflow-ai/linkflow/internal/domain/models"

// Workspace requests

type CreateWorkspaceRequest struct {
	Name         string  `json:"name" validate:"required,min=1,max=100"`
	Slug         string  `json:"slug" validate:"required,slug"`
	Description  *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Timezone     *string `json:"timezone,omitempty" validate:"omitempty,max=50"`
	Language     *string `json:"language,omitempty" validate:"omitempty,max=10"`
	Currency     *string `json:"currency,omitempty" validate:"omitempty,len=3"`
	Country      *string `json:"country,omitempty" validate:"omitempty,len=2"`
	Industry     *string `json:"industry,omitempty" validate:"omitempty,max=50"`
	CompanySize  *string `json:"company_size,omitempty" validate:"omitempty,oneof=1-10 11-50 51-200 201-500 501-1000 1000+"`
	Website      *string `json:"website,omitempty" validate:"omitempty,url"`
	BillingEmail *string `json:"billing_email,omitempty" validate:"omitempty,email"`
}

type UpdateWorkspaceRequest struct {
	Name         *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description  *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	LogoURL      *string     `json:"logo_url,omitempty" validate:"omitempty,url"`
	Timezone     *string     `json:"timezone,omitempty" validate:"omitempty,max=50"`
	Language     *string     `json:"language,omitempty" validate:"omitempty,max=10"`
	Currency     *string     `json:"currency,omitempty" validate:"omitempty,len=3"`
	Country      *string     `json:"country,omitempty" validate:"omitempty,len=2"`
	Industry     *string     `json:"industry,omitempty" validate:"omitempty,max=50"`
	CompanySize  *string     `json:"company_size,omitempty" validate:"omitempty,oneof=1-10 11-50 51-200 201-500 501-1000 1000+"`
	Website      *string     `json:"website,omitempty" validate:"omitempty,url"`
	BillingEmail *string     `json:"billing_email,omitempty" validate:"omitempty,email"`
	Settings     models.JSON `json:"settings,omitempty"`
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
	ID           string      `json:"id"`
	OwnerID      string      `json:"owner_id"`
	Name         string      `json:"name"`
	Slug         string      `json:"slug"`
	Description  *string     `json:"description,omitempty"`
	LogoURL      *string     `json:"logo_url,omitempty"`
	Website      *string     `json:"website,omitempty"`
	Timezone     string      `json:"timezone"`
	Language     string      `json:"language"`
	Currency     string      `json:"currency"`
	Country      *string     `json:"country,omitempty"`
	Industry     *string     `json:"industry,omitempty"`
	CompanySize  *string     `json:"company_size,omitempty"`
	BillingEmail *string     `json:"billing_email,omitempty"`
	Settings     interface{} `json:"settings,omitempty"`
	PlanID       string      `json:"plan_id"`
	CreatedAt    int64       `json:"created_at"`
	UpdatedAt    int64       `json:"updated_at"`
}

type WorkspaceMemberResponse struct {
	ID        string       `json:"id"`
	User      UserResponse `json:"user"`
	Role      string       `json:"role"`
	JoinedAt  *int64       `json:"joined_at,omitempty"`
	InvitedAt *int64       `json:"invited_at,omitempty"`
}
