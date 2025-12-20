package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"gorm.io/gorm"
)

type Manager struct {
	db        *gorm.DB
	endpoints sync.Map
	handlers  map[string]WebhookHandler
	mu        sync.RWMutex
}

type WebhookHandler func(ctx context.Context, req *WebhookRequest) (*WebhookResponse, error)

type WebhookRequest struct {
	ID         string
	Method     string
	Path       string
	Headers    http.Header
	Query      map[string][]string
	Body       []byte
	RemoteAddr string
	ReceivedAt time.Time
}

type WebhookResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       interface{}
}

type EndpointConfig struct {
	ID                string
	WorkflowID        uuid.UUID
	WorkspaceID       uuid.UUID
	Path              string
	Methods           []string
	AuthType          string
	AuthValue         string
	HMACSecret        string
	HMACHeader        string
	IPWhitelist       []string
	ResponseMode      string
	ResponseTimeout   time.Duration
	IsActive          bool
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:       db,
		handlers: make(map[string]WebhookHandler),
	}
}

func (m *Manager) RegisterEndpoint(config *EndpointConfig) error {
	if config.Path == "" {
		config.Path = fmt.Sprintf("/webhook/%s", uuid.New().String())
	}

	var secret *string
	if config.HMACSecret != "" {
		secret = &config.HMACSecret
	}

	method := "POST"
	if len(config.Methods) > 0 {
		method = config.Methods[0]
	}

	endpoint := &models.WebhookEndpoint{
		WorkflowID:  config.WorkflowID,
		WorkspaceID: config.WorkspaceID,
		Path:        config.Path,
		Method:      method,
		Secret:      secret,
		IsActive:    config.IsActive,
	}

	if err := m.db.Create(endpoint).Error; err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	config.ID = endpoint.ID.String()
	m.endpoints.Store(config.Path, config)

	return nil
}

func (m *Manager) UnregisterEndpoint(path string) error {
	m.endpoints.Delete(path)
	return m.db.Where("path = ?", path).Delete(&models.WebhookEndpoint{}).Error
}

func (m *Manager) GetEndpoint(path string) (*EndpointConfig, bool) {
	if v, ok := m.endpoints.Load(path); ok {
		return v.(*EndpointConfig), true
	}

	var endpoint models.WebhookEndpoint
	if err := m.db.Where("path = ? AND is_active = ?", path, true).First(&endpoint).Error; err != nil {
		return nil, false
	}

	hmacSecret := ""
	if endpoint.Secret != nil {
		hmacSecret = *endpoint.Secret
	}
	config := &EndpointConfig{
		ID:          endpoint.ID.String(),
		WorkflowID:  endpoint.WorkflowID,
		WorkspaceID: endpoint.WorkspaceID,
		Path:        endpoint.Path,
		Methods:     []string{endpoint.Method},
		HMACSecret:  hmacSecret,
		IsActive:    endpoint.IsActive,
	}
	m.endpoints.Store(path, config)

	return config, true
}

func (m *Manager) HandleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	endpoint, ok := m.GetEndpoint(path)
	if !ok {
		http.Error(w, "Webhook endpoint not found", http.StatusNotFound)
		return
	}

	if !endpoint.IsActive {
		http.Error(w, "Webhook endpoint is disabled", http.StatusServiceUnavailable)
		return
	}

	if !m.isMethodAllowed(r.Method, endpoint.Methods) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := m.authenticate(r, endpoint); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if !m.checkIPWhitelist(r.RemoteAddr, endpoint.IPWhitelist) {
		http.Error(w, "IP not allowed", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	webhookReq := &WebhookRequest{
		ID:         uuid.New().String(),
		Method:     r.Method,
		Path:       path,
		Headers:    r.Header,
		Query:      r.URL.Query(),
		Body:       body,
		RemoteAddr: r.RemoteAddr,
		ReceivedAt: time.Now(),
	}

	m.logRequest(endpoint, webhookReq)

	ctx := context.Background()
	if endpoint.ResponseTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, endpoint.ResponseTimeout)
		defer cancel()
	}

	m.mu.RLock()
	handler, ok := m.handlers[endpoint.WorkflowID.String()]
	m.mu.RUnlock()

	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message":    "Webhook received",
			"webhookId":  webhookReq.ID,
			"receivedAt": webhookReq.ReceivedAt,
		})
		return
	}

	response, err := handler(ctx, webhookReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range response.Headers {
		w.Header().Set(k, v)
	}

	if response.Headers["Content-Type"] == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(response.StatusCode)
	if response.Body != nil {
		json.NewEncoder(w).Encode(response.Body)
	}
}

