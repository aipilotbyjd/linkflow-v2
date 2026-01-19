package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
)

// Gateway is the unified AI gateway that handles all AI requests
type Gateway struct {
	providers     map[ai.Provider]ai.ProviderAdapter
	router        *Router
	fallback      *FallbackChain
	rateLimiter   *RateLimiter
	cache         ai.Cache
	usageRepo     ai.UsageRepository
	logger        logger.Logger
	mu            sync.RWMutex
}

// GatewayConfig holds gateway configuration
type GatewayConfig struct {
	EnableCache      bool
	EnableFallback   bool
	EnableRateLimit  bool
	DefaultTimeout   time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
}

// DefaultGatewayConfig returns default gateway configuration
func DefaultGatewayConfig() GatewayConfig {
	return GatewayConfig{
		EnableCache:     true,
		EnableFallback:  true,
		EnableRateLimit: true,
		DefaultTimeout:  30 * time.Second,
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
	}
}

// NewGateway creates a new AI gateway
func NewGateway(config GatewayConfig, cache ai.Cache, usageRepo ai.UsageRepository, log logger.Logger) *Gateway {
	return &Gateway{
		providers:   make(map[ai.Provider]ai.ProviderAdapter),
		router:      NewRouter(),
		fallback:    NewFallbackChain(config.MaxRetries, config.RetryDelay),
		rateLimiter: NewRateLimiter(),
		cache:       cache,
		usageRepo:   usageRepo,
		logger:      log,
	}
}

// RegisterProvider registers a provider adapter
func (g *Gateway) RegisterProvider(provider ai.Provider, adapter ai.ProviderAdapter) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.providers[provider] = adapter
}

// GetProvider returns a provider adapter
func (g *Gateway) GetProvider(provider ai.Provider) (ai.ProviderAdapter, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	adapter, ok := g.providers[provider]
	if !ok {
		return nil, ai.ErrProviderNotFound
	}
	return adapter, nil
}

// Chat sends a chat completion request
func (g *Gateway) Chat(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error) {
	startTime := time.Now()

	// Determine provider from model
	model, ok := ai.GetModel(req.Model)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ai.ErrModelNotFound, req.Model)
	}

	provider := model.Provider
	if req.ProviderConfig != nil && req.ProviderConfig.Provider != "" {
		provider = req.ProviderConfig.Provider
	}

	// Check cache first
	if g.cache != nil {
		cacheKey := g.generateCacheKey(req.Messages, req.Model, req.Temperature)
		if cached, err := g.cache.Get(ctx, cacheKey, req.Model); err == nil && cached != nil {
			var resp ai.ChatResponse
			if err := json.Unmarshal(cached.Response, &resp); err == nil {
				resp.Cached = true
				resp.LatencyMS = time.Since(startTime).Milliseconds()
				return &resp, nil
			}
		}
	}

	// Check rate limit
	if err := g.rateLimiter.Allow(ctx, req.WorkspaceID, provider); err != nil {
		return nil, err
	}

	// Get provider adapter
	adapter, err := g.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	// Execute request with fallback
	var resp *ai.ChatResponse
	resp, err = g.fallback.Execute(ctx, func(ctx context.Context) (*ai.ChatResponse, error) {
		return adapter.Chat(ctx, req)
	}, func() []ai.ProviderAdapter {
		return g.getFallbackProviders(provider)
	})

	if err != nil {
		return nil, err
	}

	resp.LatencyMS = time.Since(startTime).Milliseconds()

	// Cache response
	if g.cache != nil && !req.Stream {
		g.cacheResponse(ctx, req, resp)
	}

	// Track usage
	g.trackUsage(ctx, req.WorkspaceID, req.ExecutionID, provider, req.Model, "chat", &resp.Usage, resp.CostUSD, resp.LatencyMS)

	return resp, nil
}

