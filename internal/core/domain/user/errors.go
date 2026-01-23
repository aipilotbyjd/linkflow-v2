package user

import "errors"

// Domain-specific errors for user aggregate
var (
	ErrUserNotFound            = errors.New("user not found")
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrUsernameAlreadyExists   = errors.New("username already exists")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrAccountLocked           = errors.New("account is locked")
	ErrAccountSuspended        = errors.New("account is suspended")
	ErrEmailNotVerified        = errors.New("email not verified")
	ErrMFARequired             = errors.New("MFA verification required")
	ErrMFAInvalid              = errors.New("invalid MFA code")
	ErrMFAAlreadyEnabled       = errors.New("MFA already enabled")
	ErrMFANotEnabled           = errors.New("MFA not enabled")
	ErrPasswordTooWeak         = errors.New("password does not meet requirements")
	ErrPasswordSameAsOld       = errors.New("new password must be different from old password")
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionExpired          = errors.New("session expired")
	ErrSessionRevoked          = errors.New("session revoked")
	ErrAPIKeyNotFound          = errors.New("API key not found")
	ErrAPIKeyExpired           = errors.New("API key expired")
	ErrAPIKeyRevoked           = errors.New("API key revoked")
	ErrAPIKeyInvalidScope      = errors.New("API key does not have required scope")
	ErrResetTokenNotFound      = errors.New("reset token not found")
	ErrResetTokenExpired       = errors.New("reset token expired")
	ErrResetTokenUsed          = errors.New("reset token already used")
	ErrOAuthConnectionNotFound = errors.New("OAuth connection not found")
	ErrTooManyLoginAttempts    = errors.New("too many login attempts")

	// Email validation errors
	ErrEmailRequired = errors.New("email is required")
	ErrInvalidEmail  = errors.New("invalid email format")

	// Password validation errors
	ErrPasswordTooShort     = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong      = errors.New("password must be at most 72 characters")
	ErrPasswordWeak         = errors.New("password must contain uppercase, lowercase, and digit")
	ErrPasswordHashFailed   = errors.New("failed to hash password")
	ErrPasswordHashRequired = errors.New("password hash is required")
	ErrFirstNameRequired    = errors.New("first name is required")
	ErrFirstNameTooLong     = errors.New("first name must be at most 100 characters")
	ErrLastNameRequired     = errors.New("last name is required")
	ErrLastNameTooLong      = errors.New("last name must be at most 100 characters")
)
