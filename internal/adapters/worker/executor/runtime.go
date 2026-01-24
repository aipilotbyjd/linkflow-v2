package executor

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

var expressionRegex = regexp.MustCompile(`{{\s*([^}]+)\s*}}`)

type Runtime struct {
	Execution   *execution.Execution
	Workflow    *workflow.Workflow
	Logger      logger.Logger
	nodeOutputs map[string]types.JSON
	variables   map[string]interface{}
	nodeCount   int64
	mu          sync.RWMutex
}

func NewRuntime(exec *execution.Execution, wf *workflow.Workflow, log logger.Logger) *Runtime {
	return &Runtime{
		Execution:   exec,
		Workflow:    wf,
		Logger:      log,
		nodeOutputs: make(map[string]types.JSON),
		variables:   make(map[string]interface{}),
		nodeCount:   0,
	}
}

func (r *Runtime) IncrementNodeCount() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodeCount++
	return r.nodeCount
}

func (r *Runtime) GetNodeCount() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nodeCount
}

func (r *Runtime) GetNodes() []map[string]interface{} {
	var nodes []map[string]interface{}
	for _, item := range r.Workflow.Nodes {
		if m, ok := item.(map[string]interface{}); ok {
			nodes = append(nodes, m)
		}
	}
	return nodes
}

func (r *Runtime) GetConnections() []map[string]interface{} {
	var connections []map[string]interface{}
	for _, item := range r.Workflow.Connections {
		if m, ok := item.(map[string]interface{}); ok {
			connections = append(connections, m)
		}
	}
	return connections
}

func (r *Runtime) GetNextNodes(nodeID string, sourcePort string) []map[string]interface{} {
	connections := r.GetConnections()
	nodes := r.GetNodes()

	var nextNodeIDs []string
	for _, conn := range connections {
		source, _ := conn["source"].(string)
		if source == "" {
			source, _ = conn["source_node"].(string)
		}

		if source == nodeID {
			connPort, _ := conn["source_port"].(string)
			if connPort == "" {
				connPort = "main"
			}

			// If sourcePort is "main" or empty, and connection is "main" or empty, match.
			match := false
			if sourcePort == "" || sourcePort == "main" {
				if connPort == "" || connPort == "main" {
					match = true
				}
			} else {
				if connPort == sourcePort {
					match = true
				}
			}

			if match {
				target, _ := conn["target"].(string)
				if target == "" {
					target, _ = conn["target_node"].(string)
				}
				if target != "" {
					nextNodeIDs = append(nextNodeIDs, target)
				}
			}
		}
	}

	var nextNodes []map[string]interface{}
	for _, node := range nodes {
		if id, ok := node["id"].(string); ok {
			for _, nextID := range nextNodeIDs {
				if id == nextID {
					nextNodes = append(nextNodes, node)
				}
			}
		}
	}

	return nextNodes
}

func (r *Runtime) SetNodeOutput(nodeID string, output types.JSON) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodeOutputs[nodeID] = output
}

func (r *Runtime) GetNodeOutput(nodeID string) types.JSON {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nodeOutputs[nodeID]
}

func (r *Runtime) GetInputData() types.JSON {
	return r.Execution.InputData
}

func (r *Runtime) GetOutputData() types.JSON {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.nodeOutputs) == 0 {
		return nil
	}

	result := make(types.JSON)
	for k, v := range r.nodeOutputs {
		result[k] = v
	}
	return result
}

func (r *Runtime) SetVariable(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.variables[key] = value
}

func (r *Runtime) GetVariable(key string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.variables[key]
}

// GetCredentialValue gets a credential value by type and key
func (r *Runtime) GetCredentialValue(credType, key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	credKey := credType + ":" + key
	if val, ok := r.variables[credKey].(string); ok {
		return val
	}
	return ""
}

// SetCredentialValues sets credential values
func (r *Runtime) SetCredentialValues(credType string, values map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, value := range values {
		r.variables[credType+":"+key] = value
	}
}

// SubWorkflowResult represents the result of a sub-workflow execution
type SubWorkflowResult struct {
	ExecutionID string
	Status      string
	Output      map[string]interface{}
	StartedAt   string
	CompletedAt string
}

