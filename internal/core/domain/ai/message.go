package ai

import (
	"encoding/json"
)

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

func (r Role) String() string {
	return string(r)
}

// ContentType represents the type of content in a message
type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeImage    ContentType = "image"
	ContentTypeAudio    ContentType = "audio"
	ContentTypeDocument ContentType = "document"
	ContentTypeToolCall ContentType = "tool_call"
	ContentTypeToolResult ContentType = "tool_result"
)

// Content represents a piece of content in a message
type Content struct {
	Type ContentType `json:"type"`

	// For text content
	Text string `json:"text,omitempty"`

	// For image content
	ImageURL    string `json:"image_url,omitempty"`
	ImageBase64 string `json:"image_base64,omitempty"`
	ImageType   string `json:"image_type,omitempty"` // jpeg, png, gif, webp

	// For audio content
	AudioURL    string `json:"audio_url,omitempty"`
	AudioBase64 string `json:"audio_base64,omitempty"`
	AudioFormat string `json:"audio_format,omitempty"` // mp3, wav, ogg

	// For document content
	DocumentURL  string `json:"document_url,omitempty"`
	DocumentData []byte `json:"document_data,omitempty"`
	DocumentType string `json:"document_type,omitempty"` // pdf, txt, etc.

	// For tool call content
	ToolCallID string `json:"tool_call_id,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	ToolInput  string `json:"tool_input,omitempty"`

	// For tool result content
	ToolResult string `json:"tool_result,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    Role      `json:"role"`
	Content []Content `json:"content"`

	// Name for user identification (optional)
	Name string `json:"name,omitempty"`

	// Tool calls made by the assistant
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// Tool call ID for tool result messages
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// NewTextMessage creates a new text message
func NewTextMessage(role Role, text string) Message {
	return Message{
		Role: role,
		Content: []Content{
			{Type: ContentTypeText, Text: text},
		},
	}
}

// NewSystemMessage creates a new system message
func NewSystemMessage(text string) Message {
	return NewTextMessage(RoleSystem, text)
}

// NewUserMessage creates a new user message
func NewUserMessage(text string) Message {
	return NewTextMessage(RoleUser, text)
}

// NewAssistantMessage creates a new assistant message
func NewAssistantMessage(text string) Message {
	return NewTextMessage(RoleAssistant, text)
}

// NewImageMessage creates a new message with an image
func NewImageMessage(role Role, text string, imageURL string) Message {
	content := []Content{}
	if text != "" {
		content = append(content, Content{Type: ContentTypeText, Text: text})
	}
	content = append(content, Content{Type: ContentTypeImage, ImageURL: imageURL})
	return Message{
		Role:    role,
		Content: content,
	}
}

// NewToolResultMessage creates a new tool result message
func NewToolResultMessage(toolCallID, result string) Message {
	return Message{
		Role:       RoleTool,
		ToolCallID: toolCallID,
		Content: []Content{
			{Type: ContentTypeToolResult, ToolCallID: toolCallID, ToolResult: result},
		},
	}
}

// GetText returns the text content of the message
func (m *Message) GetText() string {
	for _, c := range m.Content {
		if c.Type == ContentTypeText {
			return c.Text
		}
	}
	return ""
}

// HasImage returns true if the message contains an image
func (m *Message) HasImage() bool {
	for _, c := range m.Content {
		if c.Type == ContentTypeImage {
			return true
		}
	}
	return false
}

// MarshalJSON custom marshaling for messages
func (m Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(m),
	})
}

// ToolCall represents a function call made by the model
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
}

// Tool represents a tool/function that can be called by the model
type Tool struct {
	Type     string   `json:"type"` // "function"
	Function Function `json:"function"`
}

// Function defines a callable function
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
	Strict      bool                   `json:"strict,omitempty"`
}

// NewTool creates a new tool definition
func NewTool(name, description string, parameters map[string]interface{}) Tool {
	return Tool{
		Type: "function",
		Function: Function{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}
