package credential

import "errors"

// Domain-specific errors for credential aggregate
var (
	ErrCredentialNotFound   = errors.New("credential not found")
	ErrCredentialNameExists = errors.New("credential name already exists")
	ErrCredentialExpired    = errors.New("credential has expired")
	ErrCredentialInvalid    = errors.New("invalid credential data")
	ErrAccessDenied         = errors.New("access denied to credential")
	ErrCannotShare          = errors.New("cannot share credential")
	ErrShareNotFound        = errors.New("credential share not found")
	ErrAlreadyShared        = errors.New("credential already shared with user")
	ErrCannotShareToSelf    = errors.New("cannot share credential to yourself")
	ErrEncryptionFailed     = errors.New("encryption failed")
	ErrDecryptionFailed     = errors.New("decryption failed")
	ErrInvalidType          = errors.New("invalid credential type")
	ErrOAuthRefreshFailed   = errors.New("OAuth token refresh failed")
	ErrProviderNotSupported = errors.New("OAuth provider not supported")

	// Validation errors
	ErrCredentialNameRequired = errors.New("credential name is required")
	ErrCredentialNameTooLong  = errors.New("credential name must be at most 100 characters")
	ErrInvalidWorkspaceID     = errors.New("valid workspace ID is required")
	ErrInvalidCreatedBy       = errors.New("valid creator ID is required")
	ErrDataRequired           = errors.New("credential data is required")
	ErrInvalidSharingScope    = errors.New("invalid sharing scope")
)
