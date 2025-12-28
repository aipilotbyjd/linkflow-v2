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
	ErrNotFound        = errors.New("resource not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidInput    = errors.New("invalid input")
	ErrConflict        = errors.New("resource conflict")
	ErrDuplicateEntry  = errors.New("duplicate entry")
	ErrOperationFailed = errors.New("operation failed")
)

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
	return errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrNotFound)
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
	return errors.Is(err, ErrConflict) || errors.Is(err, ErrDuplicateEntry) || IsUniqueConstraintError(err)
}
