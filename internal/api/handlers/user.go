package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type UserHandler struct {
	userSvc *services.UserService
}

func NewUserHandler(userSvc *services.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	dto.JSON(w, http.StatusOK, dto.UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		AvatarURL:     user.AvatarURL,
		EmailVerified: user.EmailVerified,
		MFAEnabled:    user.MFAEnabled,
		CreatedAt:     user.CreatedAt.Unix(),
	})
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	user, err := h.userSvc.Update(r.Context(), claims.UserID, services.UpdateUserInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	dto.JSON(w, http.StatusOK, dto.UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		AvatarURL:     user.AvatarURL,
		EmailVerified: user.EmailVerified,
		MFAEnabled:    user.MFAEnabled,
		CreatedAt:     user.CreatedAt.Unix(),
	})
}
