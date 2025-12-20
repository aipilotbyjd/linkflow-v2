package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

type UpdateUserInput struct {
	FirstName *string
	LastName  *string
	Username  *string
	AvatarURL *string
}

func (s *UserService) Update(ctx context.Context, userID uuid.UUID, input UpdateUserInput) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.Username != nil {
		user.Username = input.Username
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.Delete(ctx, userID)
}

func (s *UserService) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.VerifyEmail(ctx, userID)
}

func (s *UserService) List(ctx context.Context, opts *repositories.ListOptions) ([]models.User, int64, error) {
	return s.userRepo.FindAll(ctx, opts)
}
