package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
	"github.com/linkflow-ai/linkflow/internal/worker/events"
	"github.com/rs/zerolog/log"
)

// Middleware interface for node execution middleware
type Middleware interface {
	Execute(ctx context.Context, rctx *RuntimeContext, node *NodeDefinition, next func(ctx context.Context) (*NodeResult, error)) (map[string]interface{}, error)
}

// MiddlewareChain manages middleware execution
type MiddlewareChain interface {
	Execute(ctx context.Context, rctx *RuntimeContext, node *NodeDefinition, handler func(ctx context.Context) (*NodeResult, error)) (map[string]interface{}, error)
}

// Processor is the core workflow execution engine
type Processor struct {
	middleware MiddlewareChain
	cache      Cache
	metrics    MetricsCollector
}

// Cache interface for result caching
type Cache interface {
	Get(ctx context.Context, key string) (map[string]interface{}, bool)
	Set(ctx context.Context, key string, value map[string]interface{}) error
}

// MetricsCollector interface for metrics
type MetricsCollector interface {
	RecordNodeExecution(workspaceID, nodeType string, duration time.Duration, err error)
	RecordWorkflowExecution(workspaceID string, duration time.Duration, nodesCount int, err error)
}

// Config configures the processor
type Config struct {
	Middleware MiddlewareChain
	Cache      Cache
	Metrics    MetricsCollector
}

// New creates a new processor
func New(cfg Config) *Processor {
	return &Processor{
		middleware: cfg.Middleware,
		cache:      cfg.Cache,
		metrics:    cfg.Metrics,
	}
}

// Execute runs a workflow with the given input
func (p *Processor) Execute(
	ctx context.Context,
	workflow *WorkflowDefinition,
	input Input,
	opts ExecutionOptions,
	executionID uuid.UUID,
	getCredential func(uuid.UUID) (*models.CredentialData, error),
	publisher *events.Publisher,
) (*Result, error) {
	startTime := time.Now()

	// Create runtime context
	rctx := NewRuntimeContext(ctx, executionID, workflow.ID, workflow.WorkspaceID, input, getCredential, publisher)

	// Apply workflow timeout
	if opts.WorkflowTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.WorkflowTimeout)
		defer cancel()
	}

	// Build DAG
	dag := BuildDAG(workflow)
	rctx.SetTotalNodes(dag.NodeCount())

	// Validate DAG
	if validationErrors := dag.Validate(); len(validationErrors) > 0 {
		return &Result{
			ExecutionID: executionID,
			Status:      StatusFailed,
			Error:       fmt.Sprintf("validation failed: %s", validationErrors[0].Message),
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			Duration:    time.Since(startTime),
		}, nil
	}

	// Execute workflow
	var execErr error
	if opts.MaxParallelNodes > 1 {
		execErr = p.executeParallel(ctx, rctx, dag, opts)
	} else {
		execErr = p.executeSequential(ctx, rctx, dag, opts)
	}

	// Build result
	result := &Result{
		ExecutionID:   executionID,
		StartedAt:     startTime,
		CompletedAt:   time.Now(),
		Duration:      time.Since(startTime),
		NodesExecuted: int(rctx.completedNodes.Load()),
		NodeResults:   make(map[string]*NodeResult),
		Output:        rctx.GetAllNodeOutputs(),
	}

	if execErr != nil {
		result.Status = StatusFailed
		result.Error = execErr.Error()
		if err, nodeID := rctx.GetError(); err != nil {
			result.ErrorNodeID = nodeID
		}
	} else if rctx.IsCancelled() {
		result.Status = StatusCancelled
	} else {
		result.Status = StatusCompleted
	}

	// Record metrics
	if p.metrics != nil {
		p.metrics.RecordWorkflowExecution(workflow.WorkspaceID.String(), result.Duration, result.NodesExecuted, execErr)
	}

	return result, nil
}

