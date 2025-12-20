package processor

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// WorkflowDefinition represents a parsed workflow ready for execution
type WorkflowDefinition struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Nodes       []*NodeDefinition
	Connections []*Connection
	Settings    WorkflowSettings
}

// NodeDefinition represents a node in the workflow
type NodeDefinition struct {
	ID         string
	Type       string
	Name       string
	Config     map[string]interface{}
	Position   Position
	Inputs     []ConnectionRef
	Disabled   bool
	Notes      string
	RetryOnFail bool
	MaxRetries  int
	Timeout     time.Duration
}

// Position represents node position in the editor
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Connection represents an edge between nodes
type Connection struct {
	ID           string
	SourceNodeID string
	SourceHandle string
	TargetNodeID string
	TargetHandle string
}

// ConnectionRef represents an input connection reference
type ConnectionRef struct {
	SourceNodeID string `json:"source_node_id"`
	SourceHandle string `json:"source_handle"`
}

// WorkflowSettings contains workflow-level configuration
type WorkflowSettings struct {
	Timezone           string
	ErrorWorkflow      string
	CallerPolicy       string
	SaveExecutionData  bool
	SaveSuccessfulData bool
	ExecutionTimeout   time.Duration
}

// Input represents workflow input data
type Input map[string]interface{}

// Result represents workflow execution result
type Result struct {
	ExecutionID    uuid.UUID
	Status         ExecutionStatus
	Output         map[string]interface{}
	NodeResults    map[string]*NodeResult
	StartedAt      time.Time
	CompletedAt    time.Time
	Duration       time.Duration
	NodesExecuted  int
	Error          string
	ErrorNodeID    string
}

// NodeResult represents a single node execution result
type NodeResult struct {
	NodeID      string
	NodeType    string
	NodeName    string
	Status      NodeStatus
	Input       map[string]interface{}
	Output      map[string]interface{}
	Error       string
	StartedAt   time.Time
	CompletedAt time.Time
	Duration    time.Duration
	Retries     int
	Cached      bool
}

// PreviewResult represents a dry-run preview result
type PreviewResult struct {
	Valid          bool
	Errors         []ValidationError
	Warnings       []ValidationWarning
	ExecutionOrder []string
	NodePreviews   map[string]*NodePreview
}

// NodePreview represents preview data for a node
type NodePreview struct {
	NodeID        string
	WouldExecute  bool
	ResolvedInput map[string]interface{}
	DependsOn     []string
}

// ValidationError represents a workflow validation error
type ValidationError struct {
	NodeID  string
	Field   string
	Message string
	Code    string
}

// ValidationWarning represents a workflow validation warning
type ValidationWarning struct {
	NodeID  string
	Field   string
	Message string
	Code    string
}

// ExecutionStatus represents workflow execution status
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
	StatusTimedOut  ExecutionStatus = "timed_out"
)

// NodeStatus represents node execution status
type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusCompleted NodeStatus = "completed"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusSkipped   NodeStatus = "skipped"
	NodeStatusCached    NodeStatus = "cached"
)

// ExecutionOptions configures how a workflow is executed
type ExecutionOptions struct {
	MaxParallelNodes   int
	DefaultNodeTimeout time.Duration
	WorkflowTimeout    time.Duration
	EnableCaching      bool
	DryRun             bool
	StartFromNode      string
	StopAtNode         string
	SkipNodes          []string
	NodeOverrides      map[string]map[string]interface{}
}

// DefaultExecutionOptions returns sensible defaults
func DefaultExecutionOptions() ExecutionOptions {
	return ExecutionOptions{
		MaxParallelNodes:   10,
		DefaultNodeTimeout: 5 * time.Minute,
		WorkflowTimeout:    30 * time.Minute,
		EnableCaching:      true,
		DryRun:             false,
	}
}

// Branch represents a parallel execution branch
type Branch struct {
	ID    string
	Nodes []string
}

// BranchResult represents result from a parallel branch
type BranchResult struct {
	BranchID string
	Results  map[string]*NodeResult
	Error    error
}

