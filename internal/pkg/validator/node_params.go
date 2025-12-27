package validator

import (
	"fmt"
	"net/url"
	"strings"
)

// NodeParameterValidator validates node-specific parameters
type NodeParameterValidator struct {
	schemas map[string]NodeParamSchema
}

// NodeParamSchema defines required and optional parameters for a node type
type NodeParamSchema struct {
	NodeType   string
	Required   []ParamDef
	Optional   []ParamDef
	Validators []ParamValidator
}

// ParamDef defines a parameter
type ParamDef struct {
	Name        string
	Type        ParamType
	Description string
	Default     interface{}
	Enum        []string // For enum types
}

// ParamType represents parameter data types
type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeNumber  ParamType = "number"
	ParamTypeBool    ParamType = "boolean"
	ParamTypeArray   ParamType = "array"
	ParamTypeObject  ParamType = "object"
	ParamTypeEnum    ParamType = "enum"
	ParamTypeURL     ParamType = "url"
	ParamTypeEmail   ParamType = "email"
	ParamTypeCron    ParamType = "cron"
	ParamTypeJSON    ParamType = "json"
	ParamTypeExpr    ParamType = "expression"
	ParamTypeCredRef ParamType = "credential_ref"
)

// ParamValidator is a custom validation function
type ParamValidator func(params map[string]interface{}) *NodeParamError

