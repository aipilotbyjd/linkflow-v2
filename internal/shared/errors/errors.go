package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Standard domain errors
var (
	ErrNotFound        = errors.New("resource not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidInput    = errors.New("invalid input")
	ErrConflict        = errors.New("resource conflict")
	ErrDuplicateEntry  = errors.New("duplicate entry")
	ErrOperationFailed = errors.New("operation failed")
	ErrValidation      = errors.New("validation failed")
	ErrRateLimited     = errors.New("rate limited")
	ErrTimeout         = errors.New("operation timeout")
	ErrInternal        = errors.New("internal error")
)

// DomainError represents a domain-specific error with context
type DomainError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	Cause      error             `json:"-"`
	StatusCode int               `json:"-"`
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

func (e *DomainError) WithDetail(key, value string) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

func (e *DomainError) WithCause(err error) *DomainError {
	e.Cause = err
	return e
}

// Error constructors
func NewNotFoundError(resource string) *DomainError {
	return &DomainError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
		Cause:      ErrNotFound,
	}
}

func NewUnauthorizedError(message string) *DomainError {
	if message == "" {
		message = "unauthorized"
	}
	return &DomainError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Cause:      ErrUnauthorized,
	}
}

func NewForbiddenError(message string) *DomainError {
	if message == "" {
		message = "access denied"
	}
	return &DomainError{
		Code:       "FORBIDDEN",
		Message:    message,
		StatusCode: http.StatusForbidden,
		Cause:      ErrForbidden,
	}
}

func NewValidationError(message string) *DomainError {
	return &DomainError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Cause:      ErrValidation,
	}
}

func NewConflictError(message string) *DomainError {
	return &DomainError{
		Code:       "CONFLICT",
		Message:    message,
		StatusCode: http.StatusConflict,
		Cause:      ErrConflict,
	}
}

func NewDuplicateError(field string) *DomainError {
	return &DomainError{
		Code:       "DUPLICATE_ENTRY",
		Message:    fmt.Sprintf("%s already exists", field),
		StatusCode: http.StatusConflict,
		Cause:      ErrDuplicateEntry,
	}
}

func NewInternalError(message string) *DomainError {
	if message == "" {
		message = "internal server error"
	}
	return &DomainError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      ErrInternal,
	}
}

func NewRateLimitError() *DomainError {
	return &DomainError{
		Code:       "RATE_LIMITED",
		Message:    "too many requests",
		StatusCode: http.StatusTooManyRequests,
		Cause:      ErrRateLimited,
	}
}

// Wrap wraps an error with operation context
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// Error type checkers
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrNotFound)
}

func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}

func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation) || errors.Is(err, ErrInvalidInput)
}

func IsConflictError(err error) bool {
	return errors.Is(err, ErrConflict) || errors.Is(err, ErrDuplicateEntry) || IsUniqueConstraintError(err)
}

// Database error checkers
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "UNIQUE constraint failed")
}

func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	errStr := err.Error()
	return strings.Contains(errStr, "foreign key constraint") ||
		strings.Contains(errStr, "FOREIGN KEY constraint failed")
}

func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection timed out") ||
		strings.Contains(errStr, "EOF")
}

func IsDeadlockError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40P01"
	}
	return strings.Contains(err.Error(), "deadlock")
}

// GetHTTPStatus extracts HTTP status code from error
func GetHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.StatusCode
	}
	if IsNotFoundError(err) {
		return http.StatusNotFound
	}
	if IsUnauthorizedError(err) {
		return http.StatusUnauthorized
	}
	if IsForbiddenError(err) {
		return http.StatusForbidden
	}
	if IsValidationError(err) {
		return http.StatusBadRequest
	}
	if IsConflictError(err) {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
