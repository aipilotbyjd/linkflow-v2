package processor

import (
	"errors"
	"fmt"
	"sort"
)

// DAG represents a directed acyclic graph of workflow nodes
type DAG struct {
	Nodes      map[string]*NodeDefinition
	Edges      map[string][]string // node -> outgoing nodes
	InDegree   map[string]int      // node -> incoming edge count
	Levels     [][]string          // nodes grouped by execution level
	RootNodes  []string
	LeafNodes  []string
}

// BuildDAG constructs a DAG from workflow definition
func BuildDAG(workflow *WorkflowDefinition) *DAG {
	dag := &DAG{
		Nodes:    make(map[string]*NodeDefinition),
		Edges:    make(map[string][]string),
		InDegree: make(map[string]int),
	}

	// Add all nodes
	for _, node := range workflow.Nodes {
		if node.Disabled {
			continue
		}
		dag.Nodes[node.ID] = node
		dag.Edges[node.ID] = []string{}
		dag.InDegree[node.ID] = 0
	}

	// Add edges from connections
	for _, conn := range workflow.Connections {
		// Skip if either node is disabled/missing
		if _, ok := dag.Nodes[conn.SourceNodeID]; !ok {
			continue
		}
		if _, ok := dag.Nodes[conn.TargetNodeID]; !ok {
			continue
		}

		dag.Edges[conn.SourceNodeID] = append(dag.Edges[conn.SourceNodeID], conn.TargetNodeID)
		dag.InDegree[conn.TargetNodeID]++

		// Add input reference to target node
		if targetNode, ok := dag.Nodes[conn.TargetNodeID]; ok {
			targetNode.Inputs = append(targetNode.Inputs, ConnectionRef{
				SourceNodeID: conn.SourceNodeID,
				SourceHandle: conn.SourceHandle,
			})
		}
	}

	// Compute root and leaf nodes
	dag.computeRootAndLeafNodes()

	return dag
}

func (d *DAG) computeRootAndLeafNodes() {
	d.RootNodes = nil
	d.LeafNodes = nil

	for nodeID, degree := range d.InDegree {
		if degree == 0 {
			d.RootNodes = append(d.RootNodes, nodeID)
		}
	}

	for nodeID, edges := range d.Edges {
		if len(edges) == 0 {
			d.LeafNodes = append(d.LeafNodes, nodeID)
		}
	}

	// Sort for deterministic order
	sort.Strings(d.RootNodes)
	sort.Strings(d.LeafNodes)
}

// TopologicalSort returns nodes in execution order
func (d *DAG) TopologicalSort() ([]string, error) {
	var result []string
	queue := make([]string, 0, len(d.RootNodes))

	// Copy in-degree map
	inDegree := make(map[string]int)
	for k, v := range d.InDegree {
		inDegree[k] = v
	}

	// Start with root nodes
	for _, nodeID := range d.RootNodes {
		queue = append(queue, nodeID)
	}

	for len(queue) > 0 {
		// Sort queue for deterministic order
		sort.Strings(queue)

		// Dequeue first
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		// Process outgoing edges
		for _, targetID := range d.Edges[nodeID] {
			inDegree[targetID]--
			if inDegree[targetID] == 0 {
				queue = append(queue, targetID)
			}
		}
	}

	// Check for cycles
	if len(result) != len(d.Nodes) {
		return nil, errors.New("cycle detected in workflow")
	}

	return result, nil
}

// GetLevels returns nodes grouped by execution level for parallel execution
func (d *DAG) GetLevels() ([][]string, error) {
	if d.Levels != nil {
		return d.Levels, nil
	}

	levels, err := d.computeLevels()
	if err != nil {
		return nil, err
	}
	d.Levels = levels
	return d.Levels, nil
}

func (d *DAG) computeLevels() ([][]string, error) {
	var levels [][]string

	// Copy in-degree map
	inDegree := make(map[string]int)
	for k, v := range d.InDegree {
		inDegree[k] = v
	}

	remaining := len(d.Nodes)

	for remaining > 0 {
		var level []string

		// Find all nodes with no remaining dependencies
		for nodeID, degree := range inDegree {
			if degree == 0 {
				level = append(level, nodeID)
			}
		}

		if len(level) == 0 && remaining > 0 {
			return nil, errors.New("cycle detected in workflow")
		}

		// Sort for deterministic order
		sort.Strings(level)

		// Remove processed nodes and update dependencies
		for _, nodeID := range level {
			delete(inDegree, nodeID)
			for _, target := range d.Edges[nodeID] {
				inDegree[target]--
			}
			remaining--
		}

		levels = append(levels, level)
	}

	return levels, nil
}

