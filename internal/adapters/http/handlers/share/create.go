package share

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// CreateShareRequest represents create share request
type CreateShareRequest struct {
	WorkflowID string     `json:"workflowId"`
	Email      string     `json:"email,omitempty"`
	UserID     string     `json:"userId,omitempty"`
	Permission string     `json:"permission"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
}

// CreateHandler handles create share request
type CreateHandler struct {
	repo ShareRepository
}

// NewCreateHandler creates a new handler
func NewCreateHandler(repo ShareRepository) *CreateHandler {
	return &CreateHandler{repo: repo}
}

// Handle handles the create share request
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.WorkflowID == "" {
		common.BadRequest(w, "workflow ID is required")
		return
	}

	if req.Email == "" && req.UserID == "" {
		common.BadRequest(w, "email or user ID is required")
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

	common.Created(w, share)
}
