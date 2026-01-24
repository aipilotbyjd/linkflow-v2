package common

import (
	"encoding/json"
	stderrors "errors"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/shared/errors"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"github.com/rs/zerolog/log"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
	Meta    *MetaData   `json:"meta,omitempty"`
}

// ErrorData represents error information in response
type ErrorData struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []ValidationDetail `json:"details,omitempty"`
}

// ValidationDetail represents a single field validation error
type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// MetaData represents pagination/meta information
type MetaData struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"page_size,omitempty"`
	TotalItems int64 `json:"total_items,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	HasMore    bool  `json:"has_more,omitempty"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Success writes a success response
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created writes a created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent writes a no content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// List writes a paginated list response
func List(w http.ResponseWriter, data interface{}, pagination types.PageResponse) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &MetaData{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: pagination.TotalItems,
			TotalPages: pagination.TotalPages,
			HasMore:    pagination.HasMore,
		},
	})
}

// Error writes an error response (supports both 3 and 4 argument forms)
func Error(w http.ResponseWriter, status int, codeOrMessage string, messageOptional ...string) {
	var code, message string
	if len(messageOptional) > 0 {
		code = codeOrMessage
		message = messageOptional[0]
	} else {
		code = "ERROR"
		message = codeOrMessage
	}
	JSON(w, status, Response{
		Success: false,
		Error: &ErrorData{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDetails writes an error response with field validation details
func ErrorWithDetails(w http.ResponseWriter, status int, code, message string, details []ValidationDetail) {
	JSON(w, status, Response{
		Success: false,
		Error: &ErrorData{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// ValidationErrors writes a validation error response
func ValidationErrors(w http.ResponseWriter, errors []ValidationDetail) {
	JSON(w, http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorData{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: errors,
		},
	})
}

// ValidateRequest validates a request struct and writes error response if invalid
// Returns true if valid, false if invalid (and response already written)
func ValidateRequest(w http.ResponseWriter, req interface{}, validate func(interface{}) []struct{ Field, Message string }) bool {
	errors := validate(req)
	if len(errors) == 0 {
		return true
	}
	details := make([]ValidationDetail, len(errors))
	for i, e := range errors {
		details[i] = ValidationDetail{Field: e.Field, Message: e.Message}
	}
	ValidationErrors(w, details)
	return false
}

// BadRequest writes a bad request error
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized writes an unauthorized error
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "unauthorized"
	}
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden writes a forbidden error
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "access denied"
	}
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound writes a not found error
func NotFound(w http.ResponseWriter, resource string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", resource+" not found")
}

// Conflict writes a conflict error
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}

// InternalError writes an internal server error
func InternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "internal server error"
	}
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// RateLimited writes a rate limit error
func RateLimited(w http.ResponseWriter) {
	Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
}

// ValidationError writes a validation error (deprecated, use ValidationErrors)
func ValidationError(w http.ResponseWriter, message string, details map[string]string) {
	detailList := make([]ValidationDetail, 0, len(details))
	for field, msg := range details {
		detailList = append(detailList, ValidationDetail{Field: field, Message: msg})
	}
	ErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", message, detailList)
}

// HandleError handles an error and writes appropriate response
func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	status := errors.GetHTTPStatus(err)

	var domainErr *errors.DomainError
	if stderrors.As(err, &domainErr) {
		var detailList []ValidationDetail
		if domainErr.Details != nil {
			detailList = make([]ValidationDetail, 0, len(domainErr.Details))
			for field, msg := range domainErr.Details {
				detailList = append(detailList, ValidationDetail{Field: field, Message: msg})
			}
		}
		ErrorWithDetails(w, status, domainErr.Code, domainErr.Message, detailList)
		return
	}

	// Map common error messages to proper responses
	errMsg := err.Error()
	switch errMsg {
	// Auth errors
	case "invalid credentials":
		Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	case "account is locked":
		Error(w, http.StatusForbidden, "ACCOUNT_LOCKED", "Account is temporarily locked due to too many failed login attempts")
	case "account is suspended":
		Error(w, http.StatusForbidden, "ACCOUNT_SUSPENDED", "Account has been suspended")
	case "email already exists":
		Error(w, http.StatusConflict, "EMAIL_EXISTS", "An account with this email already exists")
	case "workspace slug already exists":
		Error(w, http.StatusConflict, "SLUG_EXISTS", "A workspace with this name already exists")
	case "user not found":
		Error(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case "email not verified":
		Error(w, http.StatusForbidden, "EMAIL_NOT_VERIFIED", "Please verify your email address")
	case "MFA verification required":
		Error(w, http.StatusUnauthorized, "MFA_REQUIRED", "Multi-factor authentication is required")
	case "invalid MFA code":
		Error(w, http.StatusUnauthorized, "MFA_INVALID", "Invalid MFA code")
	case "password is required":
		Error(w, http.StatusBadRequest, "PASSWORD_REQUIRED", "Password is required")
	case "password must be at least 8 characters":
		Error(w, http.StatusBadRequest, "PASSWORD_TOO_SHORT", "Password must be at least 8 characters")
	case "password does not meet requirements":
		Error(w, http.StatusBadRequest, "PASSWORD_WEAK", "Password must contain uppercase, lowercase, number, and special character")
	case "session not found", "session expired", "session revoked":
		Error(w, http.StatusUnauthorized, "SESSION_INVALID", "Session is invalid or expired")
	case "too many login attempts":
		Error(w, http.StatusTooManyRequests, "TOO_MANY_ATTEMPTS", "Too many login attempts. Please try again later")

	// Workflow errors
	case "workflow has no nodes":
		Error(w, http.StatusBadRequest, "WORKFLOW_EMPTY", "Workflow must have at least one node before activation")
	case "workflow requires a trigger node":
		Error(w, http.StatusBadRequest, "WORKFLOW_NO_TRIGGER", "Workflow must have a trigger node to be activated")
	case "workflow is not active":
		Error(w, http.StatusBadRequest, "WORKFLOW_INACTIVE", "Workflow is not active")
	case "workflow is already active":
		Error(w, http.StatusConflict, "WORKFLOW_ALREADY_ACTIVE", "Workflow is already active")
	case "version not found":
		Error(w, http.StatusNotFound, "VERSION_NOT_FOUND", "Workflow version not found")
	case "active scenarios limit exceeded":
		Error(w, http.StatusForbidden, "LIMIT_EXCEEDED", "Active scenarios limit exceeded for your plan. Please upgrade or deactivate other workflows.")

	// Webhook errors
	case "webhook path already exists":
		Error(w, http.StatusConflict, "WEBHOOK_PATH_EXISTS", "A webhook with this path already exists")
	case "webhook endpoint not found":
		Error(w, http.StatusNotFound, "WEBHOOK_NOT_FOUND", "Webhook endpoint not found")
	case "webhook endpoint is not active", "webhook endpoint is inactive":
		Error(w, http.StatusGone, "WEBHOOK_INACTIVE", "Webhook endpoint is inactive")
	case "method not allowed":
		Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed for this webhook")

	// Billing errors
	case "subscription not found":
		Error(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "No subscription found for this workspace")
	case "usage record not found":
		Error(w, http.StatusNotFound, "USAGE_NOT_FOUND", "Usage record not found")
	case "invoice not found":
		Error(w, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")

	// Execution errors
	case "execution not found":
		Error(w, http.StatusNotFound, "EXECUTION_NOT_FOUND", "Execution not found")
	case "execution already completed":
		Error(w, http.StatusBadRequest, "EXECUTION_COMPLETED", "Execution has already completed")
	case "execution is not running":
		Error(w, http.StatusBadRequest, "EXECUTION_NOT_RUNNING", "Execution is not running")
	case "cannot cancel execution":
		Error(w, http.StatusBadRequest, "CANNOT_CANCEL", "Cannot cancel this execution")
	case "cannot retry execution":
		Error(w, http.StatusBadRequest, "CANNOT_RETRY", "Cannot retry this execution")

	default:
		// Fall back to generic handling
		switch {
		case errors.IsNotFoundError(err):
			NotFound(w, "resource")
		case errors.IsUnauthorizedError(err):
			Unauthorized(w, err.Error())
		case errors.IsForbiddenError(err):
			Forbidden(w, err.Error())
		case errors.IsValidationError(err):
			BadRequest(w, err.Error())
		case errors.IsConflictError(err):
			Conflict(w, err.Error())
		default:
			log.Error().Err(err).Msg("Unhandled server error")
			InternalError(w, "")
		}
	}
}
