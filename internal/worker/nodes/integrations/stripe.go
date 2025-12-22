package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// StripeNode handles Stripe payment operations
type StripeNode struct{}

func (n *StripeNode) Type() string {
	return "integration.stripe"
}

func (n *StripeNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	resource := getString(config, "resource", "customers")
	operation := getString(config, "operation", "list")

	// Get credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("Stripe credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	apiKey := getStringFromMap(cred.Data, "secretKey", getStringFromMap(cred.Data, "api_key", ""))
	if apiKey == "" {
		apiKey = cred.APIKey
	}

	if apiKey == "" {
		return nil, fmt.Errorf("Stripe secret key is required")
	}

	switch resource {
	case "customers":
		return n.handleCustomers(ctx, operation, config, execCtx.Input, apiKey)
	case "charges":
		return n.handleCharges(ctx, operation, config, execCtx.Input, apiKey)
	case "paymentIntents":
		return n.handlePaymentIntents(ctx, operation, config, execCtx.Input, apiKey)
	case "subscriptions":
		return n.handleSubscriptions(ctx, operation, config, execCtx.Input, apiKey)
	case "invoices":
		return n.handleInvoices(ctx, operation, config, execCtx.Input, apiKey)
	case "products":
		return n.handleProducts(ctx, operation, config, execCtx.Input, apiKey)
	case "prices":
		return n.handlePrices(ctx, operation, config, execCtx.Input, apiKey)
	case "refunds":
		return n.handleRefunds(ctx, operation, config, execCtx.Input, apiKey)
	default:
		return nil, fmt.Errorf("unsupported resource: %s", resource)
	}
}

