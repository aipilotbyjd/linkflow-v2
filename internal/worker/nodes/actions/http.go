package actions

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			
			if formData, ok := body.(map[string]interface{}); ok {
				for k, v := range formData {
					switch val := v.(type) {
					case map[string]interface{}:
						// File upload: {"filename": "...", "content": "base64...", "contentType": "..."}
						if filename, ok := val["filename"].(string); ok {
							part, err := writer.CreateFormFile(k, filename)
							if err != nil {
								return nil, fmt.Errorf("failed to create form file: %w", err)
							}
							
							if content, ok := val["content"].(string); ok {
								// Check if base64 encoded
								if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
									_, _ = part.Write(decoded)
								} else {
									_, _ = part.Write([]byte(content))
								}
							} else if filePath, ok := val["path"].(string); ok {
								// Read from file path
								fileContent, err := os.ReadFile(filePath)
								if err != nil {
									return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
								}
								_, _ = part.Write(fileContent)
							}
						}
					default:
						_ = writer.WriteField(k, fmt.Sprintf("%v", v))
					}
				}
			}
			writer.Close()
			reqBody = &buf
			contentType = writer.FormDataContentType()

		case "raw":
			if str, ok := body.(string); ok {
				reqBody = strings.NewReader(str)
			}

		case "binary":
			switch val := body.(type) {
			case string:
				// Base64 encoded binary
				if decoded, err := base64.StdEncoding.DecodeString(val); err == nil {
					reqBody = bytes.NewReader(decoded)
				} else {
					reqBody = strings.NewReader(val)
				}
			case map[string]interface{}:
				// File reference: {"path": "/path/to/file"}
				if filePath, ok := val["path"].(string); ok {
					fileContent, err := os.ReadFile(filePath)
					if err != nil {
						return nil, fmt.Errorf("failed to read binary file: %w", err)
					}
					reqBody = bytes.NewReader(fileContent)
					// Try to detect content type
					ext := strings.ToLower(filepath.Ext(filePath))
					switch ext {
					case ".pdf":
						contentType = "application/pdf"
					case ".png":
						contentType = "image/png"
					case ".jpg", ".jpeg":
						contentType = "image/jpeg"
					case ".gif":
						contentType = "image/gif"
					case ".zip":
						contentType = "application/zip"
					default:
						contentType = "application/octet-stream"
					}
				}
			}
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
		// Get access token from config (should be resolved from credential)
		token := getStringHTTP(config, "accessToken", "")
		if token == "" {
			token = getStringHTTP(config, "token", "")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

	case "digest":
		// Digest authentication requires a challenge-response
		username := getStringHTTP(config, "username", "")
		password := getStringHTTP(config, "password", "")
		if username != "" && password != "" {
			// First request to get WWW-Authenticate header
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreSsl},
			}
			probeClient := &http.Client{
				Timeout:   time.Duration(timeout) * time.Second,
				Transport: transport,
			}
			probeReq, _ := http.NewRequestWithContext(ctx, method, urlStr, nil)
			probeResp, err := probeClient.Do(probeReq)
			if err == nil && probeResp.StatusCode == 401 {
				wwwAuth := probeResp.Header.Get("WWW-Authenticate")
				probeResp.Body.Close()
				if strings.HasPrefix(wwwAuth, "Digest ") {
					authHeader := computeDigestAuth(wwwAuth, username, password, method, urlStr)
					req.Header.Set("Authorization", authHeader)
				}
			} else if probeResp != nil {
				probeResp.Body.Close()
			}
		}

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
		_ = json.Unmarshal(respBody, &jsonData)
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

// computeDigestAuth computes the Digest authentication header
func computeDigestAuth(wwwAuth, username, password, method, uri string) string {
	params := parseDigestChallenge(wwwAuth)
	
	realm := params["realm"]
	nonce := params["nonce"]
	qop := params["qop"]
	opaque := params["opaque"]
	algorithm := params["algorithm"]
	if algorithm == "" {
		algorithm = "MD5"
	}

	// Parse URI path
	parsedURL, _ := url.Parse(uri)
	digestURI := parsedURL.RequestURI()
	
	// Generate client nonce and nonce count
	cnonce := hex.EncodeToString(md5.New().Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))[:16]
	nc := "00000001"
	
	// Compute HA1 = MD5(username:realm:password)
	ha1 := md5Hash(fmt.Sprintf("%s:%s:%s", username, realm, password))
	
	// Compute HA2 = MD5(method:digestURI)
	ha2 := md5Hash(fmt.Sprintf("%s:%s", method, digestURI))
	
	// Compute response
	var response string
	if qop == "auth" || qop == "auth-int" {
		response = md5Hash(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, nonce, nc, cnonce, qop, ha2))
	} else {
		response = md5Hash(fmt.Sprintf("%s:%s:%s", ha1, nonce, ha2))
	}
	
	// Build Authorization header
	auth := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
		username, realm, nonce, digestURI, response)
	
	if qop != "" {
		auth += fmt.Sprintf(`, qop=%s, nc=%s, cnonce="%s"`, qop, nc, cnonce)
	}
	if opaque != "" {
		auth += fmt.Sprintf(`, opaque="%s"`, opaque)
	}
	if algorithm != "" {
		auth += fmt.Sprintf(`, algorithm=%s`, algorithm)
	}
	
	return auth
}

// parseDigestChallenge parses the WWW-Authenticate header
func parseDigestChallenge(header string) map[string]string {
	params := make(map[string]string)
	header = strings.TrimPrefix(header, "Digest ")
	
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, "="); idx > 0 {
			key := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])
			value = strings.Trim(value, `"`)
			params[key] = value
		}
	}
	return params
}

// md5Hash computes MD5 hash
func md5Hash(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