func (m *Manager) RegisterHandler(workflowID string, handler WebhookHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[workflowID] = handler
}

func (m *Manager) UnregisterHandler(workflowID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.handlers, workflowID)
}

func (m *Manager) isMethodAllowed(method string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, m := range allowed {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func (m *Manager) authenticate(r *http.Request, endpoint *EndpointConfig) error {
	switch endpoint.AuthType {
	case "none", "":
		return nil
	case "header":
		return m.authenticateHeader(r, endpoint)
	case "basicAuth":
		return m.authenticateBasic(r, endpoint)
	case "hmac":
		return m.authenticateHMAC(r, endpoint)
	case "apiKey":
		return m.authenticateAPIKey(r, endpoint)
	default:
		return nil
	}
}

func (m *Manager) authenticateHeader(r *http.Request, endpoint *EndpointConfig) error {
	headerName := "X-Webhook-Token"
	if endpoint.HMACHeader != "" {
		headerName = endpoint.HMACHeader
	}
	token := r.Header.Get(headerName)
	if token != endpoint.AuthValue {
		return fmt.Errorf("invalid webhook token")
	}
	return nil
}

func (m *Manager) authenticateBasic(r *http.Request, endpoint *EndpointConfig) error {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return fmt.Errorf("basic auth required")
	}
	expected := fmt.Sprintf("%s:%s", user, pass)
	if expected != endpoint.AuthValue {
		return fmt.Errorf("invalid credentials")
	}
	return nil
}

func (m *Manager) authenticateHMAC(r *http.Request, endpoint *EndpointConfig) error {
	signatureHeader := endpoint.HMACHeader
	if signatureHeader == "" {
		signatureHeader = "X-Hub-Signature-256"
	}

	signature := r.Header.Get(signatureHeader)
	if signature == "" {
		return fmt.Errorf("missing signature header")
	}

	signature = strings.TrimPrefix(signature, "sha256=")

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	mac := hmac.New(sha256.New, []byte(endpoint.HMACSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (m *Manager) authenticateAPIKey(r *http.Request, endpoint *EndpointConfig) error {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.URL.Query().Get("api_key")
	}
	if apiKey != endpoint.AuthValue {
		return fmt.Errorf("invalid API key")
	}
	return nil
}

func (m *Manager) checkIPWhitelist(remoteAddr string, whitelist []string) bool {
	if len(whitelist) == 0 {
		return true
	}

	ip := strings.Split(remoteAddr, ":")[0]
	for _, allowed := range whitelist {
		if ip == allowed || allowed == "*" {
			return true
		}
	}
	return false
}

func (m *Manager) logRequest(endpoint *EndpointConfig, req *WebhookRequest) {
	var endpointID uuid.UUID
	endpointID, _ = uuid.Parse(endpoint.ID)
	body := string(req.Body)

	log := &models.WebhookLog{
		EndpointID: endpointID,
		Path:       req.Path,
		Method:     req.Method,
		StatusCode: 0,
		Body:       &body,
		DurationMs: 0,
	}
	m.db.Create(log)
}

func (m *Manager) LoadActiveEndpoints() error {
	var endpoints []models.WebhookEndpoint
	if err := m.db.Where("is_active = ?", true).Find(&endpoints).Error; err != nil {
		return err
	}

	for _, e := range endpoints {
		hmacSecret := ""
		if e.Secret != nil {
			hmacSecret = *e.Secret
		}
		config := &EndpointConfig{
			ID:          e.ID.String(),
			WorkflowID:  e.WorkflowID,
			WorkspaceID: e.WorkspaceID,
			Path:        e.Path,
			Methods:     []string{e.Method},
			HMACSecret:  hmacSecret,
			IsActive:    e.IsActive,
		}
		m.endpoints.Store(e.Path, config)
	}

	return nil
}

func (m *Manager) ActivateWorkflowWebhooks(workflowID uuid.UUID) error {
	return m.db.Model(&models.WebhookEndpoint{}).
		Where("workflow_id = ?", workflowID).
		Update("is_active", true).Error
}

func (m *Manager) DeactivateWorkflowWebhooks(workflowID uuid.UUID) error {
	var endpoints []models.WebhookEndpoint
	m.db.Where("workflow_id = ?", workflowID).Find(&endpoints)

	for _, e := range endpoints {
		m.endpoints.Delete(e.Path)
	}

	return m.db.Model(&models.WebhookEndpoint{}).
		Where("workflow_id = ?", workflowID).
		Update("is_active", false).Error
}