func (n *StripeNode) handleCustomers(ctx context.Context, operation string, config map[string]interface{}, input map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	switch operation {
	case "list":
		return n.makeRequest(ctx, "GET", "customers", nil, apiKey, config)
	case "get":
		id := getString(config, "customerId", "")
		if id == "" {
			return nil, fmt.Errorf("customerId is required")
		}
		return n.makeRequest(ctx, "GET", "customers/"+id, nil, apiKey, nil)
	case "create":
		data := n.buildCustomerData(config, input)
		return n.makeRequest(ctx, "POST", "customers", data, apiKey, nil)
	case "update":
		id := getString(config, "customerId", "")
		if id == "" {
			return nil, fmt.Errorf("customerId is required")
		}
		data := n.buildCustomerData(config, input)
		return n.makeRequest(ctx, "POST", "customers/"+id, data, apiKey, nil)
	case "delete":
		id := getString(config, "customerId", "")
		if id == "" {
			return nil, fmt.Errorf("customerId is required")
		}
		return n.makeRequest(ctx, "DELETE", "customers/"+id, nil, apiKey, nil)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *StripeNode) handleCharges(ctx context.Context, operation string, config map[string]interface{}, input map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	switch operation {
	case "list":
		return n.makeRequest(ctx, "GET", "charges", nil, apiKey, config)
	case "get":
		id := getString(config, "chargeId", "")
		if id == "" {
			return nil, fmt.Errorf("chargeId is required")
		}
		return n.makeRequest(ctx, "GET", "charges/"+id, nil, apiKey, nil)
	case "create":
		data := url.Values{}
		data.Set("amount", fmt.Sprintf("%d", getInt(config, "amount", 0)))
		data.Set("currency", getString(config, "currency", "usd"))
		if customer := getString(config, "customer", ""); customer != "" {
			data.Set("customer", customer)
		}
		if source := getString(config, "source", ""); source != "" {
			data.Set("source", source)
		}
		if desc := getString(config, "description", ""); desc != "" {
			data.Set("description", desc)
		}
		return n.makeRequest(ctx, "POST", "charges", data, apiKey, nil)
	case "capture":
		id := getString(config, "chargeId", "")
		if id == "" {
			return nil, fmt.Errorf("chargeId is required")
		}
		return n.makeRequest(ctx, "POST", "charges/"+id+"/capture", nil, apiKey, nil)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *StripeNode) handlePaymentIntents(ctx context.Context, operation string, config map[string]interface{}, input map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	switch operation {
	case "list":
		return n.makeRequest(ctx, "GET", "payment_intents", nil, apiKey, config)
	case "get":
		id := getString(config, "paymentIntentId", "")
		if id == "" {
			return nil, fmt.Errorf("paymentIntentId is required")
		}
		return n.makeRequest(ctx, "GET", "payment_intents/"+id, nil, apiKey, nil)
	case "create":
		data := url.Values{}
		data.Set("amount", fmt.Sprintf("%d", getInt(config, "amount", 0)))
		data.Set("currency", getString(config, "currency", "usd"))
		if customer := getString(config, "customer", ""); customer != "" {
			data.Set("customer", customer)
		}
		if desc := getString(config, "description", ""); desc != "" {
			data.Set("description", desc)
		}
		if paymentMethod := getString(config, "paymentMethod", ""); paymentMethod != "" {
			data.Set("payment_method", paymentMethod)
		}
		if getBool(config, "confirmImmediately", false) {
			data.Set("confirm", "true")
		}
		return n.makeRequest(ctx, "POST", "payment_intents", data, apiKey, nil)
	case "confirm":
		id := getString(config, "paymentIntentId", "")
		if id == "" {
			return nil, fmt.Errorf("paymentIntentId is required")
		}
		return n.makeRequest(ctx, "POST", "payment_intents/"+id+"/confirm", nil, apiKey, nil)
	case "cancel":
		id := getString(config, "paymentIntentId", "")
		if id == "" {
			return nil, fmt.Errorf("paymentIntentId is required")
		}
		return n.makeRequest(ctx, "POST", "payment_intents/"+id+"/cancel", nil, apiKey, nil)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *StripeNode) handleSubscriptions(ctx context.Context, operation string, config map[string]interface{}, input map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	switch operation {
	case "list":
		return n.makeRequest(ctx, "GET", "subscriptions", nil, apiKey, config)
	case "get":
		id := getString(config, "subscriptionId", "")
		if id == "" {
			return nil, fmt.Errorf("subscriptionId is required")
		}
		return n.makeRequest(ctx, "GET", "subscriptions/"+id, nil, apiKey, nil)
	case "create":
		data := url.Values{}
		data.Set("customer", getString(config, "customer", ""))
		if priceId := getString(config, "priceId", ""); priceId != "" {
			data.Set("items[0][price]", priceId)
		}
		if trialDays := getInt(config, "trialPeriodDays", 0); trialDays > 0 {
			data.Set("trial_period_days", fmt.Sprintf("%d", trialDays))
		}
		return n.makeRequest(ctx, "POST", "subscriptions", data, apiKey, nil)
	case "update":
		id := getString(config, "subscriptionId", "")
		if id == "" {
			return nil, fmt.Errorf("subscriptionId is required")
		}
		data := url.Values{}
		if priceId := getString(config, "priceId", ""); priceId != "" {
			data.Set("items[0][price]", priceId)
		}
		return n.makeRequest(ctx, "POST", "subscriptions/"+id, data, apiKey, nil)
	case "cancel":
		id := getString(config, "subscriptionId", "")
		if id == "" {
			return nil, fmt.Errorf("subscriptionId is required")
		}
		return n.makeRequest(ctx, "DELETE", "subscriptions/"+id, nil, apiKey, nil)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *StripeNode) handleInvoices(ctx context.Context, operation string, config map[string]interface{}, input map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	switch operation {
	case "list":
		return n.makeRequest(ctx, "GET", "invoices", nil, apiKey, config)
	case "get":
		id := getString(config, "invoiceId", "")
		if id == "" {
			return nil, fmt.Errorf("invoiceId is required")
		}
		return n.makeRequest(ctx, "GET", "invoices/"+id, nil, apiKey, nil)
	case "create":
		data := url.Values{}
		data.Set("customer", getString(config, "customer", ""))
		if desc := getString(config, "description", ""); desc != "" {
			data.Set("description", desc)
		}
		return n.makeRequest(ctx, "POST", "invoices", data, apiKey, nil)
	case "finalize":
		id := getString(config, "invoiceId", "")
		if id == "" {
			return nil, fmt.Errorf("invoiceId is required")
		}
		return n.makeRequest(ctx, "POST", "invoices/"+id+"/finalize", nil, apiKey, nil)
	case "pay":
		id := getString(config, "invoiceId", "")
		if id == "" {
			return nil, fmt.Errorf("invoiceId is required")
		}
		return n.makeRequest(ctx, "POST", "invoices/"+id+"/pay", nil, apiKey, nil)
	case "void":
		id := getString(config, "invoiceId", "")
		if id == "" {
			return nil, fmt.Errorf("invoiceId is required")
		}
		return n.makeRequest(ctx, "POST", "invoices/"+id+"/void", nil, apiKey, nil)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *StripeNode) handleProducts(ctx context.Context, operation string, config map[string]interface{}, input map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	switch operation {
	case "list":
		return n.makeRequest(ctx, "GET", "products", nil, apiKey, config)
	case "get":
		id := getString(config, "productId", "")
		if id == "" {
			return nil, fmt.Errorf("productId is required")
		}
		return n.makeRequest(ctx, "GET", "products/"+id, nil, apiKey, nil)
	case "create":
		data := url.Values{}
		data.Set("name", getString(config, "name", ""))
		if desc := getString(config, "description", ""); desc != "" {
			data.Set("description", desc)
		}
		return n.makeRequest(ctx, "POST", "products", data, apiKey, nil)
	case "update":
		id := getString(config, "productId", "")
		if id == "" {
			return nil, fmt.Errorf("productId is required")
		}
		data := url.Values{}
		if name := getString(config, "name", ""); name != "" {
			data.Set("name", name)
		}
		if desc := getString(config, "description", ""); desc != "" {
			data.Set("description", desc)
		}
		return n.makeRequest(ctx, "POST", "products/"+id, data, apiKey, nil)
	case "delete":
		id := getString(config, "productId", "")
		if id == "" {
			return nil, fmt.Errorf("productId is required")
		}
		return n.makeRequest(ctx, "DELETE", "products/"+id, nil, apiKey, nil)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *StripeNode) handlePrices(ctx context.Context, operation string, config map[string]interface{}, input map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	switch operation {
	case "list":
		return n.makeRequest(ctx, "GET", "prices", nil, apiKey, config)
	case "get":
		id := getString(config, "priceId", "")
		if id == "" {
			return nil, fmt.Errorf("priceId is required")
		}
		return n.makeRequest(ctx, "GET", "prices/"+id, nil, apiKey, nil)
	case "create":
		data := url.Values{}
		data.Set("product", getString(config, "product", ""))
		data.Set("unit_amount", fmt.Sprintf("%d", getInt(config, "unitAmount", 0)))
		data.Set("currency", getString(config, "currency", "usd"))
		if recurring := getString(config, "recurring", ""); recurring != "" {
			data.Set("recurring[interval]", recurring)
		}
		return n.makeRequest(ctx, "POST", "prices", data, apiKey, nil)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *StripeNode) handleRefunds(ctx context.Context, operation string, config map[string]interface{}, input map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	switch operation {
	case "list":
		return n.makeRequest(ctx, "GET", "refunds", nil, apiKey, config)
	case "get":
		id := getString(config, "refundId", "")
		if id == "" {
			return nil, fmt.Errorf("refundId is required")
		}
		return n.makeRequest(ctx, "GET", "refunds/"+id, nil, apiKey, nil)
	case "create":
		data := url.Values{}
		if charge := getString(config, "charge", ""); charge != "" {
			data.Set("charge", charge)
		}
		if paymentIntent := getString(config, "paymentIntent", ""); paymentIntent != "" {
			data.Set("payment_intent", paymentIntent)
		}
		if amount := getInt(config, "amount", 0); amount > 0 {
			data.Set("amount", fmt.Sprintf("%d", amount))
		}
		return n.makeRequest(ctx, "POST", "refunds", data, apiKey, nil)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *StripeNode) buildCustomerData(config map[string]interface{}, input map[string]interface{}) url.Values {
	data := url.Values{}

	if email := getString(config, "email", ""); email != "" {
		data.Set("email", email)
	}
	if name := getString(config, "name", ""); name != "" {
		data.Set("name", name)
	}
	if phone := getString(config, "phone", ""); phone != "" {
		data.Set("phone", phone)
	}
	if desc := getString(config, "description", ""); desc != "" {
		data.Set("description", desc)
	}

	// Metadata
	if metadata, ok := config["metadata"].(map[string]interface{}); ok {
		for k, v := range metadata {
			data.Set(fmt.Sprintf("metadata[%s]", k), fmt.Sprintf("%v", v))
		}
	}

	return data
}

func (n *StripeNode) makeRequest(ctx context.Context, method, endpoint string, data url.Values, apiKey string, queryParams map[string]interface{}) (map[string]interface{}, error) {
	baseURL := "https://api.stripe.com/v1/"
	fullURL := baseURL + endpoint

	// Add query parameters for GET requests
	if method == "GET" && queryParams != nil {
		params := url.Values{}
		if limit := getInt(queryParams, "limit", 0); limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", limit))
		}
		if customer := getString(queryParams, "customer", ""); customer != "" {
			params.Set("customer", customer)
		}
		if startingAfter := getString(queryParams, "startingAfter", ""); startingAfter != "" {
			params.Set("starting_after", startingAfter)
		}
		if len(params) > 0 {
			fullURL += "?" + params.Encode()
		}
	}

	var bodyReader io.Reader
	if len(data) > 0 {
		bodyReader = strings.NewReader(data.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if errObj, ok := result["error"].(map[string]interface{}); ok {
		msg := getString(errObj, "message", "Unknown error")
		return nil, fmt.Errorf("Stripe error: %s", msg)
	}

	return result, nil
}
