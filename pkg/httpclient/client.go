// Package httpclient provides a configurable HTTP client with retry and timeout support.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a configurable HTTP client
type Client struct {
	httpClient *http.Client
	baseURL    string
	headers    map[string]string
	retries    int
	retryDelay time.Duration
}

// Config holds client configuration
type Config struct {
	BaseURL       string
	Timeout       time.Duration
	Headers       map[string]string
	Retries       int
	RetryDelay    time.Duration
	Transport     http.RoundTripper
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Timeout:    30 * time.Second,
		Retries:    3,
		RetryDelay: time.Second,
	}
}

// New creates a new HTTP client
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	transport := cfg.Transport
	if transport == nil {
		transport = &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		baseURL:    cfg.BaseURL,
		headers:    cfg.Headers,
		retries:    cfg.Retries,
		retryDelay: cfg.RetryDelay,
	}
}

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Query   map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Do executes an HTTP request
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	var body io.Reader
	if req.Body != nil {
		switch v := req.Body.(type) {
		case []byte:
			body = bytes.NewReader(v)
		case string:
			body = bytes.NewReader([]byte(v))
		case io.Reader:
			body = v
		default:
			jsonBody, err := json.Marshal(req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = bytes.NewReader(jsonBody)
		}
	}

	url := req.URL
	if c.baseURL != "" && url[0] == '/' {
		url = c.baseURL + url
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	// Set request headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set query parameters
	if len(req.Query) > 0 {
		q := httpReq.URL.Query()
		for k, v := range req.Query {
			q.Set(k, v)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Set Content-Type if body is present and not set
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Execute with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}
		}

		resp, lastErr = c.httpClient.Do(httpReq)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{Method: http.MethodGet, URL: url, Headers: headers})
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{Method: http.MethodPost, URL: url, Body: body, Headers: headers})
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{Method: http.MethodPut, URL: url, Body: body, Headers: headers})
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{Method: http.MethodPatch, URL: url, Body: body, Headers: headers})
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{Method: http.MethodDelete, URL: url, Headers: headers})
}

// JSON decodes the response body as JSON
func (r *Response) JSON(dest interface{}) error {
	return json.Unmarshal(r.Body, dest)
}

// String returns the response body as a string
func (r *Response) String() string {
	return string(r.Body)
}

// IsSuccess checks if the response status is 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError checks if the response status is 4xx or 5xx
func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}
