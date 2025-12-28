package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/rs/zerolog/log"
)

// UserService handles user management operations.
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new UserService with required dependencies.
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	if userRepo == nil {
		panic("user service: userRepo is required")
	}
	return &UserService{userRepo: userRepo}
}

// GetByID returns a user by their ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetByEmail returns a user by their email address.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// UpdateUserInput holds the fields that can be updated on a user.
type UpdateUserInput struct {
	FirstName *string
	LastName  *string
	Username  *string
	AvatarURL *string
}

// Update updates a user's profile information.
func (s *UserService) Update(ctx context.Context, userID uuid.UUID, input UpdateUserInput) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
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
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	log.Info().Str("user_id", userID.String()).Msg("User profile updated")

	return user, nil
}

// Delete deletes a user account.
func (s *UserService) Delete(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	log.Info().Str("user_id", userID.String()).Msg("User deleted")
	return nil
}

// VerifyEmail marks a user's email as verified.
func (s *UserService) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.VerifyEmail(ctx, userID); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}
	log.Info().Str("user_id", userID.String()).Msg("User email verified")
	return nil
}

// List returns a paginated list of all users.
func (s *UserService) List(ctx context.Context, opts *repositories.ListOptions) ([]models.User, int64, error) {
	users, total, err := s.userRepo.FindAll(ctx, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	return users, total, nil
}
