package webhook

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	webhookCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/webhook"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type TriggerHandler struct {
	handler *webhookCmd.TriggerWebhookHandler
}

func NewTriggerHandler(handler *webhookCmd.TriggerWebhookHandler) *TriggerHandler {
	return &TriggerHandler{handler: handler}
}

func (h *TriggerHandler) Handle(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	if path == "" {
		common.BadRequest(w, "webhook path is required")
		return
	}
	// Ensure path starts with /
	if path[0] != '/' {
		path = "/" + path
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	var body types.JSON
	if r.Body != nil {
		data, err := io.ReadAll(r.Body)
		if err == nil && len(data) > 0 {
			body = types.JSON{"raw": string(data)}
		}
	}

	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	result, err := h.handler.Handle(r.Context(), webhookCmd.TriggerWebhookCommand{
		Path:      path,
		Method:    r.Method,
		Headers:   headers,
		Body:      body,
		IPAddress: ipAddress,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"success":     true,
		"endpoint_id": result.EndpointID.String(),
		"workflow_id": result.WorkflowID.String(),
	})
}
