package dto

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorData struct {
	Code    string                      `json:"code"`
	Message string                      `json:"message"`
	Details []validator.ValidationError `json:"details,omitempty"`
}

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func JSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: status >= 200 && status < 300,
		Data:    data,
		Meta:    meta,
	}

	json.NewEncoder(w).Encode(response)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: false,
		Error: &ErrorData{
			Code:    http.StatusText(status),
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func ValidationErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := Response{
		Success: false,
		Error: &ErrorData{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: validator.FormatErrors(err),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

func Accepted(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusAccepted, data)
}

// Auth responses
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    int64         `json:"expires_at"`
}

type MFARequiredResponse struct {
	RequiresMFA bool   `json:"requires_mfa"`
	Message     string `json:"message"`
}

type MFASetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

// User responses
type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      *string `json:"username,omitempty"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	EmailVerified bool    `json:"email_verified"`
	MFAEnabled    bool    `json:"mfa_enabled"`
	CreatedAt     int64   `json:"created_at"`
}

// Workspace responses
type WorkspaceResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
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

// Workflow responses
type WorkflowResponse struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    *string     `json:"description,omitempty"`
	Status         string      `json:"status"`
	Version        int         `json:"version"`
	Nodes          interface{} `json:"nodes"`
	Connections    interface{} `json:"connections"`
	Settings       interface{} `json:"settings,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	ExecutionCount int         `json:"execution_count"`
	LastExecutedAt *int64      `json:"last_executed_at,omitempty"`
	CreatedAt      int64       `json:"created_at"`
	UpdatedAt      int64       `json:"updated_at"`
}

type WorkflowVersionResponse struct {
	ID            string      `json:"id"`
	Version       int         `json:"version"`
	Nodes         interface{} `json:"nodes"`
	Connections   interface{} `json:"connections"`
	Settings      interface{} `json:"settings,omitempty"`
	ChangeMessage *string     `json:"change_message,omitempty"`
	CreatedAt     int64       `json:"created_at"`
}

// Execution responses
type ExecutionResponse struct {
	ID              string      `json:"id"`
	WorkflowID      string      `json:"workflow_id"`
	WorkflowVersion int         `json:"workflow_version"`
	Status          string      `json:"status"`
	TriggerType     string      `json:"trigger_type"`
	InputData       interface{} `json:"input_data,omitempty"`
	OutputData      interface{} `json:"output_data,omitempty"`
	ErrorMessage    *string     `json:"error_message,omitempty"`
	ErrorNodeID     *string     `json:"error_node_id,omitempty"`
	NodesTotal      int         `json:"nodes_total"`
	NodesCompleted  int         `json:"nodes_completed"`
	QueuedAt        int64       `json:"queued_at"`
	StartedAt       *int64      `json:"started_at,omitempty"`
	CompletedAt     *int64      `json:"completed_at,omitempty"`
}

type NodeExecutionResponse struct {
	ID           string      `json:"id"`
	NodeID       string      `json:"node_id"`
	NodeType     string      `json:"node_type"`
	NodeName     *string     `json:"node_name,omitempty"`
	Status       string      `json:"status"`
	InputData    interface{} `json:"input_data,omitempty"`
	OutputData   interface{} `json:"output_data,omitempty"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	DurationMs   *int        `json:"duration_ms,omitempty"`
	StartedAt    *int64      `json:"started_at,omitempty"`
	CompletedAt  *int64      `json:"completed_at,omitempty"`
}

// Credential responses
type CredentialResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description *string `json:"description,omitempty"`
	LastUsedAt  *int64  `json:"last_used_at,omitempty"`
	CreatedAt   int64   `json:"created_at"`
}

// Schedule responses
type ScheduleResponse struct {
	ID              string      `json:"id"`
	WorkflowID      string      `json:"workflow_id"`
	Name            string      `json:"name"`
	Description     *string     `json:"description,omitempty"`
	CronExpression  string      `json:"cron_expression"`
	Timezone        string      `json:"timezone"`
	IsActive        bool        `json:"is_active"`
	InputData       interface{} `json:"input_data,omitempty"`
	NextRunAt       *int64      `json:"next_run_at,omitempty"`
	LastRunAt       *int64      `json:"last_run_at,omitempty"`
	RunCount        int         `json:"run_count"`
	CreatedAt       int64       `json:"created_at"`
}

// Billing responses
type PlanResponse struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Description      *string     `json:"description,omitempty"`
	PriceMonthly     int         `json:"price_monthly"`
	PriceYearly      int         `json:"price_yearly"`
	ExecutionsLimit  int         `json:"executions_limit"`
	WorkflowsLimit   int         `json:"workflows_limit"`
	MembersLimit     int         `json:"members_limit"`
	CredentialsLimit int         `json:"credentials_limit"`
	Features         interface{} `json:"features,omitempty"`
}

type SubscriptionResponse struct {
	ID                 string `json:"id"`
	PlanID             string `json:"plan_id"`
	Status             string `json:"status"`
	BillingCycle       string `json:"billing_cycle"`
	CurrentPeriodStart int64  `json:"current_period_start"`
	CurrentPeriodEnd   int64  `json:"current_period_end"`
	CancelAt           *int64 `json:"cancel_at,omitempty"`
}

type UsageResponse struct {
	Executions   int   `json:"executions"`
	Workflows    int   `json:"workflows"`
	Members      int   `json:"members"`
	Credentials  int   `json:"credentials"`
	StorageBytes int64 `json:"storage_bytes"`
	PeriodStart  int64 `json:"period_start"`
	PeriodEnd    int64 `json:"period_end"`
}
