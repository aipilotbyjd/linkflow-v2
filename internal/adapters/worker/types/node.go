package types

// NodeMetadata describes a node type
type NodeMetadata struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Version     string          `json:"version"`
	Icon        string          `json:"icon,omitempty"`
	Color       string          `json:"color,omitempty"`
	Deprecated  bool            `json:"deprecated,omitempty"`
	Inputs      []NodePort      `json:"inputs"`
	Outputs     []NodePort      `json:"outputs"`
	Parameters  []NodeParameter `json:"parameters"`
	Credentials []string        `json:"credentials,omitempty"`
}

// NodePort represents an input or output port
type NodePort struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // any, string, number, boolean, array, object
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// NodeParameter represents a configurable parameter
type NodeParameter struct {
	Name               string        `json:"name"`
	DisplayName        string        `json:"display_name"`
	Type               string        `json:"type"` // string, number, boolean, options, json, code, credential
	Default            interface{}   `json:"default,omitempty"`
	Required           bool          `json:"required,omitempty"`
	Description        string        `json:"description,omitempty"`
	Options            []ParamOption `json:"options,omitempty"`
	Placeholder        string        `json:"placeholder,omitempty"`
	DependsOn          string        `json:"depends_on,omitempty"`
	ShowIf             string        `json:"show_if,omitempty"`
	Validation         *Validation   `json:"validation,omitempty"`
	ExpressionDisabled bool          `json:"expression_disabled,omitempty"`
	CodeLanguage       string        `json:"code_language,omitempty"` // json, sql, javascript, html
}

// Validation defines validation rules for a parameter
type Validation struct {
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
}

// ParamOption represents a selectable option for a parameter
type ParamOption struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
}

// Helpers for creating pointers to literals
func Float64Ptr(v float64) *float64 { return &v }
func IntPtr(v int) *int             { return &v }
func StringPtr(v string) *string    { return &v }
func BoolPtr(v bool) *bool          { return &v }
