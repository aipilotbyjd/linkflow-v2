package actions

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type HTTPRequestNode struct{}

func (n *HTTPRequestNode) Type() string {
	return "action.http"
}

func (n *HTTPRequestNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	method := getStringHTTP(config, "method", "GET")
	urlStr := getStringHTTP(config, "url", "")
	headers := getMapHTTP(config, "headers")
	queryParams := getMapHTTP(config, "queryParams")
	body := config["body"]
	bodyType := getStringHTTP(config, "bodyType", "json")
	timeout := getIntHTTP(config, "timeout", 30)
	followRedirects := getBoolHTTP(config, "followRedirects", true)
	ignoreSsl := getBoolHTTP(config, "ignoreSsl", false)
	authType := getStringHTTP(config, "authType", "none")

	// Build URL with query params
	if len(queryParams) > 0 {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		u.RawQuery = q.Encode()
		urlStr = u.String()
	}

	// Build request body
	var reqBody io.Reader
	var contentType string

	if body != nil && method != "GET" && method != "HEAD" {
		switch bodyType {
		case "json":
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
			}
			reqBody = bytes.NewReader(jsonBody)
			contentType = "application/json"

		case "form":
			if formData, ok := body.(map[string]interface{}); ok {
				form := url.Values{}
				for k, v := range formData {
					form.Set(k, fmt.Sprintf("%v", v))
				}
				reqBody = strings.NewReader(form.Encode())
				contentType = "application/x-www-form-urlencoded"
			}

		case "multipart":
			// TODO: Handle multipart form data
			contentType = "multipart/form-data"

		case "raw":
			if str, ok := body.(string); ok {
				reqBody = strings.NewReader(str)
			}

		case "binary":
			// TODO: Handle binary data
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, urlStr, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, fmt.Sprintf("%v", v))
	}

	// Set authentication
	switch authType {
	case "basic":
		username := getStringHTTP(config, "username", "")
		password := getStringHTTP(config, "password", "")
		req.SetBasicAuth(username, password)

	case "bearer":
		token := getStringHTTP(config, "token", "")
		req.Header.Set("Authorization", "Bearer "+token)

	case "apiKey":
		apiKey := getStringHTTP(config, "apiKey", "")
		apiKeyName := getStringHTTP(config, "apiKeyName", "X-API-Key")
		apiKeyLocation := getStringHTTP(config, "apiKeyLocation", "header")
		if apiKeyLocation == "header" {
			req.Header.Set(apiKeyName, apiKey)
		} else if apiKeyLocation == "query" {
			q := req.URL.Query()
			q.Set(apiKeyName, apiKey)
			req.URL.RawQuery = q.Encode()
		}

	case "oauth2":
		// Get credential and use access token
		if credID := getStringHTTP(config, "credentialId", ""); credID != "" {
			// TODO: Get credential from context
		}

	case "digest":
		// TODO: Implement digest auth

	case "none":
		// No authentication
	}

	// Create HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreSsl},
	}

	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response headers
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	// Try to parse JSON response
	var jsonData interface{}
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		json.Unmarshal(respBody, &jsonData)
	}

	result := map[string]interface{}{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    respHeaders,
		"body":       string(respBody),
		"json":       jsonData,
		"duration":   duration,
		"ok":         resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	// Check for error status codes if configured
	if getBoolHTTP(config, "throwOnError", false) && resp.StatusCode >= 400 {
		return result, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return result, nil
}

// Helper functions
func getStringHTTP(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}

func getIntHTTP(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return defaultVal
}

func getBoolHTTP(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultVal
}

func getMapHTTP(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return make(map[string]interface{})
}
