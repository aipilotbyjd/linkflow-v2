package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type CreateEndpointRequest struct {
	WorkflowID         string   `json:"workflow_id" validate:"required,uuid"`
	Path               string   `json:"path" validate:"required"`
	Method             string   `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE"`
	AuthenticationType string   `json:"authentication_type,omitempty"`
	AllowedIPs         []string `json:"allowed_ips,omitempty"`
	RateLimit          int      `json:"rate_limit,omitempty"`
}

type CreateEndpointHandler struct{}

func NewCreateEndpointHandler() *CreateEndpointHandler {
	return &CreateEndpointHandler{}
}

func (h *CreateEndpointHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	if workspaceID == "" {
		common.BadRequest(w, "Workspace ID is required")
		return
	}

	var req CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// TODO: Implement endpoint creation
	// 1. Validate workflow exists and belongs to workspace
	// 2. Generate webhook secret
	// 3. Create endpoint

	common.Created(w, map[string]interface{}{
		"id":     "endpoint-id",
		"path":   req.Path,
		"secret": "generated-secret",
		"url":    "https://api.linkflow.ai/webhooks/endpoint-id",
	})
}
