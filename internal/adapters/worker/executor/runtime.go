package executor

import (
	"sync"

	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type Runtime struct {
	Execution   *execution.Execution
	Workflow    *workflow.Workflow
	Logger      logger.Logger
	nodeOutputs map[string]types.JSON
	variables   map[string]interface{}
	mu          sync.RWMutex
}

func NewRuntime(exec *execution.Execution, wf *workflow.Workflow, log logger.Logger) *Runtime {
	return &Runtime{
		Execution:   exec,
		Workflow:    wf,
		Logger:      log,
		nodeOutputs: make(map[string]types.JSON),
		variables:   make(map[string]interface{}),
	}
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

func (r *Runtime) GetNextNodes(nodeID string) []map[string]interface{} {
	connections := r.GetConnections()
	nodes := r.GetNodes()

	var nextNodeIDs []string
	for _, conn := range connections {
		if source, ok := conn["source"].(string); ok && source == nodeID {
			if target, ok := conn["target"].(string); ok {
				nextNodeIDs = append(nextNodeIDs, target)
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

// NodeOutput represents the output of a node execution
type NodeOutput struct {
	Data     interface{}            `json:"data,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
