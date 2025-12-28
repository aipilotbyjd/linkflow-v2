package services

import (
	"errors"

	"gorm.io/gorm"
)

// IsNotFoundError checks if an error represents a "not found" condition.
// This centralizes the gorm dependency for not-found checks.
func IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
