package integrations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// TwilioNode handles Twilio SMS and Voice operations
type TwilioNode struct{}

func (n *TwilioNode) Type() string {
	return "integrations.twilio"
}

func (n *TwilioNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	accountSID := core.GetString(config, "accountSid", "")
	authToken := core.GetString(config, "authToken", "")

	if accountSID == "" || authToken == "" {
		return nil, fmt.Errorf("accountSid and authToken are required")
	}

	operation := core.GetString(config, "operation", "sendSms")

	switch operation {
	case "sendSms":
		return n.sendSMS(ctx, config, accountSID, authToken)
	case "sendMms":
		return n.sendMMS(ctx, config, accountSID, authToken)
	case "makeCall":
		return n.makeCall(ctx, config, accountSID, authToken)
	case "getMessages":
		return n.getMessages(ctx, config, accountSID, authToken)
	case "getMessage":
		return n.getMessage(ctx, config, accountSID, authToken)
	case "getCalls":
		return n.getCalls(ctx, config, accountSID, authToken)
	case "lookupNumber":
		return n.lookupNumber(ctx, config, accountSID, authToken)
	default:
		return n.sendSMS(ctx, config, accountSID, authToken)
	}
}

func (n *TwilioNode) sendSMS(ctx context.Context, config map[string]interface{}, accountSID, authToken string) (map[string]interface{}, error) {
	from := core.GetString(config, "from", "")
	to := core.GetString(config, "to", "")
	body := core.GetString(config, "body", "")

	if from == "" || to == "" || body == "" {
		return nil, fmt.Errorf("from, to, and body are required")
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	data := url.Values{}
	data.Set("From", from)
	data.Set("To", to)
	data.Set("Body", body)

	// Optional callback URL
	if statusCallback := core.GetString(config, "statusCallback", ""); statusCallback != "" {
		data.Set("StatusCallback", statusCallback)
	}

	result, err := n.makeRequest(ctx, "POST", endpoint, data, accountSID, authToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sid":         result["sid"],
		"status":      result["status"],
		"to":          result["to"],
		"from":        result["from"],
		"body":        result["body"],
		"dateCreated": result["date_created"],
		"price":       result["price"],
		"priceUnit":   result["price_unit"],
	}, nil
}

func (n *TwilioNode) sendMMS(ctx context.Context, config map[string]interface{}, accountSID, authToken string) (map[string]interface{}, error) {
	from := core.GetString(config, "from", "")
	to := core.GetString(config, "to", "")
	body := core.GetString(config, "body", "")
	mediaUrl := core.GetString(config, "mediaUrl", "")

	if from == "" || to == "" {
		return nil, fmt.Errorf("from and to are required")
	}
	if body == "" && mediaUrl == "" {
		return nil, fmt.Errorf("body or mediaUrl is required")
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	data := url.Values{}
	data.Set("From", from)
	data.Set("To", to)
	if body != "" {
		data.Set("Body", body)
	}
	if mediaUrl != "" {
		data.Set("MediaUrl", mediaUrl)
	}

	result, err := n.makeRequest(ctx, "POST", endpoint, data, accountSID, authToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sid":       result["sid"],
		"status":    result["status"],
		"to":        result["to"],
		"from":      result["from"],
		"body":      result["body"],
		"numMedia":  result["num_media"],
	}, nil
}

func (n *TwilioNode) makeCall(ctx context.Context, config map[string]interface{}, accountSID, authToken string) (map[string]interface{}, error) {
	from := core.GetString(config, "from", "")
	to := core.GetString(config, "to", "")
	twimlUrl := core.GetString(config, "url", "")
	twiml := core.GetString(config, "twiml", "")

	if from == "" || to == "" {
		return nil, fmt.Errorf("from and to are required")
	}
	if twimlUrl == "" && twiml == "" {
		return nil, fmt.Errorf("url or twiml is required")
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Calls.json", accountSID)

	data := url.Values{}
	data.Set("From", from)
	data.Set("To", to)
	if twimlUrl != "" {
		data.Set("Url", twimlUrl)
	}
	if twiml != "" {
		data.Set("Twiml", twiml)
	}

	// Optional parameters
	if statusCallback := core.GetString(config, "statusCallback", ""); statusCallback != "" {
		data.Set("StatusCallback", statusCallback)
	}
	if timeout := core.GetInt(config, "timeout", 0); timeout > 0 {
		data.Set("Timeout", fmt.Sprintf("%d", timeout))
	}

	result, err := n.makeRequest(ctx, "POST", endpoint, data, accountSID, authToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sid":         result["sid"],
		"status":      result["status"],
		"to":          result["to"],
		"from":        result["from"],
		"dateCreated": result["date_created"],
		"direction":   result["direction"],
	}, nil
}

func (n *TwilioNode) getMessages(ctx context.Context, config map[string]interface{}, accountSID, authToken string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	params := url.Values{}
	if to := core.GetString(config, "to", ""); to != "" {
		params.Set("To", to)
	}
	if from := core.GetString(config, "from", ""); from != "" {
		params.Set("From", from)
	}
	if pageSize := core.GetInt(config, "pageSize", 0); pageSize > 0 {
		params.Set("PageSize", fmt.Sprintf("%d", pageSize))
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	result, err := n.makeRequest(ctx, "GET", endpoint, nil, accountSID, authToken)
	if err != nil {
		return nil, err
	}

	messages := result["messages"]
	if messages == nil {
		messages = []interface{}{}
	}

	return map[string]interface{}{
		"messages": messages,
	}, nil
}

func (n *TwilioNode) getMessage(ctx context.Context, config map[string]interface{}, accountSID, authToken string) (map[string]interface{}, error) {
	messageSid := core.GetString(config, "messageSid", "")
	if messageSid == "" {
		return nil, fmt.Errorf("messageSid is required")
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages/%s.json", accountSID, messageSid)

	result, err := n.makeRequest(ctx, "GET", endpoint, nil, accountSID, authToken)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (n *TwilioNode) getCalls(ctx context.Context, config map[string]interface{}, accountSID, authToken string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Calls.json", accountSID)

	params := url.Values{}
	if to := core.GetString(config, "to", ""); to != "" {
		params.Set("To", to)
	}
	if from := core.GetString(config, "from", ""); from != "" {
		params.Set("From", from)
	}
	if status := core.GetString(config, "status", ""); status != "" {
		params.Set("Status", status)
	}
	if pageSize := core.GetInt(config, "pageSize", 0); pageSize > 0 {
		params.Set("PageSize", fmt.Sprintf("%d", pageSize))
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	result, err := n.makeRequest(ctx, "GET", endpoint, nil, accountSID, authToken)
	if err != nil {
		return nil, err
	}

	calls := result["calls"]
	if calls == nil {
		calls = []interface{}{}
	}

	return map[string]interface{}{
		"calls": calls,
	}, nil
}

func (n *TwilioNode) lookupNumber(ctx context.Context, config map[string]interface{}, accountSID, authToken string) (map[string]interface{}, error) {
	phoneNumber := core.GetString(config, "phoneNumber", "")
	if phoneNumber == "" {
		return nil, fmt.Errorf("phoneNumber is required")
	}

	// Lookup API v2
	endpoint := fmt.Sprintf("https://lookups.twilio.com/v2/PhoneNumbers/%s", url.PathEscape(phoneNumber))

	params := url.Values{}
	fields := core.GetString(config, "fields", "")
	if fields != "" {
		params.Set("Fields", fields)
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	result, err := n.makeRequest(ctx, "GET", endpoint, nil, accountSID, authToken)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (n *TwilioNode) makeRequest(ctx context.Context, method, endpoint string, data url.Values, accountSID, authToken string) (map[string]interface{}, error) {
	var req *http.Request
	var err error

	if data != nil && method == "POST" {
		req, err = http.NewRequestWithContext(ctx, method, endpoint, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	// Basic auth
	auth := base64.StdEncoding.EncodeToString([]byte(accountSID + ":" + authToken))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		msg := "Twilio API error"
		if errMsg, ok := result["message"].(string); ok {
			msg = errMsg
		}
		return nil, fmt.Errorf("%s (status %d)", msg, resp.StatusCode)
	}

	return result, nil
}

// Note: TwilioNode is registered in integrations/init.go
