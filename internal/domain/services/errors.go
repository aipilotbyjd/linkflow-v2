package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Domain-specific errors for service layer
var (
	ErrNotFound         = errors.New("resource not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidInput     = errors.New("invalid input")
	ErrConflict         = errors.New("resource conflict")
	ErrDuplicateEntry   = errors.New("duplicate entry")
	ErrRateLimited      = errors.New("rate limit exceeded")
	ErrOperationFailed  = errors.New("operation failed")
	ErrDependencyFailed = errors.New("dependency check failed")
	ErrInternalError    = errors.New("internal error")
)

// ErrorCode represents a unique error code for service errors
type ErrorCode string

const (
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	CodeForbidden        ErrorCode = "FORBIDDEN"
	CodeInvalidInput     ErrorCode = "INVALID_INPUT"
	CodeConflict         ErrorCode = "CONFLICT"
	CodeDuplicateEntry   ErrorCode = "DUPLICATE_ENTRY"
	CodeRateLimited      ErrorCode = "RATE_LIMITED"
	CodeOperationFailed  ErrorCode = "OPERATION_FAILED"
	CodeDependencyFailed ErrorCode = "DEPENDENCY_FAILED"
	CodeInternalError    ErrorCode = "INTERNAL_ERROR"
)

// ServiceError provides rich error context for service layer errors
type ServiceError struct {
	Code       ErrorCode
	Message    string
	Op         string // Operation that failed (e.g., "WorkflowService.Create")
	Resource   string // Resource type (e.g., "workflow", "credential")
	ResourceID string // Resource identifier if available
	Err        error  // Underlying error
}

func (e *ServiceError) Error() string {
	var b strings.Builder

	if e.Op != "" {
		b.WriteString(e.Op)
		b.WriteString(": ")
	}

	if e.Resource != "" {
		b.WriteString(e.Resource)
		if e.ResourceID != "" {
			b.WriteString("[")
			b.WriteString(e.ResourceID)
			b.WriteString("]")
		}
		b.WriteString(": ")
	}

	b.WriteString(e.Message)

	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}

	return b.String()
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is for ServiceError
func (e *ServiceError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.Code == CodeNotFound
	case ErrUnauthorized:
		return e.Code == CodeUnauthorized
	case ErrForbidden:
		return e.Code == CodeForbidden
	case ErrInvalidInput:
		return e.Code == CodeInvalidInput
	case ErrConflict:
		return e.Code == CodeConflict
	case ErrDuplicateEntry:
		return e.Code == CodeDuplicateEntry
	case ErrRateLimited:
		return e.Code == CodeRateLimited
	case ErrOperationFailed:
		return e.Code == CodeOperationFailed
	case ErrDependencyFailed:
		return e.Code == CodeDependencyFailed
	case ErrInternalError:
		return e.Code == CodeInternalError
	}
	return false
}

// NewServiceError creates a new ServiceError
func NewServiceError(code ErrorCode, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

// WithOp sets the operation context
func (e *ServiceError) WithOp(op string) *ServiceError {
	e.Op = op
	return e
}

// WithResource sets the resource context
func (e *ServiceError) WithResource(resource, id string) *ServiceError {
	e.Resource = resource
	e.ResourceID = id
	return e
}

// WithErr wraps an underlying error
func (e *ServiceError) WithErr(err error) *ServiceError {
	e.Err = err
	return e
}

// Error constructors for common scenarios

// NotFoundError creates a not found error
func NotFoundError(resource, id string) *ServiceError {
	return &ServiceError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		Resource:   resource,
		ResourceID: id,
	}
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *ServiceError {
	return &ServiceError{
		Code:    CodeUnauthorized,
		Message: message,
	}
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string) *ServiceError {
	return &ServiceError{
		Code:    CodeForbidden,
		Message: message,
	}
}

// InvalidInputError creates an invalid input error
func InvalidInputError(message string) *ServiceError {
	return &ServiceError{
		Code:    CodeInvalidInput,
		Message: message,
	}
}

// ConflictError creates a conflict error
func ConflictError(resource, message string) *ServiceError {
	return &ServiceError{
		Code:     CodeConflict,
		Message:  message,
		Resource: resource,
	}
}

// DuplicateEntryError creates a duplicate entry error
func DuplicateEntryError(resource, field string) *ServiceError {
	return &ServiceError{
		Code:     CodeDuplicateEntry,
		Message:  fmt.Sprintf("%s with this %s already exists", resource, field),
		Resource: resource,
	}
}

// WrapError wraps an error with operation context
func WrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// Database error type checkers

// IsNotFoundError checks if an error represents a "not found" condition.
// This centralizes the gorm dependency for not-found checks.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		return svcErr.Code == CodeNotFound
	}
	return false
}

// IsUniqueConstraintError checks if error is a unique constraint violation
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL unique_violation error code
		return pgErr.Code == "23505"
	}
	// Check for common error message patterns
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "UNIQUE constraint failed")
}

// IsForeignKeyError checks if error is a foreign key constraint violation
func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL foreign_key_violation error code
		return pgErr.Code == "23503"
	}
	errStr := err.Error()
	return strings.Contains(errStr, "foreign key constraint") ||
		strings.Contains(errStr, "FOREIGN KEY constraint failed")
}

// IsConnectionError checks if error is a database connection error
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

// IsDeadlockError checks if error is a database deadlock
func IsDeadlockError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL deadlock_detected error code
		return pgErr.Code == "40P01"
	}
	return strings.Contains(err.Error(), "deadlock")
}

// IsConflictError checks if error represents a conflict
func IsConflictError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrConflict) || errors.Is(err, ErrDuplicateEntry) {
		return true
	}
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		return svcErr.Code == CodeConflict || svcErr.Code == CodeDuplicateEntry
	}
	return IsUniqueConstraintError(err)
}

// IsUnauthorizedError checks if error represents unauthorized access
func IsUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrUnauthorized) {
		return true
	}
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		return svcErr.Code == CodeUnauthorized
	}
	return false
}

// IsForbiddenError checks if error represents forbidden access
func IsForbiddenError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrForbidden) {
		return true
	}
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		return svcErr.Code == CodeForbidden
	}
	return false
}

// IsInvalidInputError checks if error represents invalid input
func IsInvalidInputError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrInvalidInput) {
		return true
	}
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		return svcErr.Code == CodeInvalidInput
	}
	return false
}

// MapDatabaseError maps database errors to service errors
func MapDatabaseError(err error, resource string) error {
	if err == nil {
		return nil
	}

	if IsNotFoundError(err) {
		return NotFoundError(resource, "")
	}

	if IsUniqueConstraintError(err) {
		return DuplicateEntryError(resource, "field")
	}

	if IsForeignKeyError(err) {
		return &ServiceError{
			Code:     CodeDependencyFailed,
			Message:  "referenced resource does not exist or cannot be deleted",
			Resource: resource,
			Err:      err,
		}
	}

	if IsDeadlockError(err) {
		return &ServiceError{
			Code:     CodeConflict,
			Message:  "concurrent modification detected, please retry",
			Resource: resource,
			Err:      err,
		}
	}

	return err
}
