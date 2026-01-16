package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

type UpdateProfileCommand struct {
	UserID    uuid.UUID
	FirstName *string
	LastName  *string
	AvatarURL *string
	Phone     *string
	Bio       *string
	JobTitle  *string
	Company   *string
}

type UpdateProfileHandler struct {
	userRepo user.Repository
}

func NewUpdateProfileHandler(userRepo user.Repository) *UpdateProfileHandler {
	return &UpdateProfileHandler{userRepo: userRepo}
}

func (h *UpdateProfileHandler) Handle(ctx context.Context, cmd UpdateProfileCommand) (*user.User, error) {
	u, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	firstName := u.FirstName
	lastName := u.LastName
	if cmd.FirstName != nil {
		firstName = *cmd.FirstName
	}
	if cmd.LastName != nil {
		lastName = *cmd.LastName
	}

	u.UpdateProfile(firstName, lastName, cmd.Phone, cmd.Bio, cmd.JobTitle, cmd.Company)

	if cmd.AvatarURL != nil {
		u.AvatarURL = cmd.AvatarURL
	}

	if err := h.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