// NodeParamError represents a parameter validation error
type NodeParamError struct {
	NodeID  string `json:"node_id,omitempty"`
	Param   string `json:"param"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *NodeParamError) Error() string {
	if e.NodeID != "" {
		return fmt.Sprintf("[%s.%s] %s", e.NodeID, e.Param, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Param, e.Message)
}

// NewNodeParameterValidator creates a new validator with default schemas
func NewNodeParameterValidator() *NodeParameterValidator {
	v := &NodeParameterValidator{
		schemas: make(map[string]NodeParamSchema),
	}
	v.registerDefaultSchemas()
	return v
}

// registerDefaultSchemas registers schemas for built-in node types
func (v *NodeParameterValidator) registerDefaultSchemas() {
	// HTTP Request node
	v.RegisterSchema(NodeParamSchema{
		NodeType: "action.http",
		Required: []ParamDef{
			{Name: "url", Type: ParamTypeURL, Description: "Request URL"},
			{Name: "method", Type: ParamTypeEnum, Enum: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}},
		},
		Optional: []ParamDef{
			{Name: "headers", Type: ParamTypeObject},
			{Name: "body", Type: ParamTypeString},
			{Name: "timeout", Type: ParamTypeNumber, Default: 30},
			{Name: "retry", Type: ParamTypeNumber, Default: 0},
		},
	})

	// Code node
	v.RegisterSchema(NodeParamSchema{
		NodeType: "action.code",
		Required: []ParamDef{
			{Name: "code", Type: ParamTypeString, Description: "JavaScript code to execute"},
		},
		Optional: []ParamDef{
			{Name: "language", Type: ParamTypeEnum, Enum: []string{"javascript", "python"}, Default: "javascript"},
		},
	})

	// Email node
	v.RegisterSchema(NodeParamSchema{
		NodeType: "integration.email",
		Required: []ParamDef{
			{Name: "to", Type: ParamTypeString, Description: "Recipient email"},
			{Name: "subject", Type: ParamTypeString, Description: "Email subject"},
		},
		Optional: []ParamDef{
			{Name: "body", Type: ParamTypeString},
			{Name: "html", Type: ParamTypeString},
			{Name: "cc", Type: ParamTypeString},
			{Name: "bcc", Type: ParamTypeString},
			{Name: "attachments", Type: ParamTypeArray},
		},
	})

	// Slack node
	v.RegisterSchema(NodeParamSchema{
		NodeType: "integration.slack",
		Required: []ParamDef{
			{Name: "channel", Type: ParamTypeString, Description: "Slack channel"},
		},
		Optional: []ParamDef{
			{Name: "text", Type: ParamTypeString},
			{Name: "blocks", Type: ParamTypeArray},
			{Name: "attachments", Type: ParamTypeArray},
			{Name: "username", Type: ParamTypeString},
			{Name: "icon_emoji", Type: ParamTypeString},
		},
		Validators: []ParamValidator{
			func(params map[string]interface{}) *NodeParamError {
				_, hasText := params["text"]
				_, hasBlocks := params["blocks"]
				if !hasText && !hasBlocks {
					return &NodeParamError{
						Param:   "text",
						Code:    "MISSING_CONTENT",
						Message: "Either 'text' or 'blocks' is required",
					}
				}
				return nil
			},
		},
	})

	// IF condition node
	v.RegisterSchema(NodeParamSchema{
		NodeType: "logic.if",
		Required: []ParamDef{
			{Name: "conditions", Type: ParamTypeArray, Description: "Condition rules"},
		},
		Optional: []ParamDef{
			{Name: "combineWith", Type: ParamTypeEnum, Enum: []string{"and", "or"}, Default: "and"},
		},
	})

	// Loop node
	v.RegisterSchema(NodeParamSchema{
		NodeType: "logic.loop",
		Required: []ParamDef{
			{Name: "items", Type: ParamTypeExpr, Description: "Items to iterate over"},
		},
		Optional: []ParamDef{
			{Name: "batchSize", Type: ParamTypeNumber, Default: 1},
		},
	})

	// Schedule trigger
	v.RegisterSchema(NodeParamSchema{
		NodeType: "trigger.schedule",
		Required: []ParamDef{
			{Name: "cron", Type: ParamTypeCron, Description: "Cron expression"},
		},
		Optional: []ParamDef{
			{Name: "timezone", Type: ParamTypeString, Default: "UTC"},
		},
	})

	// Webhook trigger
	v.RegisterSchema(NodeParamSchema{
		NodeType: "trigger.webhook",
		Optional: []ParamDef{
			{Name: "path", Type: ParamTypeString},
			{Name: "method", Type: ParamTypeEnum, Enum: []string{"GET", "POST", "PUT", "DELETE", "PATCH"}, Default: "POST"},
			{Name: "responseMode", Type: ParamTypeEnum, Enum: []string{"onReceived", "lastNode"}, Default: "onReceived"},
		},
	})

	// Database nodes
	for _, nodeType := range []string{"integration.postgres", "integration.mysql"} {
		v.RegisterSchema(NodeParamSchema{
			NodeType: nodeType,
			Required: []ParamDef{
				{Name: "operation", Type: ParamTypeEnum, Enum: []string{"select", "insert", "update", "delete", "raw"}},
			},
			Optional: []ParamDef{
				{Name: "table", Type: ParamTypeString},
				{Name: "query", Type: ParamTypeString},
				{Name: "values", Type: ParamTypeObject},
				{Name: "where", Type: ParamTypeObject},
			},
		})
	}
}

// RegisterSchema registers a parameter schema for a node type
func (v *NodeParameterValidator) RegisterSchema(schema NodeParamSchema) {
	v.schemas[schema.NodeType] = schema
}

// Validate validates parameters for a node
func (v *NodeParameterValidator) Validate(nodeType string, nodeID string, params map[string]interface{}) []*NodeParamError {
	var errors []*NodeParamError

	schema, ok := v.schemas[nodeType]
	if !ok {
		// No schema defined, skip validation
		return nil
	}

	// Check required parameters
	for _, param := range schema.Required {
		val, exists := params[param.Name]
		if !exists || val == nil || val == "" {
			errors = append(errors, &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "REQUIRED_PARAM",
				Message: fmt.Sprintf("Required parameter '%s' is missing", param.Name),
			})
			continue
		}

		// Type validation
		if err := v.validateParamType(nodeID, param, val); err != nil {
			errors = append(errors, err)
		}
	}

	// Validate optional parameters if present
	for _, param := range schema.Optional {
		val, exists := params[param.Name]
		if !exists || val == nil {
			continue
		}
		if err := v.validateParamType(nodeID, param, val); err != nil {
			errors = append(errors, err)
		}
	}

	// Run custom validators
	for _, validator := range schema.Validators {
		if err := validator(params); err != nil {
			err.NodeID = nodeID
			errors = append(errors, err)
		}
	}

	return errors
}

// validateParamType validates a parameter value against its expected type
func (v *NodeParameterValidator) validateParamType(nodeID string, param ParamDef, value interface{}) *NodeParamError {
	// Allow expressions for any type
	if str, ok := value.(string); ok {
		if strings.Contains(str, "{{") && strings.Contains(str, "}}") {
			return nil // Expression, skip type validation
		}
	}

	switch param.Type {
	case ParamTypeString, ParamTypeExpr:
		if _, ok := value.(string); !ok {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_TYPE",
				Message: fmt.Sprintf("Parameter '%s' must be a string", param.Name),
			}
		}

	case ParamTypeNumber:
		switch value.(type) {
		case int, int64, float64, float32:
			// OK
		default:
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_TYPE",
				Message: fmt.Sprintf("Parameter '%s' must be a number", param.Name),
			}
		}

	case ParamTypeBool:
		if _, ok := value.(bool); !ok {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_TYPE",
				Message: fmt.Sprintf("Parameter '%s' must be a boolean", param.Name),
			}
		}

	case ParamTypeArray:
		if _, ok := value.([]interface{}); !ok {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_TYPE",
				Message: fmt.Sprintf("Parameter '%s' must be an array", param.Name),
			}
		}

	case ParamTypeObject:
		if _, ok := value.(map[string]interface{}); !ok {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_TYPE",
				Message: fmt.Sprintf("Parameter '%s' must be an object", param.Name),
			}
		}

	case ParamTypeEnum:
		str, ok := value.(string)
		if !ok {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_TYPE",
				Message: fmt.Sprintf("Parameter '%s' must be a string", param.Name),
			}
		}
		valid := false
		for _, e := range param.Enum {
			if strings.EqualFold(str, e) {
				valid = true
				break
			}
		}
		if !valid {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_ENUM",
				Message: fmt.Sprintf("Parameter '%s' must be one of: %s", param.Name, strings.Join(param.Enum, ", ")),
			}
		}

	case ParamTypeURL:
		str, ok := value.(string)
		if !ok {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_TYPE",
				Message: fmt.Sprintf("Parameter '%s' must be a string", param.Name),
			}
		}
		if _, err := url.ParseRequestURI(str); err != nil {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_URL",
				Message: fmt.Sprintf("Parameter '%s' must be a valid URL", param.Name),
			}
		}

	case ParamTypeCron:
		str, ok := value.(string)
		if !ok {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_TYPE",
				Message: fmt.Sprintf("Parameter '%s' must be a string", param.Name),
			}
		}
		parts := strings.Split(str, " ")
		if len(parts) != 5 && len(parts) != 6 {
			return &NodeParamError{
				NodeID:  nodeID,
				Param:   param.Name,
				Code:    "INVALID_CRON",
				Message: fmt.Sprintf("Parameter '%s' must be a valid cron expression", param.Name),
			}
		}
	}

	return nil
}

// ValidateNodeParameters validates parameters for multiple nodes
func ValidateNodeParameters(nodes []WorkflowNode) []*NodeParamError {
	validator := NewNodeParameterValidator()
	var allErrors []*NodeParamError

	for _, node := range nodes {
		errors := validator.Validate(node.Type, node.ID, node.Parameters)
		allErrors = append(allErrors, errors...)
	}

	return allErrors
}
