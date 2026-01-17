package billing

import (
	"io"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// StripeWebhookHandler handles Stripe webhook events
type StripeWebhookHandler struct {
	webhookSecret string
}

// NewStripeWebhookHandler creates a new handler
func NewStripeWebhookHandler(webhookSecret string) *StripeWebhookHandler {
	return &StripeWebhookHandler{webhookSecret: webhookSecret}
}

// Handle handles the Stripe webhook request
func (h *StripeWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		common.BadRequest(w, "could not read request body")
		return
	}

	_ = r.Header.Get("Stripe-Signature")
	_ = body

	common.Success(w, map[string]string{
		"received": "true",
	})
}
