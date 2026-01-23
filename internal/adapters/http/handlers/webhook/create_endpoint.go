package webhook

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	workflowDomain "github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type CreateEndpointRequest struct {
	WorkflowID         string   `json:"workflow_id" validate:"required,uuid"`
	NodeID             string   `json:"node_id" validate:"required,min=1,max=100"`
	Path               string   `json:"path" validate:"required,min=1,max=255,webhook_path"`
	Method             string   `json:"method" validate:"required,http_method"`
	AuthenticationType string   `json:"authentication_type,omitempty" validate:"omitempty,oneof=none signature basic bearer api_key"`
	AllowedIPs         []string `json:"allowed_ips,omitempty" validate:"omitempty,dive,ip|cidr"`
	RateLimit          int      `json:"rate_limit,omitempty" validate:"omitempty,min=1,max=10000"`
}

type CreateEndpointHandler struct {
	webhookRepo  webhook.Repository
	workflowRepo workflowDomain.Repository
	baseURL      string
}

func NewCreateEndpointHandler(webhookRepo webhook.Repository, workflowRepo workflowDomain.Repository, baseURL string) *CreateEndpointHandler {
	return &CreateEndpointHandler{
		webhookRepo:  webhookRepo,
		workflowRepo: workflowRepo,
		baseURL:      baseURL,
	}
}

func (h *CreateEndpointHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}
	workspaceID := wsCtx.WorkspaceID

	var req CreateEndpointRequest
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

	workflowID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		common.BadRequest(w, "Invalid workflow ID")
		return
	}

	// Validate workflow exists and belongs to workspace
	wf, err := h.workflowRepo.FindByID(r.Context(), workflowID)
	if err != nil {
		common.NotFound(w, "Workflow not found")
		return
	}

	if wf.WorkspaceID != workspaceID {
		common.Forbidden(w, "Workflow does not belong to this workspace")
		return
	}

	// Generate webhook secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		common.HandleError(w, err)
		return
	}
	secret := base64.URLEncoding.EncodeToString(secretBytes)

	// Create endpoint
	endpoint, err := webhook.NewEndpoint(workflowID, workspaceID, req.NodeID, req.Path)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	endpoint.WithMethod(req.Method)
	endpoint.WithSecret(secret)

	if err := h.webhookRepo.Create(r.Context(), endpoint); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, map[string]interface{}{
		"id":     endpoint.ID.String(),
		"path":   endpoint.Path,
		"method": endpoint.Method,
		"secret": secret,
		"url":    endpoint.GetURL(h.baseURL),
	})
}
