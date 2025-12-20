package executor

import (
	"errors"
)

type NodeData struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Position   Position               `json:"position"`
	Parameters map[string]interface{} `json:"parameters"`
	Inputs     []ConnectionRef        `json:"inputs,omitempty"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Connection struct {
	ID           string `json:"id"`
	SourceNodeID string `json:"source_node_id"`
	SourceHandle string `json:"source_handle"`
	TargetNodeID string `json:"target_node_id"`
	TargetHandle string `json:"target_handle"`
}

type ConnectionRef struct {
	SourceNodeID string `json:"source_node_id"`
	SourceHandle string `json:"source_handle"`
}

type DAG struct {
	Nodes    map[string]*NodeData
	Edges    map[string][]string // node -> outgoing nodes
	InDegree map[string]int      // node -> incoming edge count
}

func BuildDAG(nodes []NodeData, connections []Connection) *DAG {
	dag := &DAG{
		Nodes:    make(map[string]*NodeData),
		Edges:    make(map[string][]string),
		InDegree: make(map[string]int),
	}

	// Add all nodes
	for i := range nodes {
		node := &nodes[i]
		dag.Nodes[node.ID] = node
		dag.Edges[node.ID] = []string{}
		dag.InDegree[node.ID] = 0
	}

	// Add edges from connections
	for _, conn := range connections {
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

	return dag
}

func (d *DAG) TopologicalSort() ([]string, error) {
	var result []string
	queue := []string{}

	// Copy in-degree map
	inDegree := make(map[string]int)
	for k, v := range d.InDegree {
		inDegree[k] = v
	}

	// Find all nodes with no incoming edges
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	for len(queue) > 0 {
		// Dequeue
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		// For each outgoing edge
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

func (d *DAG) GetRootNodes() []string {
	var roots []string
	for nodeID, degree := range d.InDegree {
		if degree == 0 {
			roots = append(roots, nodeID)
		}
	}
	return roots
}

func (d *DAG) GetLeafNodes() []string {
	var leaves []string
	for nodeID, edges := range d.Edges {
		if len(edges) == 0 {
			leaves = append(leaves, nodeID)
		}
	}
	return leaves
}

func (d *DAG) GetPredecessors(nodeID string) []string {
	var predecessors []string
	for srcID, targets := range d.Edges {
		for _, targetID := range targets {
			if targetID == nodeID {
				predecessors = append(predecessors, srcID)
			}
		}
	}
	return predecessors
}

func (d *DAG) GetSuccessors(nodeID string) []string {
	return d.Edges[nodeID]
}