// ParseWorkflow parses a workflow model into a definition
func ParseWorkflow(workflow *models.Workflow) (*WorkflowDefinition, error) {
	def := &WorkflowDefinition{
		ID:          workflow.ID,
		WorkspaceID: workflow.WorkspaceID,
		Name:        workflow.Name,
	}

	// Parse nodes
	nodes, err := parseNodes(workflow.Nodes)
	if err != nil {
		return nil, err
	}
	def.Nodes = nodes

	// Parse connections
	connections, err := parseConnections(workflow.Connections)
	if err != nil {
		return nil, err
	}
	def.Connections = connections

	// Parse settings
	if workflow.Settings != nil {
		def.Settings = parseSettings(workflow.Settings)
	}

	return def, nil
}

func parseNodes(data models.JSON) ([]*NodeDefinition, error) {
	if data == nil {
		return nil, nil
	}

	nodesRaw, ok := data["nodes"].([]interface{})
	if !ok {
		return nil, nil
	}

	nodes := make([]*NodeDefinition, 0, len(nodesRaw))
	for _, nodeRaw := range nodesRaw {
		nodeMap, ok := nodeRaw.(map[string]interface{})
		if !ok {
			continue
		}

		node := &NodeDefinition{
			ID:   getString(nodeMap, "id", ""),
			Type: getString(nodeMap, "type", ""),
			Name: getString(nodeMap, "name", ""),
		}

		if config, ok := nodeMap["parameters"].(map[string]interface{}); ok {
			node.Config = config
		} else if config, ok := nodeMap["config"].(map[string]interface{}); ok {
			node.Config = config
		} else {
			node.Config = make(map[string]interface{})
		}

		if pos, ok := nodeMap["position"].(map[string]interface{}); ok {
			node.Position = Position{
				X: getFloat(pos, "x", 0),
				Y: getFloat(pos, "y", 0),
			}
		}

		node.Disabled = getBool(nodeMap, "disabled", false)
		node.RetryOnFail = getBool(nodeMap, "retryOnFail", false)
		node.MaxRetries = getInt(nodeMap, "maxRetries", 0)

		if timeout := getInt(nodeMap, "timeout", 0); timeout > 0 {
			node.Timeout = time.Duration(timeout) * time.Millisecond
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func parseConnections(data models.JSON) ([]*Connection, error) {
	if data == nil {
		return nil, nil
	}

	connsRaw, ok := data["connections"].([]interface{})
	if !ok {
		return nil, nil
	}

	connections := make([]*Connection, 0, len(connsRaw))
	for _, connRaw := range connsRaw {
		connMap, ok := connRaw.(map[string]interface{})
		if !ok {
			continue
		}

		conn := &Connection{
			ID:           getString(connMap, "id", ""),
			SourceNodeID: getString(connMap, "source_node_id", getString(connMap, "sourceNodeId", "")),
			SourceHandle: getString(connMap, "source_handle", getString(connMap, "sourceHandle", "output")),
			TargetNodeID: getString(connMap, "target_node_id", getString(connMap, "targetNodeId", "")),
			TargetHandle: getString(connMap, "target_handle", getString(connMap, "targetHandle", "input")),
		}

		if conn.SourceNodeID != "" && conn.TargetNodeID != "" {
			connections = append(connections, conn)
		}
	}

	return connections, nil
}

func parseSettings(data models.JSON) WorkflowSettings {
	return WorkflowSettings{
		Timezone:           getString(data, "timezone", "UTC"),
		ErrorWorkflow:      getString(data, "errorWorkflow", ""),
		CallerPolicy:       getString(data, "callerPolicy", "workflowsFromSameOwner"),
		SaveExecutionData:  getBool(data, "saveExecutionData", true),
		SaveSuccessfulData: getBool(data, "saveSuccessfulData", true),
	}
}

func getString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}

func getInt(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return defaultVal
}

func getFloat(m map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return defaultVal
}

func getBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultVal
}
