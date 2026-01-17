package share

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

type WorkflowShare struct {
	ID           string    `json:"id"`
	WorkflowID   string    `json:"workflowId"`
	WorkflowName string    `json:"workflowName"`
	SharedBy     string    `json:"sharedBy"`
	SharedByName string    `json:"sharedByName"`
	SharedWith   string    `json:"sharedWith"`
	SharedWithName string  `json:"sharedWithName,omitempty"`
	SharedWithEmail string `json:"sharedWithEmail,omitempty"`
	Permission   string    `json:"permission"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	AcceptedAt   *time.Time `json:"acceptedAt,omitempty"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
}

type ShareRepository interface {
	Create(share *WorkflowShare) error
	GetByID(id string) (*WorkflowShare, error)
	GetSharedByMe(userID string) ([]WorkflowShare, error)
	GetSharedWithMe(userID string) ([]WorkflowShare, error)
	GetPending(userID string) ([]WorkflowShare, error)
	Accept(id string) error
	Update(id string, permission string) error
	Delete(id string) error
}

type Handler struct {
	repo ShareRepository
}

func NewHandler(repo ShareRepository) *Handler {
	return &Handler{repo: repo}
}

type CreateShareRequest struct {
	WorkflowID  string `json:"workflowId"`
	Email       string `json:"email,omitempty"`
	UserID      string `json:"userId,omitempty"`
	Permission  string `json:"permission"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.WorkflowID == "" {
		common.Error(w, http.StatusBadRequest, "MISSING_WORKFLOW", "Workflow ID is required")
		return
	}

	if req.Email == "" && req.UserID == "" {
		common.Error(w, http.StatusBadRequest, "MISSING_RECIPIENT", "Email or user ID is required")
		return
	}

	if req.Permission == "" {
		req.Permission = "view"
	}

	share := WorkflowShare{
		ID:              uuid.New().String(),
		WorkflowID:      req.WorkflowID,
		WorkflowName:    "Sample Workflow",
		SharedBy:        userID.String(),
		SharedByName:    "Current User",
		SharedWith:      req.UserID,
		SharedWithEmail: req.Email,
		Permission:      req.Permission,
		Status:          "pending",
		CreatedAt:       time.Now(),
		ExpiresAt:       req.ExpiresAt,
	}

	_ = workspaceID

	common.JSON(w, http.StatusCreated, share)
}

func (h *Handler) SharedByMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares := []WorkflowShare{
		{
			ID:             uuid.New().String(),
			WorkflowID:     uuid.New().String(),
			WorkflowName:   "My Workflow",
			SharedBy:       userID.String(),
			SharedByName:   "You",
			SharedWith:     uuid.New().String(),
			SharedWithName: "John Doe",
			SharedWithEmail: "john@example.com",
			Permission:     "view",
			Status:         "accepted",
			CreatedAt:      time.Now().AddDate(0, 0, -7),
			AcceptedAt:     func() *time.Time { t := time.Now().AddDate(0, 0, -6); return &t }(),
		},
	}

	common.Success(w, map[string]interface{}{
		"shares": shares,
		"total":  len(shares),
	})
}

func (h *Handler) SharedWithMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares := []WorkflowShare{
		{
			ID:           uuid.New().String(),
			WorkflowID:   uuid.New().String(),
			WorkflowName: "Shared Workflow",
			SharedBy:     uuid.New().String(),
			SharedByName: "Jane Smith",
			SharedWith:   userID.String(),
			Permission:   "edit",
			Status:       "accepted",
			CreatedAt:    time.Now().AddDate(0, 0, -14),
			AcceptedAt:   func() *time.Time { t := time.Now().AddDate(0, 0, -13); return &t }(),
		},
	}

	common.Success(w, map[string]interface{}{
		"shares": shares,
		"total":  len(shares),
	})
}

func (h *Handler) Pending(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares := []WorkflowShare{
		{
			ID:           uuid.New().String(),
			WorkflowID:   uuid.New().String(),
			WorkflowName: "Pending Workflow",
			SharedBy:     uuid.New().String(),
			SharedByName: "Bob Wilson",
			SharedWith:   userID.String(),
			Permission:   "view",
			Status:       "pending",
			CreatedAt:    time.Now().AddDate(0, 0, -1),
		},
	}

	common.Success(w, map[string]interface{}{
		"shares": shares,
		"total":  len(shares),
	})
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	shareID := chi.URLParam(r, "shareId")

	share := WorkflowShare{
		ID:           shareID,
		WorkflowID:   uuid.New().String(),
		WorkflowName: "Accepted Workflow",
		Status:       "accepted",
		AcceptedAt:   func() *time.Time { t := time.Now(); return &t }(),
	}

	common.Success(w, share)
}

type UpdateShareRequest struct {
	Permission string `json:"permission"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	shareID := chi.URLParam(r, "shareId")

	var req UpdateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	share := WorkflowShare{
		ID:         shareID,
		Permission: req.Permission,
		Status:     "accepted",
	}

	common.Success(w, share)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	shareID := chi.URLParam(r, "shareId")
	_ = shareID

	w.WriteHeader(http.StatusNoContent)
}
