package nodes

import (
	"fmt"
	"sync"
)

// Registry manages node types and their factories
type Registry struct {
	nodes    map[string]NodeFactory
	metadata map[string]NodeMetadata
	mu       sync.RWMutex
}

// NewRegistry creates a new node registry
func NewRegistry() *Registry {
	return &Registry{
		nodes:    make(map[string]NodeFactory),
		metadata: make(map[string]NodeMetadata),
	}
}

// Register registers a node type with its factory
func (r *Registry) Register(nodeType string, factory NodeFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeType]; exists {
		return fmt.Errorf("node type %s is already registered", nodeType)
	}

	r.nodes[nodeType] = factory

	// Get metadata from a sample instance
	instance := factory()
	r.metadata[nodeType] = instance.Metadata()

	return nil
}

// Get returns a new instance of the specified node type
func (r *Registry) Get(nodeType string) (Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.nodes[nodeType]
	if !exists {
		return nil, fmt.Errorf("unknown node type: %s", nodeType)
	}

	return factory(), nil
}

// GetMetadata returns metadata for a node type
func (r *Registry) GetMetadata(nodeType string) (NodeMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meta, exists := r.metadata[nodeType]
	if !exists {
		return NodeMetadata{}, fmt.Errorf("unknown node type: %s", nodeType)
	}

	return meta, nil
}

// ListTypes returns all registered node types
func (r *Registry) ListTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.nodes))
	for t := range r.nodes {
		types = append(types, t)
	}
	return types
}

// ListMetadata returns metadata for all registered nodes
func (r *Registry) ListMetadata() []NodeMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make([]NodeMetadata, 0, len(r.metadata))
	for _, m := range r.metadata {
		metadata = append(metadata, m)
	}
	return metadata
}

// ListByCategory returns node types grouped by category
func (r *Registry) ListByCategory() map[string][]NodeMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make(map[string][]NodeMetadata)
	for _, m := range r.metadata {
		categories[m.Category] = append(categories[m.Category], m)
	}
	return categories
}

// Has checks if a node type is registered
func (r *Registry) Has(nodeType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.nodes[nodeType]
	return exists
}

// Unregister removes a node type from the registry
func (r *Registry) Unregister(nodeType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, nodeType)
	delete(r.metadata, nodeType)
}

// DefaultRegistry is the global node registry
var DefaultRegistry = NewRegistry()

// Register registers a node in the default registry
func Register(nodeType string, factory NodeFactory) error {
	return DefaultRegistry.Register(nodeType, factory)
}

// Get gets a node from the default registry
func Get(nodeType string) (Node, error) {
	return DefaultRegistry.Get(nodeType)
}
