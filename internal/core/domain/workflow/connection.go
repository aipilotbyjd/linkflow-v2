package workflow

import "encoding/json"

// Connection represents a connection between two nodes
type Connection struct {
	ID         string `json:"id,omitempty"`
	SourceNode string `json:"source_node"`
	SourcePort string `json:"source_port,omitempty"`
	TargetNode string `json:"target_node"`
	TargetPort string `json:"target_port,omitempty"`
	Condition  string `json:"condition,omitempty"` // For conditional connections
}

// NewConnection creates a new connection
func NewConnection(sourceNode, targetNode string) Connection {
	return Connection{
		SourceNode: sourceNode,
		SourcePort: "main",
		TargetNode: targetNode,
		TargetPort: "main",
	}
}

// NewConditionalConnection creates a connection with a condition
func NewConditionalConnection(sourceNode, targetNode, condition string) Connection {
	return Connection{
		SourceNode: sourceNode,
		SourcePort: "main",
		TargetNode: targetNode,
		TargetPort: "main",
		Condition:  condition,
	}
}

// WithPorts sets the source and target ports
func (c Connection) WithPorts(sourcePort, targetPort string) Connection {
	c.SourcePort = sourcePort
	c.TargetPort = targetPort
	return c
}

// IsConditional checks if the connection has a condition
func (c Connection) IsConditional() bool {
	return c.Condition != ""
}

// Connects checks if this connection connects the given nodes
func (c Connection) Connects(sourceNode, targetNode string) bool {
	return c.SourceNode == sourceNode && c.TargetNode == targetNode
}

// ConnectionsFromJSON parses connections from JSON
func ConnectionsFromJSON(data []byte) ([]Connection, error) {
	var connections []Connection
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// ConnectionsToJSON converts connections to JSON
func ConnectionsToJSON(connections []Connection) ([]byte, error) {
	return json.Marshal(connections)
}

// ConnectionMap provides efficient lookup of connections
type ConnectionMap struct {
	bySource map[string][]Connection
	byTarget map[string][]Connection
}

// NewConnectionMap creates a new connection map
func NewConnectionMap(connections []Connection) *ConnectionMap {
	cm := &ConnectionMap{
		bySource: make(map[string][]Connection),
		byTarget: make(map[string][]Connection),
	}

	for _, conn := range connections {
		cm.bySource[conn.SourceNode] = append(cm.bySource[conn.SourceNode], conn)
		cm.byTarget[conn.TargetNode] = append(cm.byTarget[conn.TargetNode], conn)
	}

	return cm
}

// GetOutgoing returns all connections from a node
func (cm *ConnectionMap) GetOutgoing(nodeID string) []Connection {
	return cm.bySource[nodeID]
}

// GetIncoming returns all connections to a node
func (cm *ConnectionMap) GetIncoming(nodeID string) []Connection {
	return cm.byTarget[nodeID]
}

// HasIncoming checks if a node has incoming connections
func (cm *ConnectionMap) HasIncoming(nodeID string) bool {
	return len(cm.byTarget[nodeID]) > 0
}

// HasOutgoing checks if a node has outgoing connections
func (cm *ConnectionMap) HasOutgoing(nodeID string) bool {
	return len(cm.bySource[nodeID]) > 0
}