// executeSequential executes nodes one at a time in topological order
func (p *Processor) executeSequential(ctx context.Context, rctx *RuntimeContext, dag *DAG, opts ExecutionOptions) error {
	order, err := dag.TopologicalSort()
	if err != nil {
		return err
	}

	for _, nodeID := range order {
		if rctx.IsCancelled() {
			return fmt.Errorf("execution cancelled")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		node := dag.GetNode(nodeID)
		if node == nil {
			continue
		}

		if err := p.executeNode(ctx, rctx, node, opts); err != nil {
			rctx.SetError(err, nodeID)
			return err
		}
	}

	return nil
}

// executeParallel executes nodes in parallel where possible
func (p *Processor) executeParallel(ctx context.Context, rctx *RuntimeContext, dag *DAG, opts ExecutionOptions) error {
	levels, err := dag.GetLevels()
	if err != nil {
		return err
	}

	semaphore := make(chan struct{}, opts.MaxParallelNodes)

	for _, level := range levels {
		if rctx.IsCancelled() {
			return fmt.Errorf("execution cancelled")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute all nodes in this level in parallel
		var wg sync.WaitGroup
		errChan := make(chan error, len(level))

		for _, nodeID := range level {
			node := dag.GetNode(nodeID)
			if node == nil {
				continue
			}

			wg.Add(1)
			semaphore <- struct{}{} // Acquire

			go func(n *NodeDefinition) {
				defer wg.Done()
				defer func() { <-semaphore }() // Release

				if err := p.executeNode(ctx, rctx, n, opts); err != nil {
					rctx.SetError(err, n.ID)
					errChan <- err
				}
			}(node)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// executeNode executes a single node
func (p *Processor) executeNode(ctx context.Context, rctx *RuntimeContext, node *NodeDefinition, opts ExecutionOptions) error {
	// Skip disabled nodes
	if node.Disabled {
		return nil
	}

	// Check skip list
	for _, skipID := range opts.SkipNodes {
		if skipID == node.ID {
			return nil
		}
	}

	log.Debug().
		Str("execution_id", rctx.ExecutionID.String()).
		Str("node_id", node.ID).
		Str("node_type", node.Type).
		Msg("Executing node")

	rctx.SetCurrentNode(node)
	rctx.PublishNodeStarted(node)

	startTime := time.Now()

	// Prepare input
	nodeInput := rctx.PrepareNodeInput(node)

	// Resolve expressions in config
	resolvedConfig, err := rctx.ResolveConfig(node.Config)
	if err != nil {
		rctx.PublishNodeFailed(node, err.Error())
		return fmt.Errorf("failed to resolve config: %w", err)
	}

	// Apply overrides
	if overrides, ok := opts.NodeOverrides[node.ID]; ok {
		for k, v := range overrides {
			resolvedConfig[k] = v
		}
	}

	// Check cache
	if opts.EnableCaching && p.cache != nil && isCacheable(node.Type) {
		cacheKey := fmt.Sprintf("%s:%s:%s", rctx.ExecutionID, node.ID, rctx.ComputeInputHash(node.ID, nodeInput))
		if cached, ok := p.cache.Get(ctx, cacheKey); ok {
			rctx.SetNodeOutput(node.ID, cached)
			durationMs := int(time.Since(startTime).Milliseconds())
			rctx.PublishNodeCompleted(node, durationMs, cached)
			return nil
		}
	}

	// Get node handler
	handler := core.Get(node.Type)
	if handler == nil {
		err := fmt.Errorf("unknown node type: %s", node.Type)
		rctx.PublishNodeFailed(node, err.Error())
		return err
	}

	// Build execution context for node
	nodeExecCtx := &core.ExecutionContext{
		ExecutionID:   rctx.ExecutionID,
		WorkflowID:    rctx.WorkflowID,
		WorkspaceID:   rctx.WorkspaceID,
		NodeID:        node.ID,
		Input:         nodeInput,
		Config:        resolvedConfig,
		Variables:     rctx.Variables,
		GetCredential: rctx.GetCredential,
	}

	// Apply node timeout
	nodeCtx := ctx
	timeout := opts.DefaultNodeTimeout
	if node.Timeout > 0 {
		timeout = node.Timeout
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		nodeCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Execute with middleware chain
	var output map[string]interface{}
	var execErr error

	if p.middleware != nil {
		output, execErr = p.middleware.Execute(nodeCtx, rctx, node, func(ctx context.Context) (*NodeResult, error) {
			out, err := handler.Execute(ctx, nodeExecCtx)
			if err != nil {
				return nil, err
			}
			return &NodeResult{Output: out}, nil
		})
	} else {
		output, execErr = handler.Execute(nodeCtx, nodeExecCtx)
	}

	durationMs := int(time.Since(startTime).Milliseconds())

	// Record metrics
	if p.metrics != nil {
		p.metrics.RecordNodeExecution(rctx.WorkspaceID.String(), node.Type, time.Since(startTime), execErr)
	}

	// Handle retry on fail
	if execErr != nil && node.RetryOnFail && node.MaxRetries > 0 {
		for retry := 1; retry <= node.MaxRetries; retry++ {
			log.Debug().
				Str("node_id", node.ID).
				Int("retry", retry).
				Msg("Retrying node execution")

			// Exponential backoff
			time.Sleep(time.Duration(retry*retry) * 100 * time.Millisecond)

			output, execErr = handler.Execute(nodeCtx, nodeExecCtx)
			if execErr == nil {
				break
			}
		}
	}

	if execErr != nil {
		rctx.PublishNodeFailed(node, execErr.Error())
		return execErr
	}

	// Store output
	rctx.SetNodeOutput(node.ID, output)

	// Cache result
	if opts.EnableCaching && p.cache != nil && isCacheable(node.Type) {
		cacheKey := fmt.Sprintf("%s:%s:%s", rctx.ExecutionID, node.ID, rctx.ComputeInputHash(node.ID, nodeInput))
		_ = p.cache.Set(ctx, cacheKey, output)
	}

	rctx.PublishNodeCompleted(node, durationMs, output)
	return nil
}

// Preview performs a dry-run validation of the workflow
func (p *Processor) Preview(ctx context.Context, workflow *WorkflowDefinition, input Input) (*PreviewResult, error) {
	dag := BuildDAG(workflow)

	result := &PreviewResult{
		Valid:        true,
		NodePreviews: make(map[string]*NodePreview),
	}

	// Validate DAG
	validationErrors := dag.Validate()
	if len(validationErrors) > 0 {
		result.Valid = false
		result.Errors = validationErrors
	}

	// Get execution order
	order, err := dag.TopologicalSort()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Message: err.Error(),
			Code:    "INVALID_DAG",
		})
		return result, nil
	}
	result.ExecutionOrder = order

	// Build previews for each node
	for _, nodeID := range order {
		node := dag.GetNode(nodeID)
		if node == nil {
			continue
		}

		preview := &NodePreview{
			NodeID:       nodeID,
			WouldExecute: !node.Disabled,
			DependsOn:    dag.GetPredecessors(nodeID),
		}

		// Check if node type exists
		if core.Get(node.Type) == nil {
			result.Errors = append(result.Errors, ValidationError{
				NodeID:  nodeID,
				Message: fmt.Sprintf("Unknown node type: %s", node.Type),
				Code:    "UNKNOWN_NODE_TYPE",
			})
			result.Valid = false
		}

		result.NodePreviews[nodeID] = preview
	}

	return result, nil
}

// isCacheable determines if a node type is safe to cache
func isCacheable(nodeType string) bool {
	nonCacheable := map[string]bool{
		"action.http":           true,
		"action.sub_workflow":   true,
		"action.execute_workflow": true,
		"integration.slack":     true,
		"integration.email":     true,
		"integration.discord":   true,
		"integration.telegram":  true,
		"logic.wait":            true,
	}
	return !nonCacheable[nodeType]
}
