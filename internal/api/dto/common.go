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

// Response is the standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Error     *ErrorData  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorData contains error details
type ErrorData struct {
	Code    string                      `json:"code"`
	Message string                      `json:"message"`
	Details []validator.ValidationError `json:"details,omitempty"`
}

// Meta contains pagination metadata
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

// JSON sends a JSON response
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

// JSONWithMeta sends a JSON response with pagination metadata
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

// ErrorResponse sends an error response
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	code := statusToErrorCode(status)
	errorWithCode(w, status, code, message)
}

// ValidationErrorResponse sends a validation error response
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

// Convenience response helpers

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
