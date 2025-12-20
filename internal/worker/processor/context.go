package processor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/worker/events"
)

// RuntimeContext holds all state for a workflow execution
type RuntimeContext struct {
	// Identity
	ExecutionID uuid.UUID
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID

	// Data flow
	Input       map[string]interface{}
	Variables   map[string]interface{}
	NodeOutputs sync.Map // Thread-safe for parallel execution

	// Control
	ctx       context.Context
	cancel    context.CancelFunc
	cancelled atomic.Bool
	progress  atomic.Int32 // 0-100

	// Dependencies
	GetCredential func(uuid.UUID) (*models.CredentialData, error)
	publisher     *events.Publisher

	// Expression evaluator
	expression *ExpressionEvaluator

	// Execution tracking
	startedAt       time.Time
	totalNodes      int
	completedNodes  atomic.Int32
	currentNode     atomic.Value // *NodeDefinition
	nodeStartTimes  sync.Map     // nodeID -> time.Time

	// Limits
	MemoryLimit int64
	TimeLimit   time.Duration

	// Tracing
	TraceID string
	SpanID  string

	// Error tracking
	lastError     error
	lastErrorNode string
	mu            sync.RWMutex
}

// NewRuntimeContext creates a new runtime context
func NewRuntimeContext(
	ctx context.Context,
	executionID, workflowID, workspaceID uuid.UUID,
	input map[string]interface{},
	getCredential func(uuid.UUID) (*models.CredentialData, error),
	publisher *events.Publisher,
) *RuntimeContext {
	ctx, cancel := context.WithCancel(ctx)

	rctx := &RuntimeContext{
		ExecutionID:   executionID,
		WorkflowID:    workflowID,
		WorkspaceID:   workspaceID,
		Input:         input,
		Variables:     make(map[string]interface{}),
		ctx:           ctx,
		cancel:        cancel,
		GetCredential: getCredential,
		publisher:     publisher,
		expression:    NewExpressionEvaluator(),
		startedAt:     time.Now(),
		TraceID:       uuid.New().String(),
		SpanID:        uuid.New().String()[:8],
	}

	return rctx
}

// Context returns the underlying context
func (rctx *RuntimeContext) Context() context.Context {
	return rctx.ctx
}

// Cancel cancels the execution
func (rctx *RuntimeContext) Cancel() {
	rctx.cancelled.Store(true)
	rctx.cancel()
}

// IsCancelled checks if execution is cancelled
func (rctx *RuntimeContext) IsCancelled() bool {
	return rctx.cancelled.Load()
}

// SetTotalNodes sets the total number of nodes for progress tracking
func (rctx *RuntimeContext) SetTotalNodes(total int) {
	rctx.totalNodes = total
}

// Progress returns current progress (0-100)
func (rctx *RuntimeContext) Progress() int {
	return int(rctx.progress.Load())
}

// UpdateProgress updates progress based on completed nodes
func (rctx *RuntimeContext) UpdateProgress() {
	if rctx.totalNodes == 0 {
		return
	}
	completed := int(rctx.completedNodes.Load())
	progress := (completed * 100) / rctx.totalNodes
	rctx.progress.Store(int32(progress))
}

// SetNodeOutput stores output for a node (thread-safe)
func (rctx *RuntimeContext) SetNodeOutput(nodeID string, output interface{}) {
	rctx.NodeOutputs.Store(nodeID, output)
	rctx.completedNodes.Add(1)
	rctx.UpdateProgress()
}

// GetNodeOutput retrieves output for a node (thread-safe)
func (rctx *RuntimeContext) GetNodeOutput(nodeID string) (interface{}, bool) {
	return rctx.NodeOutputs.Load(nodeID)
}

// GetAllNodeOutputs returns all node outputs as a map
func (rctx *RuntimeContext) GetAllNodeOutputs() map[string]interface{} {
	outputs := make(map[string]interface{})
	rctx.NodeOutputs.Range(func(key, value interface{}) bool {
		outputs[key.(string)] = value
		return true
	})
	return outputs
}

// SetVariable sets a workflow variable
func (rctx *RuntimeContext) SetVariable(key string, value interface{}) {
	rctx.mu.Lock()
	defer rctx.mu.Unlock()
	rctx.Variables[key] = value
}

// GetVariable gets a workflow variable
func (rctx *RuntimeContext) GetVariable(key string) (interface{}, bool) {
	rctx.mu.RLock()
	defer rctx.mu.RUnlock()
	v, ok := rctx.Variables[key]
	return v, ok
}

// SetCurrentNode sets the currently executing node
func (rctx *RuntimeContext) SetCurrentNode(node *NodeDefinition) {
	rctx.currentNode.Store(node)
	rctx.nodeStartTimes.Store(node.ID, time.Now())
}

// GetCurrentNode returns the currently executing node
func (rctx *RuntimeContext) GetCurrentNode() *NodeDefinition {
	v := rctx.currentNode.Load()
	if v == nil {
		return nil
	}
	return v.(*NodeDefinition)
}

// SetError records an error
func (rctx *RuntimeContext) SetError(err error, nodeID string) {
	rctx.mu.Lock()
	defer rctx.mu.Unlock()
	rctx.lastError = err
	rctx.lastErrorNode = nodeID
}

