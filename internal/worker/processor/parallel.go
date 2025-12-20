package processor

import (
	"context"
	"sync"
)

// ParallelExecutor handles parallel execution of workflow branches
type ParallelExecutor struct {
	maxConcurrency int
	semaphore      chan struct{}
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(maxConcurrency int) *ParallelExecutor {
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}
	return &ParallelExecutor{
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

// ExecuteFunc is a function that executes a unit of work
type ExecuteFunc func(ctx context.Context) error

// ExecuteAll executes all functions in parallel with limited concurrency
func (pe *ParallelExecutor) ExecuteAll(ctx context.Context, funcs []ExecuteFunc) []error {
	if len(funcs) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errors := make([]error, len(funcs))

	for i, fn := range funcs {
		wg.Add(1)

		select {
		case pe.semaphore <- struct{}{}:
			go func(idx int, f ExecuteFunc) {
				defer wg.Done()
				defer func() { <-pe.semaphore }()

				select {
				case <-ctx.Done():
					errors[idx] = ctx.Err()
				default:
					errors[idx] = f(ctx)
				}
			}(i, fn)

		case <-ctx.Done():
			wg.Done()
			errors[i] = ctx.Err()
		}
	}

	wg.Wait()
	return errors
}

// ExecuteBranches executes parallel branches and waits for all to complete
func (pe *ParallelExecutor) ExecuteBranches(ctx context.Context, rctx *RuntimeContext, branches []Branch, executeNode func(ctx context.Context, rctx *RuntimeContext, nodeID string) error) ([]BranchResult, error) {
	if len(branches) == 0 {
		return nil, nil
	}

	results := make([]BranchResult, len(branches))
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for i, branch := range branches {
		wg.Add(1)

		select {
		case pe.semaphore <- struct{}{}:
			go func(idx int, b Branch) {
				defer wg.Done()
				defer func() { <-pe.semaphore }()

				result := BranchResult{
					BranchID: b.ID,
					Results:  make(map[string]*NodeResult),
				}

				for _, nodeID := range b.Nodes {
					select {
					case <-ctx.Done():
						result.Error = ctx.Err()
						results[idx] = result
						return
					default:
					}

					if err := executeNode(ctx, rctx, nodeID); err != nil {
						result.Error = err
						errMu.Lock()
						if firstErr == nil {
							firstErr = err
						}
						errMu.Unlock()
						break
					}
				}

				results[idx] = result
			}(i, branch)

		case <-ctx.Done():
			wg.Done()
			results[i] = BranchResult{
				BranchID: branch.ID,
				Error:    ctx.Err(),
			}
		}
	}

	wg.Wait()
	return results, firstErr
}

// ExecuteLevel executes all nodes at the same DAG level in parallel
func (pe *ParallelExecutor) ExecuteLevel(ctx context.Context, rctx *RuntimeContext, dag *DAG, nodeIDs []string, executeNode func(ctx context.Context, rctx *RuntimeContext, node *NodeDefinition) error) error {
	if len(nodeIDs) == 0 {
		return nil
	}

	// Single node - execute directly
	if len(nodeIDs) == 1 {
		node := dag.GetNode(nodeIDs[0])
		if node == nil {
			return nil
		}
		return executeNode(ctx, rctx, node)
	}

	// Multiple nodes - execute in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodeIDs))

	for _, nodeID := range nodeIDs {
		node := dag.GetNode(nodeID)
		if node == nil {
			continue
		}

		wg.Add(1)

		select {
		case pe.semaphore <- struct{}{}:
			go func(n *NodeDefinition) {
				defer wg.Done()
				defer func() { <-pe.semaphore }()

				if err := executeNode(ctx, rctx, n); err != nil {
					errChan <- err
				}
			}(node)

		case <-ctx.Done():
			wg.Done()
			errChan <- ctx.Err()
		}
	}

	wg.Wait()
	close(errChan)

	// Return first error
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// WorkerPool manages a pool of workers for parallel execution
type WorkerPool struct {
	workers    int
	taskQueue  chan Task
	resultChan chan TaskResult
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// Task represents a unit of work
type Task struct {
	ID      string
	Execute func(ctx context.Context) (interface{}, error)
}

// TaskResult represents the result of a task
type TaskResult struct {
	TaskID string
	Result interface{}
	Error  error
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		workers:    workers,
		taskQueue:  make(chan Task, workers*2),
		resultChan: make(chan TaskResult, workers*2),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start workers
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				return
			}

			result, err := task.Execute(wp.ctx)
			wp.resultChan <- TaskResult{
				TaskID: task.ID,
				Result: result,
				Error:  err,
			}
		}
	}
}

// Submit submits a task to the pool
func (wp *WorkerPool) Submit(task Task) {
	select {
	case wp.taskQueue <- task:
	case <-wp.ctx.Done():
	}
}

// Results returns the result channel
func (wp *WorkerPool) Results() <-chan TaskResult {
	return wp.resultChan
}

// Shutdown gracefully shuts down the pool
func (wp *WorkerPool) Shutdown() {
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
	close(wp.resultChan)
}

// MergeStrategy defines how to merge results from parallel branches
type MergeStrategy string

const (
	MergeStrategyWaitAll     MergeStrategy = "wait_all"     // Wait for all branches
	MergeStrategyWaitFirst   MergeStrategy = "wait_first"   // Return first completed
	MergeStrategyWaitAny     MergeStrategy = "wait_any"     // Return first successful
	MergeStrategyCombine     MergeStrategy = "combine"      // Combine all results
	MergeStrategyMergeArrays MergeStrategy = "merge_arrays" // Merge array results
)

// MergeBranchResults merges results from parallel branches based on strategy
func MergeBranchResults(results []BranchResult, strategy MergeStrategy) (map[string]interface{}, error) {
	switch strategy {
	case MergeStrategyWaitAll:
		merged := make(map[string]interface{})
		for _, result := range results {
			if result.Error != nil {
				return nil, result.Error
			}
			for k, v := range result.Results {
				merged[k] = v
			}
		}
		return merged, nil

	case MergeStrategyWaitFirst:
		for _, result := range results {
			if result.Error == nil && len(result.Results) > 0 {
				return map[string]interface{}{"branch": result.BranchID, "results": result.Results}, nil
			}
		}
		return nil, nil

	case MergeStrategyWaitAny:
		for _, result := range results {
			if result.Error == nil {
				return map[string]interface{}{"branch": result.BranchID, "results": result.Results}, nil
			}
		}
		// All failed, return first error
		for _, result := range results {
			if result.Error != nil {
				return nil, result.Error
			}
		}
		return nil, nil

	case MergeStrategyCombine:
		merged := make(map[string]interface{})
		branchResults := make([]map[string]interface{}, 0, len(results))
		for _, result := range results {
			branchData := map[string]interface{}{
				"branchId": result.BranchID,
				"results":  result.Results,
			}
			if result.Error != nil {
				branchData["error"] = result.Error.Error()
			}
			branchResults = append(branchResults, branchData)
		}
		merged["branches"] = branchResults
		return merged, nil

	case MergeStrategyMergeArrays:
		var merged []interface{}
		for _, result := range results {
			if result.Error != nil {
				continue
			}
			for _, nodeResult := range result.Results {
				if nodeResult != nil && nodeResult.Output != nil {
					if arr, ok := nodeResult.Output["items"].([]interface{}); ok {
						merged = append(merged, arr...)
					}
				}
			}
		}
		return map[string]interface{}{"items": merged}, nil

	default:
		return MergeBranchResults(results, MergeStrategyWaitAll)
	}
}
