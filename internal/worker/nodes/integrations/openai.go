package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type OpenAINode struct{}

func (n *OpenAINode) Type() string {
	return "integration.openai"
}

func (n *OpenAINode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "chat")

	// Get credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	apiKey := cred.APIKey

	switch operation {
	case "chat":
		return n.chatCompletion(ctx, apiKey, config)
	case "completion":
		return n.textCompletion(ctx, apiKey, config)
	case "embedding":
		return n.createEmbedding(ctx, apiKey, config)
	case "image":
		return n.generateImage(ctx, apiKey, config)
	case "imageEdit":
		return n.editImage(ctx, apiKey, config)
	case "imageVariation":
		return n.createImageVariation(ctx, apiKey, config)
	case "transcription":
		return n.transcribeAudio(ctx, apiKey, config)
	case "translation":
		return n.translateAudio(ctx, apiKey, config)
	case "moderation":
		return n.moderateContent(ctx, apiKey, config)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *OpenAINode) chatCompletion(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	model := getString(config, "model", "gpt-4")
	messages := config["messages"]
	temperature := getFloat(config, "temperature", 0.7)
	maxTokens := getInt(config, "maxTokens", 1000)
	topP := getFloat(config, "topP", 1.0)
	frequencyPenalty := getFloat(config, "frequencyPenalty", 0)
	presencePenalty := getFloat(config, "presencePenalty", 0)
	systemMessage := getString(config, "systemMessage", "")
	userMessage := getString(config, "userMessage", "")

	// Build messages array
	var msgs []map[string]string
	if messages != nil {
		if msgArr, ok := messages.([]interface{}); ok {
			for _, m := range msgArr {
				if msg, ok := m.(map[string]interface{}); ok {
					msgs = append(msgs, map[string]string{
						"role":    getString(msg, "role", "user"),
						"content": getString(msg, "content", ""),
					})
				}
			}
		}
	} else {
		if systemMessage != "" {
			msgs = append(msgs, map[string]string{"role": "system", "content": systemMessage})
		}
		if userMessage != "" {
			msgs = append(msgs, map[string]string{"role": "user", "content": userMessage})
		}
	}

	payload := map[string]interface{}{
		"model":             model,
		"messages":          msgs,
		"temperature":       temperature,
		"max_tokens":        maxTokens,
		"top_p":             topP,
		"frequency_penalty": frequencyPenalty,
		"presence_penalty":  presencePenalty,
	}

	// Add functions if provided
	if functions := config["functions"]; functions != nil {
		payload["functions"] = functions
	}

	// Add tools if provided
	if tools := config["tools"]; tools != nil {
		payload["tools"] = tools
	}

	result, err := n.makeRequest(ctx, apiKey, "POST", "https://api.openai.com/v1/chat/completions", payload)
	if err != nil {
		return nil, err
	}

	// Extract the message content for convenience
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				result["content"] = message["content"]
				result["role"] = message["role"]
				if fc := message["function_call"]; fc != nil {
					result["function_call"] = fc
				}
				if tc := message["tool_calls"]; tc != nil {
					result["tool_calls"] = tc
				}
			}
		}
	}

	return result, nil
}

func (n *OpenAINode) textCompletion(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	model := getString(config, "model", "gpt-3.5-turbo-instruct")
	prompt := getString(config, "prompt", "")
	temperature := getFloat(config, "temperature", 0.7)
	maxTokens := getInt(config, "maxTokens", 1000)

	payload := map[string]interface{}{
		"model":       model,
		"prompt":      prompt,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}

	result, err := n.makeRequest(ctx, apiKey, "POST", "https://api.openai.com/v1/completions", payload)
	if err != nil {
		return nil, err
	}

	// Extract text for convenience
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			result["text"] = choice["text"]
		}
	}

	return result, nil
}

func (n *OpenAINode) createEmbedding(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	model := getString(config, "model", "text-embedding-ada-002")
	input := config["input"]

	payload := map[string]interface{}{
		"model": model,
		"input": input,
	}

	result, err := n.makeRequest(ctx, apiKey, "POST", "https://api.openai.com/v1/embeddings", payload)
	if err != nil {
		return nil, err
	}

	// Extract embedding for convenience
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			result["embedding"] = item["embedding"]
		}
	}

	return result, nil
}

