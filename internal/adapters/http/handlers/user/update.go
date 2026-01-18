package user

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type UpdateCurrentUserHandler struct {
	userRepo user.Repository
}

func NewUpdateCurrentUserHandler(userRepo user.Repository) *UpdateCurrentUserHandler {
	return &UpdateCurrentUserHandler{userRepo: userRepo}
}

func (h *UpdateCurrentUserHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	// Get existing user
	existingUser, err := h.userRepo.FindByID(r.Context(), userClaims.UserID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Update fields
	firstName := existingUser.FirstName
	if req.FirstName != nil {
		firstName = *req.FirstName
	}
	lastName := existingUser.LastName
	if req.LastName != nil {
		lastName = *req.LastName
	}

	phone := existingUser.Phone
	if req.Phone != nil {
		phone = req.Phone
	}
	bio := existingUser.Bio
	if req.Bio != nil {
		bio = req.Bio
	}
	jobTitle := existingUser.JobTitle
	if req.JobTitle != nil {
		jobTitle = req.JobTitle
	}
	company := existingUser.Company
	if req.Company != nil {
		company = req.Company
	}

	existingUser.UpdateProfile(firstName, lastName, phone, bio, jobTitle, company)

	if req.AvatarURL != nil {
		existingUser.AvatarURL = req.AvatarURL
	}

	// Update preferences
	timezone := ""
	if req.Timezone != nil {
		timezone = *req.Timezone
	}
	language := ""
	if req.Language != nil {
		language = *req.Language
	}
	dateFormat := ""
	if req.DateFormat != nil {
		dateFormat = *req.DateFormat
	}
	timeFormat := ""
	if req.TimeFormat != nil {
		timeFormat = *req.TimeFormat
	}
	theme := ""
	if req.Theme != nil {
		theme = *req.Theme
	}

	existingUser.UpdatePreferences(timezone, language, dateFormat, timeFormat, theme)

	// Save updates
	if err := h.userRepo.Update(r.Context(), existingUser); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToUserResponse(existingUser))
}
