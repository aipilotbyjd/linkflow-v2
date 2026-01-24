package executor

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type NodeHandler interface {
	Execute(ctx context.Context, runtime *Runtime, node map[string]interface{}) (types.JSON, error)
}

type Processor struct {
	handlers map[string]NodeHandler
	logger   logger.Logger
}

func NewProcessor(log logger.Logger) *Processor {
	return &Processor{
		handlers: make(map[string]NodeHandler),
		logger:   log,
	}
}

func (p *Processor) RegisterHandler(nodeType string, handler NodeHandler) {
	p.handlers[nodeType] = handler
}

func (p *Processor) Process(ctx context.Context, runtime *Runtime, node map[string]interface{}) (types.JSON, error) {
	nodeType, ok := node["type"].(string)
	if !ok {
		return nil, fmt.Errorf("node type not specified")
	}

	handler, ok := p.handlers[nodeType]
	if !ok {
		return nil, fmt.Errorf("no handler registered for node type: %s", nodeType)
	}

	p.logger.Debug().
		Str("node_type", nodeType).
		Str("node_id", node["id"].(string)).
		Msg("Processing node")

	// Create a shallow copy of node to modify parameters without affecting the original in workflow
	nodeCopy := make(map[string]interface{})
	for k, v := range node {
		nodeCopy[k] = v
	}

	// Resolve parameters
	if params, ok := node["parameters"].(map[string]interface{}); ok {
		nodeCopy["parameters"] = runtime.ResolveParameters(params)
	}

	return handler.Execute(ctx, runtime, nodeCopy)
}

func (p *Processor) HasHandler(nodeType string) bool {
	_, ok := p.handlers[nodeType]
	return ok
}