// GetParallelizable returns nodes that can run in parallel given current completed set
func (d *DAG) GetParallelizable(completed map[string]bool) []string {
	var ready []string

	for nodeID := range d.Nodes {
		// Skip if already completed
		if completed[nodeID] {
			continue
		}

		// Check if all dependencies are completed
		allDepsCompleted := true
		for _, dep := range d.GetPredecessors(nodeID) {
			if !completed[dep] {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			ready = append(ready, nodeID)
		}
	}

	sort.Strings(ready)
	return ready
}

// GetPredecessors returns nodes that feed into this node
func (d *DAG) GetPredecessors(nodeID string) []string {
	var predecessors []string
	for srcID, targets := range d.Edges {
		for _, targetID := range targets {
			if targetID == nodeID {
				predecessors = append(predecessors, srcID)
			}
		}
	}
	sort.Strings(predecessors)
	return predecessors
}

// GetSuccessors returns nodes that this node feeds into
func (d *DAG) GetSuccessors(nodeID string) []string {
	successors := d.Edges[nodeID]
	result := make([]string, len(successors))
	copy(result, successors)
	sort.Strings(result)
	return result
}

// GetNode returns a node by ID
func (d *DAG) GetNode(nodeID string) *NodeDefinition {
	return d.Nodes[nodeID]
}

// NodeCount returns number of nodes in the DAG
func (d *DAG) NodeCount() int {
	return len(d.Nodes)
}

// Validate checks if the DAG is valid
func (d *DAG) Validate() []ValidationError {
	var errors []ValidationError

	// Check for cycles
	if _, err := d.TopologicalSort(); err != nil {
		errors = append(errors, ValidationError{
			Message: "Workflow contains a cycle",
			Code:    "CYCLE_DETECTED",
		})
	}

	// Check for disconnected nodes
	if len(d.Nodes) > 1 {
		reachable := d.getReachableNodes()
		for nodeID := range d.Nodes {
			if !reachable[nodeID] {
				errors = append(errors, ValidationError{
					NodeID:  nodeID,
					Message: "Node is not reachable from any trigger",
					Code:    "UNREACHABLE_NODE",
				})
			}
		}
	}

	// Check for nodes with no type
	for nodeID, node := range d.Nodes {
		if node.Type == "" {
			errors = append(errors, ValidationError{
				NodeID:  nodeID,
				Message: "Node has no type defined",
				Code:    "MISSING_TYPE",
			})
		}
	}

	return errors
}

func (d *DAG) getReachableNodes() map[string]bool {
	reachable := make(map[string]bool)

	// BFS from root nodes
	queue := make([]string, len(d.RootNodes))
	copy(queue, d.RootNodes)

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		if reachable[nodeID] {
			continue
		}
		reachable[nodeID] = true

		for _, successor := range d.GetSuccessors(nodeID) {
			if !reachable[successor] {
				queue = append(queue, successor)
			}
		}
	}

	return reachable
}

// GetBranches identifies parallel branches in the DAG
func (d *DAG) GetBranches() []Branch {
	var branches []Branch

	// Find branching points (nodes with multiple successors)
	for nodeID, successors := range d.Edges {
		if len(successors) > 1 {
			for i, successor := range successors {
				branch := Branch{
					ID:    fmt.Sprintf("%s_branch_%d", nodeID, i),
					Nodes: d.collectBranchNodes(successor, nodeID),
				}
				branches = append(branches, branch)
			}
		}
	}

	return branches
}

func (d *DAG) collectBranchNodes(startNode, branchPoint string) []string {
	var nodes []string
	visited := make(map[string]bool)
	queue := []string{startNode}

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		if visited[nodeID] {
			continue
		}

		// Stop if we reach a merge point (node with multiple predecessors)
		if len(d.GetPredecessors(nodeID)) > 1 && nodeID != startNode {
			continue
		}

		visited[nodeID] = true
		nodes = append(nodes, nodeID)

		for _, successor := range d.GetSuccessors(nodeID) {
			if !visited[successor] {
				queue = append(queue, successor)
			}
		}
	}

	return nodes
}

// SubDAG creates a sub-DAG starting from a specific node
func (d *DAG) SubDAG(startNodeID string) *DAG {
	subDAG := &DAG{
		Nodes:    make(map[string]*NodeDefinition),
		Edges:    make(map[string][]string),
		InDegree: make(map[string]int),
	}

	// BFS to collect all reachable nodes
	visited := make(map[string]bool)
	queue := []string{startNodeID}

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		if visited[nodeID] {
			continue
		}
		visited[nodeID] = true

		if node, ok := d.Nodes[nodeID]; ok {
			subDAG.Nodes[nodeID] = node
			subDAG.Edges[nodeID] = d.Edges[nodeID]
			subDAG.InDegree[nodeID] = 0 // Will recalculate
		}

		for _, successor := range d.GetSuccessors(nodeID) {
			if !visited[successor] {
				queue = append(queue, successor)
			}
		}
	}

	// Recalculate in-degrees for sub-DAG
	for nodeID := range subDAG.Nodes {
		subDAG.InDegree[nodeID] = 0
	}
	for _, targets := range subDAG.Edges {
		for _, target := range targets {
			if _, ok := subDAG.Nodes[target]; ok {
				subDAG.InDegree[target]++
			}
		}
	}

	subDAG.computeRootAndLeafNodes()
	return subDAG
}


