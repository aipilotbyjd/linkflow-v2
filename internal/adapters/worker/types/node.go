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
	Beta        bool            `json:"beta,omitempty"`        // Indicates beta/experimental features
	Inputs      []NodePort      `json:"inputs"`
	Outputs     []NodePort      `json:"outputs"`
	Parameters  []NodeParameter `json:"parameters"`
	Credentials []string        `json:"credentials,omitempty"`
	Tags        []string        `json:"tags,omitempty"`        // Tags for categorization/search
	Examples    []NodeExample   `json:"examples,omitempty"`    // Example configurations
	Links       []NodeLink      `json:"links,omitempty"`       // Documentation links
}

// NodePort represents an input or output port
type NodePort struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // any, string, number, boolean, array, object
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	DefaultValue interface{} `json:"default_value,omitempty"` // Default value for input ports
	Schema      interface{} `json:"schema,omitempty"`        // JSON Schema for validation
	Example     interface{} `json:"example,omitempty"`       // Example value
	Multiple    bool        `json:"multiple,omitempty"`      // Supports multiple connections
	MaxConnections int      `json:"max_connections,omitempty"` // Maximum allowed connections
	Color       string      `json:"color,omitempty"`         // Port color
	Icon        string      `json:"icon,omitempty"`          // Port icon
	Label       string      `json:"label,omitempty"`         // Custom label
	Hidden      bool        `json:"hidden,omitempty"`        // Hide from UI
}

// NodeParameter represents a configurable parameter
type NodeParameter struct {
	Name               string        `json:"name"`
	DisplayName        string        `json:"display_name"`
	Type               string        `json:"type"` // string, number, boolean, options, json, code, credential
	Default            interface{}   `json:"default,omitempty"`
	Required           bool          `json:"required,omitempty"`
	Description        string        `json:"description,omitempty"`
	LongDescription    string        `json:"long_description,omitempty"` // Detailed help text
	Options            []ParamOption `json:"options,omitempty"`
	Placeholder        string        `json:"placeholder,omitempty"`
	DependsOn          string        `json:"depends_on,omitempty"`
	ShowIf             string        `json:"show_if,omitempty"`
	HideIf             string        `json:"hide_if,omitempty"`          // Hide parameter based on condition
	Validation         *Validation   `json:"validation,omitempty"`
	ExpressionDisabled bool          `json:"expression_disabled,omitempty"`
	CodeLanguage       string        `json:"code_language,omitempty"`    // json, sql, javascript, html
	Advanced           bool          `json:"advanced,omitempty"`         // Show in advanced section
	Sensitive          bool          `json:"sensitive,omitempty"`        // Mask value in UI
	ReadOnly           bool          `json:"read_only,omitempty"`        // Parameter cannot be edited
	Group              string        `json:"group,omitempty"`            // Group parameters in UI sections
	Order              int           `json:"order,omitempty"`            // Display order within group
	Tooltip            string        `json:"tooltip,omitempty"`          // Short tooltip text
	Warning            string        `json:"warning,omitempty"`          // Warning message to show
	Error              string        `json:"error,omitempty"`            // Error message to show
	Loading            bool          `json:"loading,omitempty"`          // Show loading indicator
	Width              string        `json:"width,omitempty"`            // UI width (e.g., "full", "half", "third")
	Height             string        `json:"height,omitempty"`           // UI height for textareas
	Rows               int           `json:"rows,omitempty"`             // Number of rows for textarea
	Cols               int           `json:"cols,omitempty"`             // Number of columns for textarea
	Autocomplete       string        `json:"autocomplete,omitempty"`     // HTML autocomplete attribute
	AutoFocus          bool          `json:"auto_focus,omitempty"`       // Auto-focus on mount
	Clearable          bool          `json:"clearable,omitempty"`        // Show clear button
	Searchable         bool          `json:"searchable,omitempty"`       // Enable search in dropdowns
	Multiple           bool          `json:"multiple,omitempty"`         // Allow multiple selections
	Delimiter          string        `json:"delimiter,omitempty"`        // Delimiter for multiple values
	Transform          string        `json:"transform,omitempty"`        // Transform function name
	Format             string        `json:"format,omitempty"`           // Display format (e.g., "currency", "date")
	Prefix             string        `json:"prefix,omitempty"`           // Prefix for input
	Suffix             string        `json:"suffix,omitempty"`           // Suffix for input
	Step               float64       `json:"step,omitempty"`             // Step increment for number inputs
}

// Validation defines validation rules for a parameter
type Validation struct {
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`
	MinLength    *int     `json:"min_length,omitempty"`
	MaxLength    *int     `json:"max_length,omitempty"`
	Pattern      string   `json:"pattern,omitempty"`
	Custom       string   `json:"custom,omitempty"`        // Custom validation function name
	UniqueItems  bool     `json:"unique_items,omitempty"`  // Array items must be unique
	Enum         []interface{} `json:"enum,omitempty"`     // Allowed values
	Format       string   `json:"format,omitempty"`        // Standard formats (email, uri, etc.)
	ExclusiveMin bool     `json:"exclusive_min,omitempty"` // Exclusive minimum
	ExclusiveMax bool     `json:"exclusive_max,omitempty"` // Exclusive maximum
	MultipleOf   *float64 `json:"multiple_of,omitempty"`   // Must be multiple of this value
	RequiredIf   string   `json:"required_if,omitempty"`   // Conditional requirement
}

// ParamOption represents a selectable option for a parameter
type ParamOption struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
	Icon        string      `json:"icon,omitempty"`         // Icon for the option
	Color       string      `json:"color,omitempty"`        // Color for the option
	Disabled    bool        `json:"disabled,omitempty"`     // Option is disabled
	Group       string      `json:"group,omitempty"`        // Group name for grouping options
	Badge       string      `json:"badge,omitempty"`        // Badge text to show
	BadgeColor  string      `json:"badge_color,omitempty"`  // Badge color
	SearchTerms []string    `json:"search_terms,omitempty"` // Additional terms for searching
	Meta        interface{} `json:"meta,omitempty"`         // Additional metadata
}

// Helpers for creating pointers to literals
func Float64Ptr(v float64) *float64 { return &v }
func IntPtr(v int) *int             { return &v }
func StringPtr(v string) *string    { return &v }
func BoolPtr(v bool) *bool          { return &v }

// NodeExample represents an example configuration for a node
type NodeExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters"`
	Input       interface{}            `json:"input,omitempty"`
	Output      interface{}            `json:"output,omitempty"`
}

// NodeLink represents a documentation or resource link
type NodeLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type,omitempty"` // "docs", "api", "tutorial", "github"
}
