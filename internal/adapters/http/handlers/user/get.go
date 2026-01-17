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

	common.Success(w, ToUserResponse(user))
}
