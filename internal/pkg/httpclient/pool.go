package httpclient

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/linkflow-ai/linkflow/internal/pkg/circuitbreaker"
)

var (
	defaultClient     *PooledClient
	defaultClientOnce sync.Once
)

type Config struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
	TLSHandshakeTimeout time.Duration
	ResponseTimeout     time.Duration
	KeepAlive           time.Duration
	DisableKeepAlives   bool
	InsecureSkipVerify  bool
}

func DefaultConfig() Config {
	return Config{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		ResponseTimeout:     30 * time.Second,
		KeepAlive:           30 * time.Second,
		DisableKeepAlives:   false,
		InsecureSkipVerify:  false,
	}
}

type PooledClient struct {
	client         *http.Client
	circuitBreaker *circuitbreaker.Manager
	config         Config
}

func NewPooledClient(config Config) *PooledClient {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		TLSHandshakeTimeout: config.TLSHandshakeTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.ResponseTimeout,
	}

	cbConfig := circuitbreaker.Config{
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 2,
	}

	return &PooledClient{
		client:         client,
		circuitBreaker: circuitbreaker.NewManager(cbConfig),
		config:         config,
	}
}

func Default() *PooledClient {
	defaultClientOnce.Do(func() {
		defaultClient = NewPooledClient(DefaultConfig())
	})
	return defaultClient
}

func (p *PooledClient) Do(req *http.Request) (*http.Response, error) {
	host := req.URL.Host
	cb := p.circuitBreaker.Get(host)

	result, err := cb.ExecuteWithContext(req.Context(), func(ctx context.Context) (interface{}, error) {
		return p.client.Do(req.WithContext(ctx))
	})

	if err != nil {
		return nil, err
	}
	return result.(*http.Response), nil
}

func (p *PooledClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return p.Do(req)
}

func (p *PooledClient) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return p.Do(req)
}

func (p *PooledClient) PostJSON(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	return p.Post(ctx, url, "application/json", body)
}

func (p *PooledClient) Put(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return p.Do(req)
}

func (p *PooledClient) Delete(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	return p.Do(req)
}

func (p *PooledClient) CircuitStates() map[string]circuitbreaker.State {
	return p.circuitBreaker.States()
}

func (p *PooledClient) CloseIdleConnections() {
	p.client.CloseIdleConnections()
}

// Request builder for more complex requests
type Request struct {
	client  *PooledClient
	method  string
	url     string
	headers map[string]string
	body    io.Reader
	timeout time.Duration
}

func (p *PooledClient) NewRequest(method, url string) *Request {
	return &Request{
		client:  p,
		method:  method,
		url:     url,
		headers: make(map[string]string),
	}
}

func (r *Request) Header(key, value string) *Request {
	r.headers[key] = value
	return r
}

func (r *Request) Headers(headers map[string]string) *Request {
	for k, v := range headers {
		r.headers[k] = v
	}
	return r
}

func (r *Request) Body(body io.Reader) *Request {
	r.body = body
	return r
}

func (r *Request) Timeout(timeout time.Duration) *Request {
	r.timeout = timeout
	return r
}

func (r *Request) Do(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, r.method, r.url, r.body)
	if err != nil {
		return nil, err
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	return r.client.Do(req)
}
