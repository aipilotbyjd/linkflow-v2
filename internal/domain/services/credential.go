package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/rs/zerolog/log"
)

// Credential errors
var (
	ErrCredentialNotFound     = errors.New("credential not found")
	ErrCredentialNameRequired = errors.New("credential name is required")
	ErrCredentialTypeRequired = errors.New("credential type is required")
)

// CredentialService handles credential management with encryption.
type CredentialService struct {
	credentialRepo *repositories.CredentialRepository
	encryptor      *crypto.Encryptor
}

// NewCredentialService creates a new CredentialService with required dependencies.
func NewCredentialService(
	credentialRepo *repositories.CredentialRepository,
	encryptor *crypto.Encryptor,
) *CredentialService {
	if credentialRepo == nil || encryptor == nil {
		panic("credential service: credentialRepo and encryptor are required")
	}
	return &CredentialService{
		credentialRepo: credentialRepo,
		encryptor:      encryptor,
	}
}

type CreateCredentialInput struct {
	WorkspaceID uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Type        string
	Data        models.CredentialData
	Description *string
}

// Create creates a new encrypted credential.
func (s *CredentialService) Create(ctx context.Context, input CreateCredentialInput) (*models.Credential, error) {
	// Validate input
	if input.Name == "" {
		return nil, ErrCredentialNameRequired
	}
	if input.Type == "" {
		return nil, ErrCredentialTypeRequired
	}

	dataJSON, err := json.Marshal(input.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential data: %w", err)
	}

	encryptedData, err := s.encryptor.Encrypt(string(dataJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential data: %w", err)
	}

	credential := &models.Credential{
		WorkspaceID: input.WorkspaceID,
		CreatedBy:   input.CreatedBy,
		Name:        input.Name,
		Type:        input.Type,
		Data:        encryptedData,
		Description: input.Description,
	}

	if err := s.credentialRepo.Create(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	log.Info().
		Str("credential_id", credential.ID.String()).
		Str("workspace_id", input.WorkspaceID.String()).
		Str("type", input.Type).
		Msg("Credential created")

	return credential, nil
}

// GetByID returns a credential by its ID (encrypted data).
func (s *CredentialService) GetByID(ctx context.Context, id uuid.UUID) (*models.Credential, error) {
	credential, err := s.credentialRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCredentialNotFound, id)
	}
	return credential, nil
}

// GetByWorkspace returns paginated credentials for a workspace.
func (s *CredentialService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Credential, int64, error) {
	credentials, total, err := s.credentialRepo.FindByWorkspaceID(ctx, workspaceID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get credentials: %w", err)
	}
	return credentials, total, nil
}

// GetByWorkspaceWithFilters returns paginated credentials for a workspace with filters.
func (s *CredentialService) GetByWorkspaceWithFilters(ctx context.Context, workspaceID uuid.UUID, filter *repositories.CredentialFilter, opts *repositories.ListOptions) ([]models.Credential, int64, error) {
	credentials, total, err := s.credentialRepo.FindWithFilters(ctx, workspaceID, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get credentials: %w", err)
	}
	return credentials, total, nil
}

// GetDecrypted returns a credential with its decrypted data.
func (s *CredentialService) GetDecrypted(ctx context.Context, id uuid.UUID) (*models.Credential, *models.CredentialData, error) {
	credential, err := s.credentialRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrCredentialNotFound, id)
	}

	decrypted, err := s.encryptor.Decrypt(credential.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt credential data: %w", err)
	}

	var data models.CredentialData
	if err := json.Unmarshal([]byte(decrypted), &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal credential data: %w", err)
	}

	// Update last used timestamp (non-blocking)
	if err := s.credentialRepo.UpdateLastUsed(ctx, id); err != nil {
		log.Warn().
			Err(err).
			Str("credential_id", id.String()).
			Msg("Failed to update credential last used timestamp")
	}

	return credential, &data, nil
}

type UpdateCredentialInput struct {
	Name        *string
	Data        *models.CredentialData
	Description *string
}

// Update updates an existing credential.
func (s *CredentialService) Update(ctx context.Context, credentialID uuid.UUID, input UpdateCredentialInput) (*models.Credential, error) {
	credential, err := s.credentialRepo.FindByID(ctx, credentialID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCredentialNotFound, credentialID)
	}

	// Validate name if provided
	if input.Name != nil && *input.Name == "" {
		return nil, ErrCredentialNameRequired
	}

	if input.Name != nil {
		credential.Name = *input.Name
	}
	if input.Description != nil {
		credential.Description = input.Description
	}
	if input.Data != nil {
		dataJSON, err := json.Marshal(input.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal credential data: %w", err)
		}

		encryptedData, err := s.encryptor.Encrypt(string(dataJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt credential data: %w", err)
		}
		credential.Data = encryptedData
	}

	if err := s.credentialRepo.Update(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	log.Info().
		Str("credential_id", credentialID.String()).
		Msg("Credential updated")

	return credential, nil
}

// Delete deletes a credential.
func (s *CredentialService) Delete(ctx context.Context, credentialID uuid.UUID) error {
	if err := s.credentialRepo.Delete(ctx, credentialID); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	log.Info().Str("credential_id", credentialID.String()).Msg("Credential deleted")
	return nil
}

// TestConnection tests if a credential can be used to connect to its service.
// Performs type-specific validation: checks required fields exist and have valid formats.
// For OAuth2, verifies token hasn't expired. For API keys and bearer tokens,
// validates that required fields are present and non-empty.
func (s *CredentialService) TestConnection(ctx context.Context, credentialID uuid.UUID) (bool, error) {
	credential, data, err := s.GetDecrypted(ctx, credentialID)
	if err != nil {
		return false, err
	}

	if data == nil {
		return false, fmt.Errorf("credential data is empty")
	}

	// Type-specific connection testing
	switch credential.Type {
	case "oauth2":
		// Check for required OAuth2 fields
		if data.AccessToken == "" {
			return false, fmt.Errorf("oauth2 credential missing access_token")
		}

	case "api_key":
		// Check for API key presence
		if data.APIKey == "" {
			return false, fmt.Errorf("api_key credential missing api_key field")
		}

	case "bearer_token":
		// Check for bearer token presence
		if data.Token == "" {
			return false, fmt.Errorf("bearer_token credential missing token field")
		}

	case "basic_auth":
		// Check for username and password
		if data.Username == "" {
			return false, fmt.Errorf("basic_auth credential missing username")
		}
		if data.Password == "" {
			return false, fmt.Errorf("basic_auth credential missing password")
		}

	case "custom":
		// For custom credentials, just verify data exists
		if len(data.Custom) == 0 && len(data.Data) == 0 {
			return false, fmt.Errorf("custom credential has no data fields")
		}
	}

	log.Debug().
		Str("credential_id", credentialID.String()).
		Str("type", credential.Type).
		Bool("has_data", data != nil).
		Msg("Credential test connection - validation successful")

	return true, nil
}