func (n *OpenAINode) generateImage(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	model := getString(config, "model", "dall-e-3")
	prompt := getString(config, "prompt", "")
	size := getString(config, "size", "1024x1024")
	quality := getString(config, "quality", "standard")
	imageCount := getInt(config, "n", 1)
	responseFormat := getString(config, "responseFormat", "url")

	payload := map[string]interface{}{
		"model":           model,
		"prompt":          prompt,
		"size":            size,
		"quality":         quality,
		"n":               imageCount,
		"response_format": responseFormat,
	}

	result, err := n.makeRequest(ctx, apiKey, "POST", "https://api.openai.com/v1/images/generations", payload)
	if err != nil {
		return nil, err
	}

	// Extract URL for convenience
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			if url := item["url"]; url != nil {
				result["url"] = url
			}
			if b64 := item["b64_json"]; b64 != nil {
				result["b64_json"] = b64
			}
		}
	}

	return result, nil
}

func (n *OpenAINode) editImage(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	prompt := getString(config, "prompt", "")
	size := getString(config, "size", "1024x1024")
	imageCount := getInt(config, "n", 1)
	responseFormat := getString(config, "responseFormat", "url")

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add image file
	if err := n.addFileToMultipart(writer, "image", config); err != nil {
		return nil, fmt.Errorf("failed to add image: %w", err)
	}

	// Add mask if provided
	if mask := config["mask"]; mask != nil {
		if maskConfig, ok := mask.(map[string]interface{}); ok {
			if err := n.addFileToMultipartFromConfig(writer, "mask", maskConfig); err != nil {
				return nil, fmt.Errorf("failed to add mask: %w", err)
			}
		}
	}

	_ = writer.WriteField("prompt", prompt)
	_ = writer.WriteField("size", size)
	_ = writer.WriteField("n", fmt.Sprintf("%d", imageCount))
	_ = writer.WriteField("response_format", responseFormat)
	_ = writer.Close()

	return n.makeMultipartRequest(ctx, apiKey, "https://api.openai.com/v1/images/edits", &buf, writer.FormDataContentType())
}

func (n *OpenAINode) createImageVariation(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	size := getString(config, "size", "1024x1024")
	imageCount := getInt(config, "n", 1)
	responseFormat := getString(config, "responseFormat", "url")

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add image file
	if err := n.addFileToMultipart(writer, "image", config); err != nil {
		return nil, fmt.Errorf("failed to add image: %w", err)
	}

	_ = writer.WriteField("size", size)
	_ = writer.WriteField("n", fmt.Sprintf("%d", imageCount))
	_ = writer.WriteField("response_format", responseFormat)
	_ = writer.Close()

	return n.makeMultipartRequest(ctx, apiKey, "https://api.openai.com/v1/images/variations", &buf, writer.FormDataContentType())
}

func (n *OpenAINode) transcribeAudio(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	model := getString(config, "model", "whisper-1")
	language := getString(config, "language", "")
	prompt := getString(config, "prompt", "")
	responseFormat := getString(config, "responseFormat", "json")
	temperature := getFloat(config, "temperature", 0)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add audio file
	if err := n.addFileToMultipart(writer, "file", config); err != nil {
		return nil, fmt.Errorf("failed to add audio file: %w", err)
	}

	_ = writer.WriteField("model", model)
	if language != "" {
		_ = writer.WriteField("language", language)
	}
	if prompt != "" {
		_ = writer.WriteField("prompt", prompt)
	}
	_ = writer.WriteField("response_format", responseFormat)
	_ = writer.WriteField("temperature", fmt.Sprintf("%f", temperature))
	_ = writer.Close()

	return n.makeMultipartRequest(ctx, apiKey, "https://api.openai.com/v1/audio/transcriptions", &buf, writer.FormDataContentType())
}

