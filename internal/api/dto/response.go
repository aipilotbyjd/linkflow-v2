package dto

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

// Error codes for consistent API responses
const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeInternalServer = "INTERNAL_SERVER_ERROR"
	ErrCodeTooManyRequest = "TOO_MANY_REQUESTS"
	ErrCodeServiceUnavail = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout        = "TIMEOUT"
)

// Common service errors for mapping
var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("resource conflict")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInternalError = errors.New("internal server error")
)

type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Error     *ErrorData  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type ErrorData struct {
	Code    string                      `json:"code"`
	Message string                      `json:"message"`
	Details []validator.ValidationError `json:"details,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// RequestIDKey is the context key for request ID
const RequestIDKey = "request_id"

// getRequestID extracts request ID from response header if set
func getRequestID(w http.ResponseWriter) string {
	return w.Header().Get("X-Request-ID")
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success:   status >= 200 && status < 300,
		Data:      data,
		RequestID: getRequestID(w),
		Timestamp: time.Now().Unix(),
	}

	_ = json.NewEncoder(w).Encode(response)
}

func JSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success:   status >= 200 && status < 300,
		Data:      data,
		Meta:      meta,
		RequestID: getRequestID(w),
		Timestamp: time.Now().Unix(),
	}

	_ = json.NewEncoder(w).Encode(response)
}

func errorWithCode(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success:   false,
		RequestID: getRequestID(w),
		Timestamp: time.Now().Unix(),
		Error: &ErrorData{
			Code:    code,
			Message: message,
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	code := statusToErrorCode(status)
	errorWithCode(w, status, code, message)
}

func ValidationErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := Response{
		Success:   false,
		RequestID: getRequestID(w),
		Timestamp: time.Now().Unix(),
		Error: &ErrorData{
			Code:    ErrCodeValidation,
			Message: "Validation failed",
			Details: validator.FormatErrors(err),
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}

// WorkflowValidationError represents a workflow-specific validation error
type WorkflowValidationError struct {
	Field   string `json:"field"`
	NodeID  string `json:"node_id,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WorkflowValidationErrorResponse returns a workflow validation error response
func WorkflowValidationErrorResponse(w http.ResponseWriter, errors []WorkflowValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	// Convert to validation error format
	details := make([]validator.ValidationError, len(errors))
	for i, e := range errors {
		field := e.Field
		if e.NodeID != "" {
			field = "node:" + e.NodeID + "." + e.Field
		}
		details[i] = validator.ValidationError{
			Field:   field,
			Message: e.Message,
		}
	}

	response := Response{
		Success:   false,
		RequestID: getRequestID(w),
		Timestamp: time.Now().Unix(),
		Error: &ErrorData{
			Code:    "WORKFLOW_VALIDATION_ERROR",
			Message: "Workflow validation failed",
			Details: details,
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}

// Convenience helpers (Laravel-style trait methods)

func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

func Accepted(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusAccepted, data)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func BadRequest(w http.ResponseWriter, message string) {
	errorWithCode(w, http.StatusBadRequest, ErrCodeBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	errorWithCode(w, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

func Forbidden(w http.ResponseWriter, message string) {
	errorWithCode(w, http.StatusForbidden, ErrCodeForbidden, message)
}

func NotFound(w http.ResponseWriter, resource string) {
	message := resource + " not found"
	errorWithCode(w, http.StatusNotFound, ErrCodeNotFound, message)
}

func Conflict(w http.ResponseWriter, message string) {
	errorWithCode(w, http.StatusConflict, ErrCodeConflict, message)
}

func TooManyRequests(w http.ResponseWriter, message string) {
	errorWithCode(w, http.StatusTooManyRequests, ErrCodeTooManyRequest, message)
}

func InternalServerError(w http.ResponseWriter, message string) {
	errorWithCode(w, http.StatusInternalServerError, ErrCodeInternalServer, message)
}

func ServiceUnavailable(w http.ResponseWriter, message string) {
	errorWithCode(w, http.StatusServiceUnavailable, ErrCodeServiceUnavail, message)
}

// HandleServiceError maps service-layer errors to appropriate HTTP responses
func HandleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		NotFound(w, "Resource")
	case errors.Is(err, ErrUnauthorized):
		Unauthorized(w, err.Error())
	case errors.Is(err, ErrForbidden):
		Forbidden(w, err.Error())
	case errors.Is(err, ErrConflict):
		Conflict(w, err.Error())
	case errors.Is(err, ErrInvalidInput):
		BadRequest(w, err.Error())
	default:
		InternalServerError(w, "An unexpected error occurred")
	}
}

// statusToErrorCode maps HTTP status codes to error codes
func statusToErrorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return ErrCodeBadRequest
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	case http.StatusTooManyRequests:
		return ErrCodeTooManyRequest
	case http.StatusInternalServerError:
		return ErrCodeInternalServer
	case http.StatusServiceUnavailable:
		return ErrCodeServiceUnavail
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return ErrCodeTimeout
	default:
		return http.StatusText(status)
	}
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
	// Core counts
	Executions   int   `json:"executions"`
	Workflows    int   `json:"workflows"`
	Members      int   `json:"members"`
	Credentials  int   `json:"credentials"`
	StorageBytes int64 `json:"storage_bytes"`

	// Credits (Make.com style)
	CreditsUsed      int `json:"credits_used"`
	CreditsIncluded  int `json:"credits_included"`
	CreditsPurchased int `json:"credits_purchased"`
	CreditsRemaining int `json:"credits_remaining"`

	// Execution details
	ExecutionsSuccess int `json:"executions_success"`
	ExecutionsFailed  int `json:"executions_failed"`
	Operations        int `json:"operations"`

	// Webhooks & Schedules
	WebhooksCalled     int `json:"webhooks_called"`
	SchedulesTriggered int `json:"schedules_triggered"`
	Schedules          int `json:"schedules"`
	Webhooks           int `json:"webhooks"`

	// Data transfer
	DataTransferIn  int64 `json:"data_transfer_in"`
	DataTransferOut int64 `json:"data_transfer_out"`

	// Overage
	OverageCredits int `json:"overage_credits"`
	OverageCost    int `json:"overage_cost"`

	// Period
	PeriodStart int64 `json:"period_start"`
	PeriodEnd   int64 `json:"period_end"`
}