// GetError returns the last error
func (rctx *RuntimeContext) GetError() (error, string) {
	rctx.mu.RLock()
	defer rctx.mu.RUnlock()
	return rctx.lastError, rctx.lastErrorNode
}

// Duration returns execution duration so far
func (rctx *RuntimeContext) Duration() time.Duration {
	return time.Since(rctx.startedAt)
}

// ResolveExpression evaluates an expression in the current context
func (rctx *RuntimeContext) ResolveExpression(expr string) (interface{}, error) {
	exprCtx := &ExpressionContext{
		Input:       rctx.Input,
		JSON:        rctx.Input["$json"],
		Node:        rctx.GetAllNodeOutputs(),
		Vars:        rctx.Variables,
		Env:         make(map[string]string), // TODO: populate from config
		Now:         time.Now(),
		Today:       time.Now().Format("2006-01-02"),
		Timestamp:   time.Now().Unix(),
		ExecutionID: rctx.ExecutionID.String(),
		WorkflowID:  rctx.WorkflowID.String(),
	}

	return rctx.expression.Evaluate(expr, exprCtx)
}

// ResolveConfig resolves all expressions in a config map
func (rctx *RuntimeContext) ResolveConfig(config map[string]interface{}) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})

	for key, value := range config {
		resolvedValue, err := rctx.resolveValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", key, err)
		}
		resolved[key] = resolvedValue
	}

	return resolved, nil
}

func (rctx *RuntimeContext) resolveValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if strings.Contains(v, "{{") {
			return rctx.ResolveExpression(v)
		}
		return v, nil

	case map[string]interface{}:
		resolved := make(map[string]interface{})
		for k, val := range v {
			resolvedVal, err := rctx.resolveValue(val)
			if err != nil {
				return nil, err
			}
			resolved[k] = resolvedVal
		}
		return resolved, nil

	case []interface{}:
		resolved := make([]interface{}, len(v))
		for i, val := range v {
			resolvedVal, err := rctx.resolveValue(val)
			if err != nil {
				return nil, err
			}
			resolved[i] = resolvedVal
		}
		return resolved, nil

	default:
		return value, nil
	}
}

// PrepareNodeInput prepares input for a node based on connections
func (rctx *RuntimeContext) PrepareNodeInput(node *NodeDefinition) map[string]interface{} {
	input := make(map[string]interface{})

	// Add workflow input
	input["$input"] = rctx.Input
	if json, ok := rctx.Input["$json"]; ok {
		input["$json"] = json
	}

	// Add outputs from connected nodes
	for _, conn := range node.Inputs {
		if output, ok := rctx.GetNodeOutput(conn.SourceNodeID); ok {
			input[conn.SourceNodeID] = output

			// Also set as $json if single input
			if len(node.Inputs) == 1 {
				if outMap, ok := output.(map[string]interface{}); ok {
					input["$json"] = outMap
				}
			}
		}
	}

	// Add all node outputs for $node reference
	input["$node"] = rctx.GetAllNodeOutputs()

	// Add variables
	input["$vars"] = rctx.Variables

	// Add execution metadata
	input["$execution"] = map[string]interface{}{
		"id":          rctx.ExecutionID.String(),
		"workflowId":  rctx.WorkflowID.String(),
		"workspaceId": rctx.WorkspaceID.String(),
		"startedAt":   rctx.startedAt.Format(time.RFC3339),
	}

	return input
}

// ComputeInputHash computes a hash of node input for caching
func (rctx *RuntimeContext) ComputeInputHash(nodeID string, input map[string]interface{}) string {
	data, _ := json.Marshal(input)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8])
}

// PublishEvent publishes an event if publisher is configured
func (rctx *RuntimeContext) PublishEvent(event *events.Event) error {
	if rctx.publisher == nil {
		return nil
	}
	return rctx.publisher.Publish(rctx.ctx, event)
}

// PublishNodeStarted publishes node started event
func (rctx *RuntimeContext) PublishNodeStarted(node *NodeDefinition) {
	if rctx.publisher != nil {
		rctx.publisher.NodeStarted(rctx.ctx, rctx.WorkspaceID, rctx.WorkflowID, rctx.ExecutionID, node.ID, node.Type, node.Name)
	}
}

// PublishNodeCompleted publishes node completed event
func (rctx *RuntimeContext) PublishNodeCompleted(node *NodeDefinition, durationMs int, output interface{}) {
	if rctx.publisher != nil {
		// Truncate output for preview
		preview := truncateOutput(output, 1000)
		rctx.publisher.NodeCompleted(rctx.ctx, rctx.WorkspaceID, rctx.WorkflowID, rctx.ExecutionID, node.ID, durationMs, preview)
	}
}

// PublishNodeFailed publishes node failed event
func (rctx *RuntimeContext) PublishNodeFailed(node *NodeDefinition, errMsg string) {
	if rctx.publisher != nil {
		rctx.publisher.NodeFailed(rctx.ctx, rctx.WorkspaceID, rctx.WorkflowID, rctx.ExecutionID, node.ID, errMsg)
	}
}

func truncateOutput(output interface{}, maxLen int) interface{} {
	data, err := json.Marshal(output)
	if err != nil {
		return nil
	}
	if len(data) <= maxLen {
		return output
	}
	return map[string]interface{}{
		"_truncated": true,
		"_size":      len(data),
		"_preview":   string(data[:maxLen]),
	}
}