func (n *OpenAINode) translateAudio(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	model := getString(config, "model", "whisper-1")
	prompt := getString(config, "prompt", "")
	responseFormat := getString(config, "responseFormat", "json")
	temperature := getFloat(config, "temperature", 0)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add audio file
	if err := n.addFileToMultipart(writer, "file", config); err != nil {
		return nil, fmt.Errorf("failed to add audio file: %w", err)
	}

	_ = writer.WriteField("model", model)
	if prompt != "" {
		_ = writer.WriteField("prompt", prompt)
	}
	_ = writer.WriteField("response_format", responseFormat)
	_ = writer.WriteField("temperature", fmt.Sprintf("%f", temperature))
	_ = writer.Close()

	return n.makeMultipartRequest(ctx, apiKey, "https://api.openai.com/v1/audio/translations", &buf, writer.FormDataContentType())
}

func (n *OpenAINode) addFileToMultipart(writer *multipart.Writer, fieldName string, config map[string]interface{}) error {
	// Look for file data in config
	if fileData, ok := config["file"].(map[string]interface{}); ok {
		return n.addFileToMultipartFromConfig(writer, fieldName, fileData)
	}
	if imageData, ok := config["image"].(map[string]interface{}); ok {
		return n.addFileToMultipartFromConfig(writer, fieldName, imageData)
	}
	return fmt.Errorf("no file data provided")
}

func (n *OpenAINode) addFileToMultipartFromConfig(writer *multipart.Writer, fieldName string, fileConfig map[string]interface{}) error {
	filename := getString(fileConfig, "filename", "file")
	
	// Get file content
	var content []byte
	var err error
	
	if path, ok := fileConfig["path"].(string); ok {
		content, err = os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if filename == "file" {
			filename = filepath.Base(path)
		}
	} else if b64, ok := fileConfig["content"].(string); ok {
		content, err = base64.StdEncoding.DecodeString(b64)
		if err != nil {
			content = []byte(b64)
		}
	} else if raw, ok := fileConfig["data"].([]byte); ok {
		content = raw
	} else {
		return fmt.Errorf("no file content provided (use 'path', 'content', or 'data')")
	}
	
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return err
	}
	_, err = part.Write(content)
	return err
}

func (n *OpenAINode) makeMultipartRequest(ctx context.Context, apiKey, url string, body *bytes.Buffer, contentType string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	_ = json.Unmarshal(respBody, &result)

	if resp.StatusCode >= 400 {
		if errObj, ok := result["error"].(map[string]interface{}); ok {
			return result, fmt.Errorf("OpenAI API error: %s", errObj["message"])
		}
		return result, fmt.Errorf("OpenAI API error: %d", resp.StatusCode)
	}

	// Extract data for convenience
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			if url := item["url"]; url != nil {
				result["url"] = url
			}
			if b64 := item["b64_json"]; b64 != nil {
				result["b64_json"] = b64
			}
		}
	}
	
	// Extract text for transcription/translation
	if text, ok := result["text"].(string); ok {
		result["transcript"] = text
	}

	return result, nil
}

func (n *OpenAINode) moderateContent(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	input := getString(config, "input", "")

	payload := map[string]interface{}{
		"input": input,
	}

	result, err := n.makeRequest(ctx, apiKey, "POST", "https://api.openai.com/v1/moderations", payload)
	if err != nil {
		return nil, err
	}

	// Extract results for convenience
	if results, ok := result["results"].([]interface{}); ok && len(results) > 0 {
		if item, ok := results[0].(map[string]interface{}); ok {
			result["flagged"] = item["flagged"]
			result["categories"] = item["categories"]
			result["category_scores"] = item["category_scores"]
		}
	}

	return result, nil
}

func (n *OpenAINode) makeRequest(ctx context.Context, apiKey, method, url string, payload map[string]interface{}) (map[string]interface{}, error) {
	jsonBody, _ := json.Marshal(payload)
	body := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	_ = json.Unmarshal(respBody, &result)

	if resp.StatusCode >= 400 {
		if errObj, ok := result["error"].(map[string]interface{}); ok {
			return result, fmt.Errorf("OpenAI API error: %s", errObj["message"])
		}
		return result, fmt.Errorf("OpenAI API error: %d", resp.StatusCode)
	}

	return result, nil
}

var _ core.Node = (*OpenAINode)(nil)
