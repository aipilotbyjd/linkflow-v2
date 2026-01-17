package user

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	userQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/user"
)

type GetCurrentUserHandler struct {
	handler *userQuery.GetUserHandler
}

func NewGetCurrentUserHandler(handler *userQuery.GetUserHandler) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{handler: handler}
}

type UserResponse struct {
	ID                      string  `json:"id"`
	Email                   string  `json:"email"`
	Username                *string `json:"username,omitempty"`
	FirstName               string  `json:"first_name"`
	LastName                string  `json:"last_name"`
	AvatarURL               *string `json:"avatar_url,omitempty"`
	Phone                   *string `json:"phone,omitempty"`
	Bio                     *string `json:"bio,omitempty"`
	JobTitle                *string `json:"job_title,omitempty"`
	Company                 *string `json:"company,omitempty"`
	Timezone                string  `json:"timezone"`
	Language                string  `json:"language"`
	DateFormat              string  `json:"date_format"`
	TimeFormat              string  `json:"time_format"`
	Theme                   string  `json:"theme"`
	EmailVerified           bool    `json:"email_verified"`
	MFAEnabled              bool    `json:"mfa_enabled"`
	CreatedAt               string  `json:"created_at"`
}

func (h *GetCurrentUserHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	user, err := h.handler.Handle(r.Context(), userQuery.GetUserQuery{
		UserID: userClaims.UserID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	response := UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		AvatarURL:     user.AvatarURL,
		Phone:         user.Phone,
		Bio:           user.Bio,
		JobTitle:      user.JobTitle,
		Company:       user.Company,
		Timezone:      user.Timezone,
		Language:      user.Language,
		DateFormat:    user.DateFormat,
		TimeFormat:    user.TimeFormat,
		Theme:         user.Theme,
		EmailVerified: user.EmailVerified,
		MFAEnabled:    user.MFAEnabled,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	common.Success(w, response)
}
