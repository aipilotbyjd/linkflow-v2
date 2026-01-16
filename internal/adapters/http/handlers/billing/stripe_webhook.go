package billing

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

type StripeWebhookHandler struct {
	subscriptionRepo billing.SubscriptionRepository
	invoiceRepo      billing.InvoiceRepository
	webhookSecret    string
}

func NewStripeWebhookHandler(
	subscriptionRepo billing.SubscriptionRepository,
	invoiceRepo billing.InvoiceRepository,
	webhookSecret string,
) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		subscriptionRepo: subscriptionRepo,
		invoiceRepo:      invoiceRepo,
		webhookSecret:    webhookSecret,
	}
}

func (h *StripeWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		common.BadRequest(w, "failed to read request body")
		return
	}

	// TODO: Verify Stripe signature using h.webhookSecret

	var event struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		common.BadRequest(w, "invalid webhook payload")
		return
	}

	switch event.Type {
	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		// Handle subscription events
	case "invoice.paid",
		"invoice.payment_failed",
		"invoice.finalized":
		// Handle invoice events
	case "payment_intent.succeeded",
		"payment_intent.payment_failed":
		// Handle payment events
	}

	common.Success(w, map[string]string{"received": "true"})
}
