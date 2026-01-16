package user

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72 // bcrypt limit
	BcryptCost        = 10
)

// Password represents a hashed password
type Password struct {
	hash string
}

// NewPassword creates a new Password from a plain text password
func NewPassword(plaintext string) (Password, error) {
	if err := ValidatePasswordStrength(plaintext); err != nil {
		return Password{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), BcryptCost)
	if err != nil {
		return Password{}, ErrPasswordHashFailed
	}

	return Password{hash: string(hash)}, nil
}

// NewPasswordFromHash creates a Password from an existing hash
func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

// Hash returns the password hash
func (p Password) Hash() string {
	return p.hash
}

// Verify checks if the given plaintext matches the password hash
func (p Password) Verify(plaintext string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintext))
	return err == nil
}

// ValidatePasswordStrength validates password strength requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrPasswordWeak
	}

	return nil
}
