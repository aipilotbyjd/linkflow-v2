package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

var (
	ErrShareNotFound   = errors.New("share not found")
	ErrShareExpired    = errors.New("share link has expired")
	ErrShareMaxViews   = errors.New("share link has reached maximum views")
	ErrInvalidPassword = errors.New("invalid password")
)

type ExecutionShareService struct {
	repo *repositories.BaseRepository[models.ExecutionShare]
}

func NewExecutionShareService(repo *repositories.BaseRepository[models.ExecutionShare]) *ExecutionShareService {
	return &ExecutionShareService{repo: repo}
}

type CreateShareInput struct {
	ExecutionID   uuid.UUID
	WorkspaceID   uuid.UUID
	CreatedBy     uuid.UUID
	ExpiresAt     *time.Time
	Password      *string
	MaxViews      *int
	AllowDownload bool
	IncludeLogs   bool
	IncludeData   bool
}

func (s *ExecutionShareService) Create(ctx context.Context, input CreateShareInput) (*models.ExecutionShare, error) {
	token := generateShareToken(32)

	share := &models.ExecutionShare{
		ExecutionID:   input.ExecutionID,
		WorkspaceID:   input.WorkspaceID,
		CreatedBy:     input.CreatedBy,
		Token:         token,
		ExpiresAt:     input.ExpiresAt,
		Password:      input.Password,
		MaxViews:      input.MaxViews,
		AllowDownload: input.AllowDownload,
		IncludeLogs:   input.IncludeLogs,
		IncludeData:   input.IncludeData,
	}

	if err := s.repo.Create(ctx, share); err != nil {
		return nil, err
	}
	return share, nil
}

func (s *ExecutionShareService) GetByToken(ctx context.Context, token string) (*models.ExecutionShare, error) {
	var share models.ExecutionShare
	err := s.repo.DB().WithContext(ctx).Where("token = ?", token).First(&share).Error
	if err != nil {
		return nil, ErrShareNotFound
	}
	return &share, nil
}

func (s *ExecutionShareService) ValidateAccess(ctx context.Context, share *models.ExecutionShare, password *string) error {
	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		return ErrShareExpired
	}
	if share.MaxViews != nil && share.ViewCount >= *share.MaxViews {
		return ErrShareMaxViews
	}
	if share.Password != nil && (password == nil || *password != *share.Password) {
		return ErrInvalidPassword
	}
	return nil
}

func (s *ExecutionShareService) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Model(&models.ExecutionShare{}).
		Where("id = ?", id).
		UpdateColumn("view_count", s.repo.DB().Raw("view_count + 1")).Error
}

func (s *ExecutionShareService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Delete(&models.ExecutionShare{}, "id = ?", id).Error
}

func generateShareToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}
