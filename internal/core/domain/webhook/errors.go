package webhook

import "errors"

// Domain-specific errors for webhook aggregate
var (
	ErrEndpointNotFound  = errors.New("webhook endpoint not found")
	ErrPathAlreadyExists = errors.New("webhook path already exists")
	ErrEndpointNotActive = errors.New("webhook endpoint is not active")
	ErrEndpointInactive  = errors.New("webhook endpoint is inactive")
	ErrInvalidSignature  = errors.New("invalid webhook signature")
	ErrMethodNotAllowed  = errors.New("HTTP method not allowed")
	ErrPayloadTooLarge   = errors.New("webhook payload too large")
	ErrInvalidPayload    = errors.New("invalid webhook payload")
	ErrWorkflowNotActive = errors.New("workflow is not active")
	ErrRateLimitExceeded = errors.New("webhook rate limit exceeded")
	ErrSecretRequired    = errors.New("webhook secret is required")

	// Validation errors
	ErrInvalidWorkflowID  = errors.New("valid workflow ID is required")
	ErrInvalidWorkspaceID = errors.New("valid workspace ID is required")
	ErrNodeIDRequired     = errors.New("node ID is required")
	ErrNodeIDTooLong      = errors.New("node ID must be at most 100 characters")
	ErrPathRequired       = errors.New("webhook path is required")
	ErrPathTooLong        = errors.New("webhook path must be at most 255 characters")
	ErrInvalidPath        = errors.New("webhook path contains invalid characters or path traversal")
	ErrInvalidMethod      = errors.New("invalid HTTP method")
)
