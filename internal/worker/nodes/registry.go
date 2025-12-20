package nodes

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

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

type Node interface {
	Type() string
	Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error)
}

type Registry struct {
	nodes map[string]Node
}

// NewRegistry creates a basic registry with stub nodes (for backward compatibility)
// For production use, prefer NewNodeFactory().CreateRegistry() for full implementations
func NewRegistry() *Registry {
	return NewNodeFactory(nil).CreateRegistry()
}

// NewRegistryWithDeps creates a registry with full node implementations and dependencies
func NewRegistryWithDeps(deps *Dependencies) *Registry {
	return NewNodeFactory(deps).CreateRegistry()
}

func (r *Registry) Register(node Node) {
	r.nodes[node.Type()] = node
}

func (r *Registry) Get(nodeType string) Node {
	return r.nodes[nodeType]
}

func (r *Registry) List() []string {
	types := make([]string, 0, len(r.nodes))
	for t := range r.nodes {
		types = append(types, t)
	}
	return types
}
