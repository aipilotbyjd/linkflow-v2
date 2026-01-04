package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/mappers"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type UserHandler struct {
	userSvc *services.UserService
}

// NewUserHandler creates a new UserHandler for user profile management.
func NewUserHandler(userSvc *services.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetCurrentUser returns the authenticated user's profile.
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	basePath := "/api/v1/users/me"
	actions := []dto.Action{
		{Name: "update", Method: "PUT", Href: basePath, Label: "Update Profile"},
		{Name: "change_password", Method: "POST", Href: basePath + "/password", Label: "Change Password"},
	}
	if !user.MFAEnabled {
		actions = append(actions, dto.Action{Name: "setup_mfa", Method: "POST", Href: "/api/v1/auth/mfa/setup", Label: "Setup MFA"})
	} else {
		actions = append(actions, dto.Action{Name: "disable_mfa", Method: "DELETE", Href: "/api/v1/auth/mfa", Label: "Disable MFA"})
	}

	response := struct {
		dto.UserResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}{
		UserResponse: mappers.UserToResponse(user),
		Actions:      actions,
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
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
		FirstName:               req.FirstName,
		LastName:                req.LastName,
		Username:                req.Username,
		AvatarURL:               req.AvatarURL,
		Phone:                   req.Phone,
		Bio:                     req.Bio,
		JobTitle:                req.JobTitle,
		Company:                 req.Company,
		Timezone:                req.Timezone,
		Language:                req.Language,
		DateFormat:              req.DateFormat,
		TimeFormat:              req.TimeFormat,
		Theme:                   req.Theme,
		NotificationPreferences: req.NotificationPreferences,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	dto.NewResponse(mappers.UserToResponse(user)).
		WithLinks(&dto.Links{Self: "/api/v1/users/me"}).
		Send(w)
}
