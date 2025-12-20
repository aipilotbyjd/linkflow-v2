package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
)

type CredentialService struct {
	credentialRepo *repositories.CredentialRepository
	encryptor      *crypto.Encryptor
}

func NewCredentialService(
	credentialRepo *repositories.CredentialRepository,
	encryptor *crypto.Encryptor,
) *CredentialService {
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

func (s *CredentialService) Create(ctx context.Context, input CreateCredentialInput) (*models.Credential, error) {
	dataJSON, err := json.Marshal(input.Data)
	if err != nil {
		return nil, err
	}

	encryptedData, err := s.encryptor.Encrypt(string(dataJSON))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return credential, nil
}

func (s *CredentialService) GetByID(ctx context.Context, id uuid.UUID) (*models.Credential, error) {
	return s.credentialRepo.FindByID(ctx, id)
}

func (s *CredentialService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Credential, int64, error) {
	return s.credentialRepo.FindByWorkspaceID(ctx, workspaceID, opts)
}

func (s *CredentialService) GetDecrypted(ctx context.Context, id uuid.UUID) (*models.Credential, *models.CredentialData, error) {
	credential, err := s.credentialRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	decrypted, err := s.encryptor.Decrypt(credential.Data)
	if err != nil {
		return nil, nil, err
	}

	var data models.CredentialData
	if err := json.Unmarshal([]byte(decrypted), &data); err != nil {
		return nil, nil, err
	}

	s.credentialRepo.UpdateLastUsed(ctx, id)

	return credential, &data, nil
}

type UpdateCredentialInput struct {
	Name        *string
	Data        *models.CredentialData
	Description *string
}

func (s *CredentialService) Update(ctx context.Context, credentialID uuid.UUID, input UpdateCredentialInput) (*models.Credential, error) {
	credential, err := s.credentialRepo.FindByID(ctx, credentialID)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		encryptedData, err := s.encryptor.Encrypt(string(dataJSON))
		if err != nil {
			return nil, err
		}
		credential.Data = encryptedData
	}

	if err := s.credentialRepo.Update(ctx, credential); err != nil {
		return nil, err
	}

	return credential, nil
}

func (s *CredentialService) Delete(ctx context.Context, credentialID uuid.UUID) error {
	return s.credentialRepo.Delete(ctx, credentialID)
}

func (s *CredentialService) TestConnection(ctx context.Context, credentialID uuid.UUID) (bool, error) {
	_, data, err := s.GetDecrypted(ctx, credentialID)
	if err != nil {
		return false, err
	}

	_ = data
	return true, nil
}
