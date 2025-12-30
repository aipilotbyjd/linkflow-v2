package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

var ErrEnvVarNotFound = errors.New("environment variable not found")

type EnvironmentVariableService struct {
	repo      *repositories.BaseRepository[models.EnvironmentVariable]
	encryptor Encryptor
}

type Encryptor interface {
	Encrypt(data string) (string, error)
	Decrypt(data string) (string, error)
}

func NewEnvironmentVariableService(repo *repositories.BaseRepository[models.EnvironmentVariable], encryptor Encryptor) *EnvironmentVariableService {
	return &EnvironmentVariableService{repo: repo, encryptor: encryptor}
}

type CreateEnvVarInput struct {
	WorkspaceID uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Value       string
	IsSecret    bool
	Environment *string
	Description *string
}

func (s *EnvironmentVariableService) Create(ctx context.Context, input CreateEnvVarInput) (*models.EnvironmentVariable, error) {
	encrypted, err := s.encryptor.Encrypt(input.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt value: %w", err)
	}

	envVar := &models.EnvironmentVariable{
		WorkspaceID: input.WorkspaceID,
		CreatedBy:   input.CreatedBy,
		Name:        input.Name,
		Value:       encrypted,
		IsSecret:    input.IsSecret,
		Environment: input.Environment,
		Description: input.Description,
	}

	if err := s.repo.Create(ctx, envVar); err != nil {
		return nil, err
	}
	return envVar, nil
}

func (s *EnvironmentVariableService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, environment *string) ([]models.EnvironmentVariable, error) {
	var vars []models.EnvironmentVariable
	query := s.repo.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if environment != nil {
		query = query.Where("environment = ? OR environment IS NULL", *environment)
	}
	err := query.Order("name").Find(&vars).Error
	return vars, err
}

func (s *EnvironmentVariableService) GetDecrypted(ctx context.Context, workspaceID uuid.UUID, environment *string) (map[string]string, error) {
	vars, err := s.GetByWorkspace(ctx, workspaceID, environment)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, v := range vars {
		decrypted, err := s.encryptor.Decrypt(v.Value)
		if err != nil {
			continue
		}
		result[v.Name] = decrypted
	}
	return result, nil
}

func (s *EnvironmentVariableService) Update(ctx context.Context, id uuid.UUID, value string) error {
	encrypted, err := s.encryptor.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}
	return s.repo.DB().WithContext(ctx).Model(&models.EnvironmentVariable{}).
		Where("id = ?", id).
		Update("value", encrypted).Error
}

func (s *EnvironmentVariableService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Delete(&models.EnvironmentVariable{}, "id = ?", id).Error
}
