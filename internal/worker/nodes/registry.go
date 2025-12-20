package nodes

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

type ExecutionContext struct {
	ExecutionID   uuid.UUID
	WorkflowID    uuid.UUID
	NodeID        string
	Input         map[string]interface{}
	Config        map[string]interface{}
	Variables     map[string]interface{}
	GetCredential func(uuid.UUID) (*models.CredentialData, error)
}

type Node interface {
	Type() string
	Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error)
}

type Registry struct {
	nodes map[string]Node
}

func NewRegistry() *Registry {
	r := &Registry{
		nodes: make(map[string]Node),
	}

	// Register built-in nodes
	// Triggers
	r.Register(&ManualTrigger{})
	r.Register(&WebhookTrigger{})
	r.Register(&ScheduleTrigger{})

	// Logic (stubs - see logic/ package for full implementations)
	r.Register(&ConditionNode{})
	r.Register(&SwitchNode{})
	r.Register(&LoopNode{})
	r.Register(&MergeNode{})
	r.Register(&WaitNode{})

	// Actions (stubs - see actions/ package for full implementations)
	r.Register(&HTTPRequestNode{})
	r.Register(&CodeNode{})
	r.Register(&SetVariableNode{})
	r.Register(&RespondNode{})

	// Integrations (stubs - see integrations/ package for full implementations)
	r.Register(&SlackNode{})
	r.Register(&EmailNode{})
	r.Register(&OpenAINode{})

	return r
}

func NewFullRegistry() *Registry {
	r := &Registry{
		nodes: make(map[string]Node),
	}

	// All nodes are registered via RegisterNode() calls from main
	return r
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