// ApprovalRequest represents a request for human approval
type ApprovalRequest struct {
	Title       string
	Description string
	Approvers   []string
	TimeoutAt   interface{}
	NotifyVia   []string
	RequireAll  bool
	Data        map[string]interface{}
}

// EventListener represents an event listener registration
type EventListener struct {
	EventType string
	Filter    map[string]interface{}
	TimeoutAt interface{}
}

// ExecuteSubWorkflow executes another workflow as a sub-workflow
func (r *Runtime) ExecuteSubWorkflow(ctx interface{}, workflowID interface{}, input map[string]interface{}, wait bool) (*SubWorkflowResult, error) {
	// This is a stub - actual implementation requires workflow service injection
	return &SubWorkflowResult{
		ExecutionID: "sub-exec-placeholder",
		Status:      "completed",
		Output:      input,
		StartedAt:   "",
		CompletedAt: "",
	}, nil
}

// CreateApprovalRequest creates an approval request and returns approval ID and resume URL
func (r *Runtime) CreateApprovalRequest(ctx interface{}, req ApprovalRequest) (string, string, error) {
	// This is a stub - actual implementation requires approval service injection
	return "approval-placeholder", "https://app.example.com/approve/placeholder", nil
}

// RegisterEventListener registers an event listener and returns listener ID and webhook URL
func (r *Runtime) RegisterEventListener(ctx interface{}, listener EventListener) (string, string, error) {
	// This is a stub - actual implementation requires event service injection
	return "listener-placeholder", "https://app.example.com/webhook/placeholder", nil
}

// Node represents a workflow node
type Node struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
	Position   Position               `json:"position,omitempty"`
}

// Position represents the x,y coordinates of a node
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ResolveParameters evaluates all expressions in the parameters map
func (r *Runtime) ResolveParameters(params map[string]interface{}) map[string]interface{} {
	resolved := make(map[string]interface{})
	for k, v := range params {
		resolved[k] = r.evaluateValue(v)
	}
	return resolved
}

func (r *Runtime) evaluateValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return r.evaluateExpression(val)
	case map[string]interface{}:
		res := make(map[string]interface{})
		for k, v2 := range val {
			res[k] = r.evaluateValue(v2)
		}
		return res
	case []interface{}:
		res := make([]interface{}, len(val))
		for i, v2 := range val {
			res[i] = r.evaluateValue(v2)
		}
		return res
	default:
		return v
	}
}

func (r *Runtime) evaluateExpression(s string) interface{} {
	// If the entire string is an expression, we try to return the raw object
	if strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}") && strings.Count(s, "{{") == 1 {
		path := strings.TrimSpace(s[2 : len(s)-2])
		return r.getValueByPath(path)
	}

	// Otherwise, we perform string replacement
	return expressionRegex.ReplaceAllStringFunc(s, func(m string) string {
		path := strings.TrimSpace(m[2 : len(m)-2])
		val := r.getValueByPath(path)
		return fmt.Sprintf("%v", val)
	})
}

func (r *Runtime) getValueByPath(path string) interface{} {
	if strings.HasPrefix(path, "$input") {
		remaining := strings.TrimPrefix(path, "$input")
		if remaining == "" {
			return r.GetInputData()
		}
		return getValueByPath(r.GetInputData(), strings.TrimPrefix(remaining, "."))
	}

	// Handle $node.NODE_ID.path
	if strings.HasPrefix(path, "$node") {
		parts := strings.SplitN(strings.TrimPrefix(path, "$node."), ".", 2)
		if len(parts) >= 1 {
			nodeID := parts[0]
			output := r.GetNodeOutput(nodeID)
			if len(parts) == 1 {
				return output
			}
			return getValueByPath(output, parts[1])
		}
	}

	return nil
}

func getValueByPath(data interface{}, path string) interface{} {
	if path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else if m, ok := current.(types.JSON); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// NodeOutput represents the output of a node execution
type NodeOutput struct {
	Data     interface{}            `json:"data,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