// Complete sends a text completion request
func (g *Gateway) Complete(ctx context.Context, req *ai.CompletionRequest) (*ai.CompletionResponse, error) {
	startTime := time.Now()

	model, ok := ai.GetModel(req.Model)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ai.ErrModelNotFound, req.Model)
	}

	provider := model.Provider
	if req.ProviderConfig != nil && req.ProviderConfig.Provider != "" {
		provider = req.ProviderConfig.Provider
	}

	adapter, err := g.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	resp, err := adapter.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	resp.LatencyMS = time.Since(startTime).Milliseconds()

	g.trackUsage(ctx, req.WorkspaceID, "", provider, req.Model, "completion", &resp.Usage, resp.CostUSD, resp.LatencyMS)

	return resp, nil
}

// Embed generates embeddings
func (g *Gateway) Embed(ctx context.Context, req *ai.EmbeddingRequest) (*ai.EmbeddingResponse, error) {
	startTime := time.Now()

	provider := ai.ProviderOpenAI // Default to OpenAI for embeddings
	if req.ProviderConfig != nil && req.ProviderConfig.Provider != "" {
		provider = req.ProviderConfig.Provider
	}

	adapter, err := g.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	resp, err := adapter.Embed(ctx, req)
	if err != nil {
		return nil, err
	}

	resp.LatencyMS = time.Since(startTime).Milliseconds()

	g.trackUsage(ctx, req.WorkspaceID, "", provider, req.Model, "embedding", &ai.Usage{TotalTokens: resp.Usage.TotalTokens}, resp.CostUSD, resp.LatencyMS)

	return resp, nil
}

// GenerateImage generates images
func (g *Gateway) GenerateImage(ctx context.Context, req *ai.ImageRequest) (*ai.ImageResponse, error) {
	startTime := time.Now()

	provider := ai.ProviderOpenAI // Default to OpenAI for images
	if req.ProviderConfig != nil && req.ProviderConfig.Provider != "" {
		provider = req.ProviderConfig.Provider
	}

	adapter, err := g.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	resp, err := adapter.GenerateImage(ctx, req)
	if err != nil {
		return nil, err
	}

	resp.LatencyMS = time.Since(startTime).Milliseconds()

	g.trackUsage(ctx, req.WorkspaceID, "", provider, req.Model, "image", nil, resp.CostUSD, resp.LatencyMS)

	return resp, nil
}

// AnalyzeImage analyzes images using vision models
func (g *Gateway) AnalyzeImage(ctx context.Context, req *ai.VisionRequest) (*ai.VisionResponse, error) {
	startTime := time.Now()

	model, ok := ai.GetModel(req.Model)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ai.ErrModelNotFound, req.Model)
	}

	if !model.SupportsVision {
		return nil, fmt.Errorf("%w: %s", ai.ErrVisionNotSupported, req.Model)
	}

	provider := model.Provider
	if req.ProviderConfig != nil && req.ProviderConfig.Provider != "" {
		provider = req.ProviderConfig.Provider
	}

	adapter, err := g.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	resp, err := adapter.AnalyzeImage(ctx, req)
	if err != nil {
		return nil, err
	}

	resp.LatencyMS = time.Since(startTime).Milliseconds()

	g.trackUsage(ctx, req.WorkspaceID, "", provider, req.Model, "vision", &resp.Usage, resp.CostUSD, resp.LatencyMS)

	return resp, nil
}

// TextToSpeech converts text to speech
func (g *Gateway) TextToSpeech(ctx context.Context, req *ai.TTSRequest) (*ai.TTSResponse, error) {
	startTime := time.Now()

	provider := ai.ProviderOpenAI // Default to OpenAI for TTS
	if req.ProviderConfig != nil && req.ProviderConfig.Provider != "" {
		provider = req.ProviderConfig.Provider
	}

	adapter, err := g.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	resp, err := adapter.TextToSpeech(ctx, req)
	if err != nil {
		return nil, err
	}

	resp.LatencyMS = time.Since(startTime).Milliseconds()

	g.trackUsage(ctx, req.WorkspaceID, "", provider, req.Model, "tts", nil, resp.CostUSD, resp.LatencyMS)

	return resp, nil
}

