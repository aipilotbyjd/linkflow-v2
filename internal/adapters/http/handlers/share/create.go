package share

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/share"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

// CreateHandler handles create share request
type CreateHandler struct {
	shareRepo share.Repository
	userRepo  user.Repository
}

// NewCreateHandler creates a new handler
func NewCreateHandler(shareRepo share.Repository, userRepo user.Repository) *CreateHandler {
	return &CreateHandler{shareRepo: shareRepo, userRepo: userRepo}
}

// Handle handles the create share request
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	claims := middleware.GetUserFromContext(r.Context())

	var req CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		common.BadRequest(w, "Invalid resource ID")
		return
	}

	// Find shared with user
	sharedWithUser, err := h.userRepo.FindByEmail(r.Context(), req.SharedWithEmail)
	if err != nil {
		common.NotFound(w, "User with that email")
		return
	}

	newShare := share.NewShare(
		req.ResourceType,
		resourceID,
		req.ResourceType, // Resource name - could be fetched from respective repo
		userID,
		claims.Email,
		sharedWithUser.ID,
		sharedWithUser.Email,
		share.SharePermission(req.Permission),
	)
	newShare.Message = req.Message

	if err := h.shareRepo.Create(r.Context(), newShare); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToShareResponse(*newShare))
}
