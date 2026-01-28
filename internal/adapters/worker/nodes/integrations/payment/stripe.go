package payment

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type StripeNode struct{}

func NewStripeNode() *StripeNode {
	return &StripeNode{}
}

func (n *StripeNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	resource, _ := params["resource"].(string)
	operation, _ := params["operation"].(string)

	switch resource {
	case "customer":
		return n.handleCustomer(ctx, operation, params)
	case "payment_intent":
		return n.handlePaymentIntent(ctx, operation, params)
	case "subscription":
		return n.handleSubscription(ctx, operation, params)
	default:
		return nil, fmt.Errorf("unsupported Stripe resource: %s", resource)
	}
}

func (n *StripeNode) handleCustomer(ctx context.Context, operation string, params map[string]interface{}) (types.JSON, error) {
	switch operation {
	case "create":
		email, _ := params["email"].(string)
		name, _ := params["name"].(string)
		return types.JSON{"id": "", "email": email, "name": name}, nil
	case "get":
		customerID, _ := params["customer_id"].(string)
		return types.JSON{"id": customerID, "customer": nil}, nil
	default:
		return nil, fmt.Errorf("unsupported customer operation: %s", operation)
	}
}

func (n *StripeNode) handlePaymentIntent(ctx context.Context, operation string, params map[string]interface{}) (types.JSON, error) {
	switch operation {
	case "create":
		amount, _ := params["amount"].(float64)
		currency, _ := params["currency"].(string)
		return types.JSON{"id": "", "amount": int(amount), "currency": currency, "status": "requires_payment_method"}, nil
	case "confirm":
		paymentIntentID, _ := params["payment_intent_id"].(string)
		return types.JSON{"id": paymentIntentID, "status": "succeeded", "success": true}, nil
	default:
		return nil, fmt.Errorf("unsupported payment_intent operation: %s", operation)
	}
}

func (n *StripeNode) handleSubscription(ctx context.Context, operation string, params map[string]interface{}) (types.JSON, error) {
	switch operation {
	case "create":
		customerID, _ := params["customer_id"].(string)
		priceID, _ := params["price_id"].(string)
		return types.JSON{"id": "", "customer_id": customerID, "price_id": priceID, "status": "active"}, nil
	case "cancel":
		subscriptionID, _ := params["subscription_id"].(string)
		return types.JSON{"id": subscriptionID, "status": "canceled", "canceled": true}, nil
	default:
		return nil, fmt.Errorf("unsupported subscription operation: %s", operation)
	}
}

func (n *StripeNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.stripe",
		Name:        "Stripe",
		Description: "Process payments with Stripe",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "CreditCard",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
