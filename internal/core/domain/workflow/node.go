package workflow

import "encoding/json"

// Node represents a workflow node
type Node struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Position    Position               `json:"position"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Credentials []NodeCredential       `json:"credentials,omitempty"`
	Disabled    bool                   `json:"disabled,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
	RetryOnFail bool                   `json:"retry_on_fail,omitempty"`
	MaxRetries  int                    `json:"max_retries,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"` // seconds
}

// Position represents node position on canvas
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeCredential represents a credential reference in a node
type NodeCredential struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// NodeType returns the node type
func (n Node) NodeType() NodeTypeInfo {
	return ParseNodeType(n.Type)
}

// IsTrigger checks if the node is a trigger
func (n Node) IsTrigger() bool {
	return n.NodeType().IsTrigger()
}

// IsAction checks if the node is an action
func (n Node) IsAction() bool {
	return n.NodeType().Category == "action"
}

// IsLogic checks if the node is a logic node
func (n Node) IsLogic() bool {
	return n.NodeType().Category == "logic"
}

// IsIntegration checks if the node is an integration
func (n Node) IsIntegration() bool {
	return n.NodeType().Category == "integration"
}

// NodeTypeInfo holds parsed node type information
type NodeTypeInfo struct {
	Category string
	Name     string
	Full     string
}

// ParseNodeType parses a node type string (e.g., "trigger.webhook")
func ParseNodeType(nodeType string) NodeTypeInfo {
	info := NodeTypeInfo{Full: nodeType}

	for i, c := range nodeType {
		if c == '.' {
			info.Category = nodeType[:i]
			info.Name = nodeType[i+1:]
			return info
		}
	}

	info.Category = nodeType
	return info
}

// IsTrigger checks if the node type is a trigger
func (t NodeTypeInfo) IsTrigger() bool {
	return t.Category == "trigger"
}

// String returns the full node type string
func (t NodeTypeInfo) String() string {
	return t.Full
}

// NodesFromJSON parses nodes from JSON
func NodesFromJSON(data []byte) ([]Node, error) {
	var nodes []Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

// NodesToJSON converts nodes to JSON
func NodesToJSON(nodes []Node) ([]byte, error) {
	return json.Marshal(nodes)
}
