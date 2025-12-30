package validation

import (
	"errors"
	"strings"
)

var (
	ErrCredentialNameRequired = errors.New("credential name is required")
	ErrCredentialNameTooLong  = errors.New("credential name must be less than 100 characters")
	ErrCredentialTypeRequired = errors.New("credential type is required")
	ErrCredentialTypeInvalid  = errors.New("invalid credential type")
	ErrCredentialDataRequired = errors.New("credential data is required")
)

// ValidCredentialTypes defines allowed credential types
var ValidCredentialTypes = map[string]bool{
	"oauth2":       true,
	"api_key":      true,
	"basic_auth":   true,
	"bearer_token": true,
	"custom":       true,
}

// CredentialInput represents input for credential validation
type CredentialInput struct {
	Name string
	Type string
	Data map[string]interface{}
}

// ValidateCredential validates credential input
func ValidateCredential(input CredentialInput) error {
	// Name validation
	if strings.TrimSpace(input.Name) == "" {
		return ErrCredentialNameRequired
	}
	if len(input.Name) > 100 {
		return ErrCredentialNameTooLong
	}

	// Type validation
	if strings.TrimSpace(input.Type) == "" {
		return ErrCredentialTypeRequired
	}
	if !ValidCredentialTypes[input.Type] {
		return ErrCredentialTypeInvalid
	}

	// Data validation
	if input.Data == nil || len(input.Data) == 0 {
		return ErrCredentialDataRequired
	}

	return nil
}
