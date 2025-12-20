package dto

import "github.com/linkflow-ai/linkflow/internal/domain/models"

// Auth
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type SetupMFARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

type VerifyMFARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// User
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
}

// Workspace
type CreateWorkspaceRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Slug        string  `json:"slug" validate:"required,slug"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

type UpdateWorkspaceRequest struct {
	Name        *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	LogoURL     *string     `json:"logo_url,omitempty" validate:"omitempty,url"`
	Settings    models.JSON `json:"settings,omitempty"`
}

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member viewer"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member viewer"`
}

// Workflow
type CreateWorkflowRequest struct {
	Name        string      `json:"name" validate:"required,min=1,max=255"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=1000"`
	Nodes       models.JSON `json:"nodes" validate:"required"`
	Connections models.JSON `json:"connections" validate:"required"`
	Settings    models.JSON `json:"settings,omitempty"`
	Tags        []string    `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
}

type UpdateWorkflowRequest struct {
	Name        *string     `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=1000"`
	Nodes       models.JSON `json:"nodes,omitempty"`
	Connections models.JSON `json:"connections,omitempty"`
	Settings    models.JSON `json:"settings,omitempty"`
	Tags        []string    `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
}

type ExecuteWorkflowRequest struct {
	InputData models.JSON `json:"input_data,omitempty"`
}

type CloneWorkflowRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

// Credential
type CreateCredentialRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=100"`
	Type        string                 `json:"type" validate:"required,oneof=api_key oauth2 basic bearer custom"`
	Data        models.CredentialData  `json:"data" validate:"required"`
	Description *string                `json:"description,omitempty" validate:"omitempty,max=500"`
}

type UpdateCredentialRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Data        *models.CredentialData `json:"data,omitempty"`
	Description *string                `json:"description,omitempty" validate:"omitempty,max=500"`
}

// Schedule
type CreateScheduleRequest struct {
	WorkflowID     string      `json:"workflow_id" validate:"required,uuid"`
	Name           string      `json:"name" validate:"required,min=1,max=100"`
	Description    *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	CronExpression string      `json:"cron_expression" validate:"required,cron"`
	Timezone       string      `json:"timezone" validate:"required"`
	InputData      models.JSON `json:"input_data,omitempty"`
}

type UpdateScheduleRequest struct {
	Name           *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description    *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	CronExpression *string     `json:"cron_expression,omitempty" validate:"omitempty,cron"`
	Timezone       *string     `json:"timezone,omitempty"`
	InputData      models.JSON `json:"input_data,omitempty"`
}

// Billing
type CreateSubscriptionRequest struct {
	PlanID       string `json:"plan_id" validate:"required"`
	BillingCycle string `json:"billing_cycle" validate:"required,oneof=monthly yearly"`
	PaymentToken string `json:"payment_token,omitempty"`
}

// Pagination
type PaginationRequest struct {
	Page    int    `json:"page" validate:"omitempty,min=1"`
	PerPage int    `json:"per_page" validate:"omitempty,min=1,max=100"`
	Search  string `json:"search,omitempty"`
	OrderBy string `json:"order_by,omitempty"`
	Order   string `json:"order,omitempty" validate:"omitempty,oneof=asc desc"`
}
