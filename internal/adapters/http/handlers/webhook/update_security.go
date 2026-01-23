package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
)

type UpdateSecurityRequest struct {
	RequireSignature   *bool    `json:"require_signature,omitempty"`
	SignatureHeader    string   `json:"signature_header,omitempty"`
	AllowedIPs         []string `json:"allowed_ips,omitempty"`
	RequireTimestamp   *bool    `json:"require_timestamp,omitempty"`
	TimestampHeader    string   `json:"timestamp_header,omitempty"`
	TimestampMaxAgeSec *int     `json:"timestamp_max_age_sec,omitempty"`
	RequireNonce       *bool    `json:"require_nonce,omitempty"`
	NonceHeader        string   `json:"nonce_header,omitempty"`
}

type UpdateSecurityHandler struct {
	repo webhook.Repository
}

func NewUpdateSecurityHandler(repo webhook.Repository) *UpdateSecurityHandler {
	return &UpdateSecurityHandler{repo: repo}
}

func (h *UpdateSecurityHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())
	if workspaceID == uuid.Nil {
		common.Unauthorized(w, "workspace context required")
		return
	}

	endpointID, err := uuid.Parse(chi.URLParam(r, "endpointId"))
	if err != nil {
		common.BadRequest(w, "invalid endpoint ID")
		return
	}

	var req UpdateSecurityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	// Get endpoint
	endpoint, err := h.repo.FindByID(r.Context(), endpointID)
	if err != nil {
		common.NotFound(w, "webhook endpoint")
		return
	}

	// Verify workspace ownership
	if endpoint.WorkspaceID != workspaceID {
		common.Forbidden(w, "access denied to this webhook endpoint")
		return
	}

	// Update security settings
	if req.RequireSignature != nil {
		endpoint.RequireSignature = *req.RequireSignature
	}
	if req.SignatureHeader != "" {
		endpoint.SignatureHeader = req.SignatureHeader
	}
	if req.AllowedIPs != nil {
		endpoint.SetAllowedIPs(req.AllowedIPs)
	}
	if req.RequireTimestamp != nil {
		endpoint.RequireTimestamp = *req.RequireTimestamp
	}
	if req.TimestampHeader != "" {
		endpoint.TimestampHeader = req.TimestampHeader
	}
	if req.TimestampMaxAgeSec != nil && *req.TimestampMaxAgeSec > 0 {
		endpoint.TimestampMaxAgeSec = *req.TimestampMaxAgeSec
	}
	if req.RequireNonce != nil {
		endpoint.RequireNonce = *req.RequireNonce
	}
	if req.NonceHeader != "" {
		endpoint.NonceHeader = req.NonceHeader
	}

	// Save updates
	if err := h.repo.Update(r.Context(), endpoint); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"id":                    endpoint.ID,
		"require_signature":     endpoint.RequireSignature,
		"signature_header":      endpoint.SignatureHeader,
		"allowed_ips":           endpoint.GetAllowedIPsList(),
		"require_timestamp":     endpoint.RequireTimestamp,
		"timestamp_header":      endpoint.TimestampHeader,
		"timestamp_max_age_sec": endpoint.TimestampMaxAgeSec,
		"require_nonce":         endpoint.RequireNonce,
		"nonce_header":          endpoint.NonceHeader,
		"is_secured":            endpoint.IsSecured(),
	})
}
