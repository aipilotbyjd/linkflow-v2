package handler

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

type WebhookAutoRegHandler struct {
	webhookRepo  webhook.Repository
	workflowRepo workflow.Repository
}

func NewWebhookAutoRegHandler(webhookRepo webhook.Repository, workflowRepo workflow.Repository) *WebhookAutoRegHandler {
	return &WebhookAutoRegHandler{
		webhookRepo:  webhookRepo,
		workflowRepo: workflowRepo,
	}
}

func (h *WebhookAutoRegHandler) Handle(ctx context.Context, event events.Event) error {
	fmt.Printf("[WebhookAutoReg] Received event: %s\n", event.EventName())
	switch e := event.(type) {
	case events.WorkflowActivated:
		return h.handleActivation(ctx, e)
	case events.WorkflowDeactivated:
		return h.handleDeactivation(ctx, e)
	default:
		fmt.Printf("[WebhookAutoReg] Unhandled event type: %T\n", event)
	}
	return nil
}

func (h *WebhookAutoRegHandler) handleActivation(ctx context.Context, e events.WorkflowActivated) error {
	wf, err := h.workflowRepo.FindByID(ctx, e.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to fetch workflow for webhook registration: %w", err)
	}

	nodes, err := wf.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to parse workflow nodes: %w", err)
	}

	for _, node := range nodes {
		if node.Type == "trigger.webhook" || node.Type == "webhook" {
			if err := h.registerEndpoint(ctx, wf, node); err != nil {
				// We log but continue for other nodes if any
				fmt.Printf("Error registering webhook for node %s: %v\n", node.ID, err)
			}
		}
	}

	return nil
}

func (h *WebhookAutoRegHandler) handleDeactivation(ctx context.Context, e events.WorkflowDeactivated) error {
	endpoints, err := h.webhookRepo.FindByWorkflowID(ctx, e.WorkflowID)
	if err != nil {
		return err
	}

	for _, ep := range endpoints {
		if ep.IsActive {
			ep.Deactivate()
			if err := h.webhookRepo.Update(ctx, &ep); err != nil {
				fmt.Printf("Error deactivating endpoint %s: %v\n", ep.ID, err)
			}
		}
	}

	return nil
}

func (h *WebhookAutoRegHandler) registerEndpoint(ctx context.Context, wf *workflow.Workflow, node workflow.Node) error {
	params := node.Parameters
	path, _ := params["path"].(string)
	if path == "" {
		return fmt.Errorf("webhook path is required")
	}

	// Ensure path starts with /
	if path[0] != '/' {
		path = "/" + path
	}

	method, _ := params["method"].(string)
	if method == "" {
		method = "POST"
	}

	// Check if endpoint already exists for this node
	existing, err := h.webhookRepo.FindByWorkflowAndNode(ctx, wf.ID, node.ID)
	if err == nil && existing != nil {
		// Update existing
		existing.UpdatePath(path)
		existing.WithMethod(method)
		existing.Activate()
		h.applySecurityParams(existing, params)
		return h.webhookRepo.Update(ctx, existing)
	}

	// Create new
	endpoint, err := webhook.NewEndpoint(wf.ID, wf.WorkspaceID, node.ID, path)
	if err != nil {
		return err
	}
	endpoint.WithMethod(method)
	endpoint.Activate()
	h.applySecurityParams(endpoint, params)

	// If path already exists for ANOTHER workflow, we might have a conflict
	// webhookRepo.Create will handle unique constraint error
	return h.webhookRepo.Create(ctx, endpoint)
}

func (h *WebhookAutoRegHandler) applySecurityParams(ep *webhook.Endpoint, params map[string]interface{}) {
	authType, _ := params["authentication"].(string)

	// Reset security first
	ep.RequireSignature = false
	ep.RequireTimestamp = false
	ep.RequireNonce = false
	ep.AllowedIPs = ""

	if authType == "signature" {
		ep.RequireSignature = true
		secret, _ := params["secret"].(string)
		if secret != "" {
			ep.WithSecret(secret)
		}
		header, _ := params["signature_header"].(string)
		if header != "" {
			ep.SignatureHeader = header
		}
	}

	// IP Allowlist
	if allowed, ok := params["allowed_ips"].([]interface{}); ok {
		ips := make([]string, 0)
		for _, v := range allowed {
			if s, ok := v.(string); ok {
				ips = append(ips, s)
			}
		}
		ep.SetAllowedIPs(ips)
	}
}
