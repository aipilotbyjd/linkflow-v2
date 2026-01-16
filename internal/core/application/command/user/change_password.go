package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
)

var (
	ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")
)

type ChangePasswordCommand struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
}

type ChangePasswordHandler struct {
	userRepo user.Repository
	hasher   *crypto.Hasher
}

func NewChangePasswordHandler(userRepo user.Repository, hasher *crypto.Hasher) *ChangePasswordHandler {
	return &ChangePasswordHandler{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (h *ChangePasswordHandler) Handle(ctx context.Context, cmd ChangePasswordCommand) error {
	u, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	// Verify current password
	if !h.hasher.Verify(cmd.CurrentPassword, u.PasswordHash) {
		return ErrCurrentPasswordIncorrect
	}

	// Hash and set new password
	newHash, err := h.hasher.Hash(cmd.NewPassword)
	if err != nil {
		return err
	}

	u.UpdatePassword(newHash)

	return h.userRepo.Update(ctx, u)
}
