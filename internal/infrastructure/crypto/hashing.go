package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultBcryptCost = 12
	MinBcryptCost     = 10
	MaxBcryptCost     = 14
)

// Hasher is an alias for PasswordHasher for convenience
type Hasher = PasswordHasher

// NewHasher creates a new hasher
func NewHasher(cost int) *Hasher {
	return NewPasswordHasher(cost)
}

// PasswordHasher provides password hashing and verification
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < MinBcryptCost {
		cost = DefaultBcryptCost
	}
	if cost > MaxBcryptCost {
		cost = MaxBcryptCost
	}
	return &PasswordHasher{cost: cost}
}

// Hash hashes a password using bcrypt
func (h *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// Verify verifies a password against a hash
func (h *PasswordHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPassword hashes a password using the default cost
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultBcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword verifies a password against a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomToken generates a random hex token
func GenerateRandomToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte(fmt.Sprintf("%d", length)))
	}
	return hex.EncodeToString(bytes)
}

// HashSHA256 creates a SHA256 hash of the input
func HashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HashSHA256Bytes creates a SHA256 hash of byte input
func HashSHA256Bytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SecureCompare performs a constant-time comparison
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
