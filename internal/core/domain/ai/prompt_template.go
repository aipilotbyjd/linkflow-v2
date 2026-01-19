package ai

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PromptTemplate represents a reusable prompt template
type PromptTemplate struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	CreatedBy   uuid.UUID `json:"created_by"`

	// Template info
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`

	// Template content
	Template  string             `json:"template"`
	Variables []TemplateVariable `json:"variables"`

	// System message (optional)
	SystemMessage string `json:"system_message,omitempty"`

	// Default model settings
	DefaultModel       string   `json:"default_model,omitempty"`
	DefaultTemperature *float64 `json:"default_temperature,omitempty"`
	DefaultMaxTokens   int      `json:"default_max_tokens,omitempty"`

	// Versioning
	Version   int  `json:"version"`
	IsActive  bool `json:"is_active"`
	IsPublic  bool `json:"is_public"`

	// Metadata
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TemplateVariable represents a variable in a prompt template
type TemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // string, number, boolean, array, object
	Description  string      `json:"description,omitempty"`
	Required     bool        `json:"required"`
	Default      interface{} `json:"default,omitempty"`
	Validation   string      `json:"validation,omitempty"` // regex pattern
}

// NewPromptTemplate creates a new prompt template
func NewPromptTemplate(workspaceID, createdBy uuid.UUID, name, template string) *PromptTemplate {
	pt := &PromptTemplate{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		CreatedBy:   createdBy,
		Name:        name,
		Template:    template,
		Version:     1,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Extract variables from template
	pt.Variables = pt.extractVariables()

	return pt
}

// extractVariables extracts variable placeholders from the template
func (pt *PromptTemplate) extractVariables() []TemplateVariable {
	// Match {{variable}} or {{variable:type}} patterns
	re := regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)(?::([a-zA-Z]+))?\}\}`)
	matches := re.FindAllStringSubmatch(pt.Template, -1)

	seen := make(map[string]bool)
	var variables []TemplateVariable

	for _, match := range matches {
		name := match[1]
		if seen[name] {
			continue
		}
		seen[name] = true

		varType := "string"
		if len(match) > 2 && match[2] != "" {
			varType = match[2]
		}

		variables = append(variables, TemplateVariable{
			Name:     name,
			Type:     varType,
			Required: true,
		})
	}

	return variables
}

// Render renders the template with provided values
func (pt *PromptTemplate) Render(values map[string]interface{}) (string, error) {
	result := pt.Template

	for _, v := range pt.Variables {
		placeholder := "{{" + v.Name + "}}"
		placeholderTyped := "{{" + v.Name + ":" + v.Type + "}}"

		value, ok := values[v.Name]
		if !ok {
			if v.Required && v.Default == nil {
				continue // Will be caught by validation
			}
			value = v.Default
		}

		var strValue string
		switch val := value.(type) {
		case string:
			strValue = val
		case []interface{}, map[string]interface{}:
			b, _ := json.Marshal(val)
			strValue = string(b)
		default:
			strValue = toString(val)
		}

		result = strings.ReplaceAll(result, placeholder, strValue)
		result = strings.ReplaceAll(result, placeholderTyped, strValue)
	}

	return result, nil
}

// Validate validates the provided values against the template variables
func (pt *PromptTemplate) Validate(values map[string]interface{}) []string {
	var errors []string

	for _, v := range pt.Variables {
		value, ok := values[v.Name]
		if !ok || value == nil {
			if v.Required && v.Default == nil {
				errors = append(errors, "missing required variable: "+v.Name)
			}
			continue
		}

		// Type validation
		if !pt.validateType(value, v.Type) {
			errors = append(errors, "invalid type for variable "+v.Name+": expected "+v.Type)
		}

		// Pattern validation
		if v.Validation != "" {
			if str, ok := value.(string); ok {
				re, err := regexp.Compile(v.Validation)
				if err == nil && !re.MatchString(str) {
					errors = append(errors, "variable "+v.Name+" does not match pattern: "+v.Validation)
				}
			}
		}
	}

	return errors
}

func (pt *PromptTemplate) validateType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			return true
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return true
	}
}

// Clone creates a copy of the template
func (pt *PromptTemplate) Clone(newWorkspaceID uuid.UUID) *PromptTemplate {
	clone := *pt
	clone.ID = uuid.New()
	clone.WorkspaceID = newWorkspaceID
	clone.Version = 1
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = time.Now()
	return &clone
}

// ToMessages converts the rendered template to messages
func (pt *PromptTemplate) ToMessages(values map[string]interface{}) ([]Message, error) {
	rendered, err := pt.Render(values)
	if err != nil {
		return nil, err
	}

	var messages []Message

	if pt.SystemMessage != "" {
		messages = append(messages, NewSystemMessage(pt.SystemMessage))
	}

	messages = append(messages, NewUserMessage(rendered))

	return messages, nil
}

// PromptTemplateRepository defines the interface for prompt template persistence
type PromptTemplateRepository interface {
	Create(template *PromptTemplate) error
	Update(template *PromptTemplate) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*PromptTemplate, error)
	FindByWorkspace(workspaceID uuid.UUID, category string, limit, offset int) ([]PromptTemplate, int64, error)
	FindPublic(category string, limit, offset int) ([]PromptTemplate, int64, error)
	Search(workspaceID uuid.UUID, query string, limit int) ([]PromptTemplate, error)
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case bool:
		if val {
			return "true"
		}
		return "false"
	case string:
		return val
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}
