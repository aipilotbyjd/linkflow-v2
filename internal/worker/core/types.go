package core

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/redis/go-redis/v9"
)

// ExecutionContext contains all data needed for node execution
type ExecutionContext struct {
	ExecutionID   uuid.UUID
	WorkflowID    uuid.UUID
	WorkspaceID   uuid.UUID
	NodeID        string
	Input         map[string]interface{}
	Config        map[string]interface{}
	Variables     map[string]interface{}
	Credentials   map[string]interface{}
	GetCredential func(uuid.UUID) (*models.CredentialData, error)
}

// Node is the interface all workflow nodes must implement
type Node interface {
	Type() string
	Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error)
}

// NodeMeta contains metadata about a node for UI and discovery
type NodeMeta struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Icon        string   `json:"icon"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags,omitempty"`
}

// Dependencies holds external dependencies for nodes that need them
type Dependencies struct {
	QueueClient *queue.Client
	RedisClient *redis.Client
}

// NodeWithDeps is for nodes that require dependencies
type NodeWithDeps interface {
	Node
	SetDependencies(deps *Dependencies)
}

// Global registry
var (
	globalRegistry = &Registry{
		nodes: make(map[string]Node),
		meta:  make(map[string]NodeMeta),
	}
	registryMu sync.RWMutex
	globalDeps *Dependencies
)

// Registry holds all registered nodes
type Registry struct {
	nodes map[string]Node
	meta  map[string]NodeMeta
}

// Register adds a node to the global registry (called from init())
func Register(node Node, meta ...NodeMeta) {
	registryMu.Lock()
	defer registryMu.Unlock()

	nodeType := node.Type()
	globalRegistry.nodes[nodeType] = node

	if len(meta) > 0 {
		m := meta[0]
		m.Type = nodeType
		globalRegistry.meta[nodeType] = m
	} else {
		globalRegistry.meta[nodeType] = NodeMeta{
			Type:     nodeType,
			Name:     nodeType,
			Category: getCategoryFromType(nodeType),
		}
	}
}

// SetGlobalDependencies sets dependencies for nodes that need them
func SetGlobalDependencies(deps *Dependencies) {
	registryMu.Lock()
	defer registryMu.Unlock()

	globalDeps = deps

	// Inject dependencies into nodes that need them
	for _, node := range globalRegistry.nodes {
		if nodeWithDeps, ok := node.(NodeWithDeps); ok {
			nodeWithDeps.SetDependencies(deps)
		}
	}
}

// GetGlobalDependencies returns the global dependencies
func GetGlobalDependencies() *Dependencies {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return globalDeps
}

// Get returns a node by type from global registry
func Get(nodeType string) Node {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return globalRegistry.nodes[nodeType]
}

// GetMeta returns metadata for a node type
func GetMeta(nodeType string) (NodeMeta, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	meta, ok := globalRegistry.meta[nodeType]
	return meta, ok
}

// List returns all registered node types
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	types := make([]string, 0, len(globalRegistry.nodes))
	for t := range globalRegistry.nodes {
		types = append(types, t)
	}
	return types
}

// ListByCategory returns nodes filtered by category
func ListByCategory(category string) []NodeMeta {
	registryMu.RLock()
	defer registryMu.RUnlock()

	var result []NodeMeta
	for _, meta := range globalRegistry.meta {
		if meta.Category == category {
			result = append(result, meta)
		}
	}
	return result
}

// ListAll returns all node metadata
func ListAll() []NodeMeta {
	registryMu.RLock()
	defer registryMu.RUnlock()

	result := make([]NodeMeta, 0, len(globalRegistry.meta))
	for _, meta := range globalRegistry.meta {
		result = append(result, meta)
	}
	return result
}

// Count returns total number of registered nodes
func Count() int {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return len(globalRegistry.nodes)
}

// getCategoryFromType extracts category from node type string
func getCategoryFromType(nodeType string) string {
	for i, c := range nodeType {
		if c == '.' {
			return nodeType[:i]
		}
	}
	return "other"
}