// SpeechToText converts speech to text
func (g *Gateway) SpeechToText(ctx context.Context, req *ai.STTRequest) (*ai.STTResponse, error) {
	startTime := time.Now()

	provider := ai.ProviderOpenAI // Default to OpenAI for STT
	if req.ProviderConfig != nil && req.ProviderConfig.Provider != "" {
		provider = req.ProviderConfig.Provider
	}

	adapter, err := g.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	resp, err := adapter.SpeechToText(ctx, req)
	if err != nil {
		return nil, err
	}

	resp.LatencyMS = time.Since(startTime).Milliseconds()

	g.trackUsage(ctx, req.WorkspaceID, "", provider, req.Model, "stt", nil, resp.CostUSD, resp.LatencyMS)

	return resp, nil
}

// Route determines the best model for a request
func (g *Gateway) Route(ctx context.Context, req *ai.RouterRequest) (*ai.RouterResponse, error) {
	return g.router.Route(ctx, req, g.providers)
}

func (g *Gateway) generateCacheKey(messages []ai.Message, model string, temperature *float64) string {
	data := struct {
		Messages    []ai.Message `json:"messages"`
		Model       string       `json:"model"`
		Temperature *float64     `json:"temperature,omitempty"`
	}{
		Messages:    messages,
		Model:       model,
		Temperature: temperature,
	}

	b, _ := json.Marshal(data)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

func (g *Gateway) cacheResponse(ctx context.Context, req *ai.ChatRequest, resp *ai.ChatResponse) {
	cacheKey := g.generateCacheKey(req.Messages, req.Model, req.Temperature)

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return
	}

	var workspaceID uuid.UUID
	if req.WorkspaceID != "" {
		workspaceID, _ = uuid.Parse(req.WorkspaceID)
	}

	entry := &ai.CacheEntry{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		PromptHash:  cacheKey,
		Model:       req.Model,
		Provider:    resp.Provider,
		Response:    respBytes,
		RequestType: "chat",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	if err := g.cache.Set(ctx, entry); err != nil {
		g.logger.Error().Err(err).Msg("failed to cache response")
	}
}

func (g *Gateway) trackUsage(ctx context.Context, workspaceID, executionID string, provider ai.Provider, model, requestType string, usage *ai.Usage, cost float64, latencyMS int64) {
	if g.usageRepo == nil {
		return
	}

	var wsID, execID uuid.UUID
	if workspaceID != "" {
		wsID, _ = uuid.Parse(workspaceID)
	}
	if executionID != "" {
		execID, _ = uuid.Parse(executionID)
	}

	record := ai.NewUsageRecord(wsID, provider, model, requestType)
	record.ExecutionID = execID
	record.LatencyMS = latencyMS
	record.CostUSD = cost

	if usage != nil {
		record.SetUsage(usage.InputTokens, usage.OutputTokens)
		if cost == 0 {
			record.CalculateCost()
		}
	}

	if err := g.usageRepo.Create(record); err != nil {
		g.logger.Error().Err(err).Msg("failed to track usage")
	}
}

func (g *Gateway) getFallbackProviders(primaryProvider ai.Provider) []ai.ProviderAdapter {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var fallbacks []ai.ProviderAdapter
	for p, adapter := range g.providers {
		if p != primaryProvider {
			fallbacks = append(fallbacks, adapter)
		}
	}
	return fallbacks
}

// ListProviders returns all registered providers
func (g *Gateway) ListProviders() []ai.Provider {
	g.mu.RLock()
	defer g.mu.RUnlock()

	providers := make([]ai.Provider, 0, len(g.providers))
	for p := range g.providers {
		providers = append(providers, p)
	}
	return providers
}

// ListModels returns all available models
func (g *Gateway) ListModels(ctx context.Context) []ai.Model {
	models := make([]ai.Model, 0, len(ai.ModelRegistry))
	for _, m := range ai.ModelRegistry {
		models = append(models, m)
	}
	return models
}
