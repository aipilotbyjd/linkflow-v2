package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// SpeechNode handles text-to-speech and speech-to-text
type SpeechNode struct {
	factory    *providers.Factory
	httpClient *http.Client
}

// NewSpeechNode creates a new speech node
func NewSpeechNode() *SpeechNode {
	return &SpeechNode{
		factory:    providers.NewFactory(),
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (n *SpeechNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	// Get credentials
	apiKey, _ := params["api_key"].(string)
	if apiKey == "" {
		if credRef, ok := params["credential"].(map[string]interface{}); ok {
			if key, ok := credRef["api_key"].(string); ok {
				apiKey = key
			}
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Get operation
	operation, _ := params["operation"].(string)
	if operation == "" {
		operation = "text_to_speech"
	}

	// Create provider adapter
	provider := ai.ProviderOpenAI
	config := &ai.ProviderConfig{
		Provider: provider,
		APIKey:   apiKey,
	}

	if orgID, ok := params["org_id"].(string); ok {
		config.OrgID = orgID
	}

	adapter, err := n.factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider adapter: %w", err)
	}

	switch operation {
	case "text_to_speech", "tts":
		return n.textToSpeech(ctx, adapter, params)
	case "speech_to_text", "stt", "transcribe":
		return n.speechToText(ctx, adapter, params)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *SpeechNode) textToSpeech(ctx context.Context, adapter ai.ProviderAdapter, params map[string]interface{}) (types.JSON, error) {
	text, _ := params["text"].(string)
	if text == "" {
		return nil, fmt.Errorf("text is required for text-to-speech")
	}

	model, _ := params["model"].(string)
	if model == "" {
		model = "tts-1"
	}

	voice, _ := params["voice"].(string)
	if voice == "" {
		voice = "alloy"
	}

	req := &ai.TTSRequest{
		Input: text,
		Model: model,
		Voice: voice,
	}

	if format, ok := params["format"].(string); ok {
		req.ResponseFormat = format
	}

	if speed, ok := params["speed"].(float64); ok {
		req.Speed = speed
	}

	resp, err := adapter.TextToSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("text-to-speech failed: %w", err)
	}

	return types.JSON{
		"audio_base64": base64.StdEncoding.EncodeToString(resp.Audio),
		"format":       resp.Format,
		"provider":     resp.Provider.String(),
		"cost_usd":     resp.CostUSD,
		"latency_ms":   resp.LatencyMS,
	}, nil
}

func (n *SpeechNode) speechToText(ctx context.Context, adapter ai.ProviderAdapter, params map[string]interface{}) (types.JSON, error) {
	// Get audio input
	var audioData []byte

	if audioBase64, ok := params["audio_base64"].(string); ok && audioBase64 != "" {
		var err error
		audioData, err = base64.StdEncoding.DecodeString(audioBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode audio base64: %w", err)
		}
	} else if audioURL, ok := params["audio_url"].(string); ok && audioURL != "" {
		data, err := n.fetchAudio(ctx, audioURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch audio: %w", err)
		}
		audioData = data
	}

	if len(audioData) == 0 {
		return nil, fmt.Errorf("audio_base64 or audio_url is required for speech-to-text")
	}

	model, _ := params["model"].(string)
	if model == "" {
		model = "whisper-1"
	}

	req := &ai.STTRequest{
		Audio: audioData,
		Model: model,
	}

	if language, ok := params["language"].(string); ok {
		req.Language = language
	}

	if prompt, ok := params["prompt"].(string); ok {
		req.Prompt = prompt
	}

	if format, ok := params["response_format"].(string); ok {
		req.ResponseFormat = format
	}

	if timestamps, ok := params["timestamps"].(bool); ok {
		req.Timestamps = timestamps
	}

	resp, err := adapter.SpeechToText(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("speech-to-text failed: %w", err)
	}

	result := types.JSON{
		"text":       resp.Text,
		"provider":   resp.Provider.String(),
		"cost_usd":   resp.CostUSD,
		"latency_ms": resp.LatencyMS,
	}

	if resp.Language != "" {
		result["language"] = resp.Language
	}
	if resp.Duration > 0 {
		result["duration"] = resp.Duration
	}
	if len(resp.Segments) > 0 {
		result["segments"] = resp.Segments
	}

	return result, nil
}

func (n *SpeechNode) fetchAudio(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch audio: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (n *SpeechNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.speech",
		Name:        "AI Speech",
		Description: "Text-to-speech and speech-to-text using OpenAI",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "Microphone01",
		Color:       "#EF4444",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "api_key",
				DisplayName: "OpenAI API Key",
				Type:        "credential",
				Required:    true,
			},
			{
				Name:        "operation",
				DisplayName: "Operation",
				Type:        "options",
				Required:    true,
				Default:     "text_to_speech",
				Options: []wtypes.ParamOption{
					{Name: "Text to Speech", Value: "text_to_speech"},
					{Name: "Speech to Text", Value: "speech_to_text"},
				},
			},
			// TTS parameters
			{
				Name:        "text",
				DisplayName: "Text",
				Type:        "string",
				Required:    false,
				Description: "Text to convert to speech",
				ShowIf:      "operation=text_to_speech",
			},
			{
				Name:        "voice",
				DisplayName: "Voice",
				Type:        "options",
				Required:    false,
				Default:     "alloy",
				ShowIf:      "operation=text_to_speech",
				Options: []wtypes.ParamOption{
					{Name: "Alloy", Value: "alloy"},
					{Name: "Echo", Value: "echo"},
					{Name: "Fable", Value: "fable"},
					{Name: "Onyx", Value: "onyx"},
					{Name: "Nova", Value: "nova"},
					{Name: "Shimmer", Value: "shimmer"},
				},
			},
			{
				Name:        "model",
				DisplayName: "TTS Model",
				Type:        "options",
				Required:    false,
				Default:     "tts-1",
				ShowIf:      "operation=text_to_speech",
				Options: []wtypes.ParamOption{
					{Name: "TTS-1 (Standard)", Value: "tts-1"},
					{Name: "TTS-1 HD (Higher Quality)", Value: "tts-1-hd"},
				},
			},
			{
				Name:        "format",
				DisplayName: "Output Format",
				Type:        "options",
				Required:    false,
				Default:     "mp3",
				ShowIf:      "operation=text_to_speech",
				Options: []wtypes.ParamOption{
					{Name: "MP3", Value: "mp3"},
					{Name: "Opus", Value: "opus"},
					{Name: "AAC", Value: "aac"},
					{Name: "FLAC", Value: "flac"},
				},
			},
			{
				Name:        "speed",
				DisplayName: "Speed",
				Type:        "number",
				Required:    false,
				Default:     1.0,
				Description: "Speed (0.25 to 4.0)",
				ShowIf:      "operation=text_to_speech",
			},
			// STT parameters
			{
				Name:        "audio_base64",
				DisplayName: "Audio Base64",
				Type:        "string",
				Required:    false,
				Description: "Base64-encoded audio data",
				ShowIf:      "operation=speech_to_text",
			},
			{
				Name:        "audio_url",
				DisplayName: "Audio URL",
				Type:        "string",
				Required:    false,
				Description: "URL of the audio file",
				ShowIf:      "operation=speech_to_text",
			},
			{
				Name:        "language",
				DisplayName: "Language",
				Type:        "string",
				Required:    false,
				Description: "Language code (e.g., en, es, fr)",
				ShowIf:      "operation=speech_to_text",
			},
			{
				Name:        "timestamps",
				DisplayName: "Include Timestamps",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				ShowIf:      "operation=speech_to_text",
			},
		},
		Credentials: []string{"openai"},
	}
}
