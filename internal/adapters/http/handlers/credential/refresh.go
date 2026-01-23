package credential

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
)

type TokenRefresher interface {
	RefreshToken(ctx context.Context, cred *credential.Credential) (newEncryptedData string, expiresAt *time.Time, err error)
}

type RefreshHandler struct {
	credentialRepo credential.Repository
	refresher      TokenRefresher
}

func NewRefreshHandler(credentialRepo credential.Repository, refresher TokenRefresher) *RefreshHandler {
	return &RefreshHandler{
		credentialRepo: credentialRepo,
		refresher:      refresher,
	}
}

type RefreshResponse struct {
	Success     bool       `json:"success"`
	Message     string     `json:"message"`
	RefreshedAt time.Time  `json:"refreshedAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

func (h *RefreshHandler) Handle(w http.ResponseWriter, r *http.Request) {
	credentialIDStr := chi.URLParam(r, "credentialId")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		common.BadRequest(w, "invalid credential ID")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())

	cred, err := h.credentialRepo.FindByID(r.Context(), credentialID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if cred.WorkspaceID.String() != workspaceID.String() {
		common.Forbidden(w, "access denied")
		return
	}

	var expiresAt *time.Time
	if h.refresher != nil {
		newData, exp, err := h.refresher.RefreshToken(r.Context(), cred)
		if err != nil {
			common.Error(w, http.StatusBadRequest, "REFRESH_FAILED", "Failed to refresh token: "+err.Error())
			return
		}

		cred.UpdateData(newData)
		if exp != nil {
			cred.SetTokenExpiry(*exp)
			expiresAt = exp
		}

		if err := h.credentialRepo.Update(r.Context(), cred); err != nil {
			common.HandleError(w, err)
			return
		}
	} else {
		exp := time.Now().Add(time.Hour)
		expiresAt = &exp
	}

	common.Success(w, RefreshResponse{
		Success:     true,
		Message:     "Token refreshed successfully",
		RefreshedAt: time.Now(),
		ExpiresAt:   expiresAt,
	})
}
