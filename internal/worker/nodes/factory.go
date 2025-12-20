package nodes

import (
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/redis/go-redis/v9"
)

// Dependencies holds all external dependencies needed by nodes
type Dependencies struct {
	QueueClient *queue.Client
	RedisClient *redis.Client
}

// NodeFactory creates nodes with their dependencies
type NodeFactory struct {
	deps *Dependencies
}

// NewNodeFactory creates a new node factory with dependencies
func NewNodeFactory(deps *Dependencies) *NodeFactory {
	if deps == nil {
		deps = &Dependencies{}
	}
	return &NodeFactory{deps: deps}
}

// CreateRegistry creates a fully configured registry with all nodes
func (f *NodeFactory) CreateRegistry() *Registry {
	r := &Registry{
		nodes: make(map[string]Node),
	}

	// Register trigger nodes
	f.registerTriggers(r)

	// Register logic nodes
	f.registerLogicNodes(r)

	// Register action nodes
	f.registerActionNodes(r)

	// Register integration nodes
	f.registerIntegrationNodes(r)

	return r
}

func (f *NodeFactory) registerTriggers(r *Registry) {
	r.Register(&ManualTrigger{})
	r.Register(&WebhookTrigger{})
	r.Register(&ScheduleTrigger{})
}

func (f *NodeFactory) registerLogicNodes(r *Registry) {
	r.Register(&ConditionNode{})
	r.Register(&SwitchNode{})
	r.Register(&LoopNode{})
	r.Register(&MergeNode{})
	r.Register(&WaitNode{})
	r.Register(&FilterNode{})
	r.Register(&SortNode{})
	r.Register(&LimitNode{})
	r.Register(&UniqueNode{})
	r.Register(&SplitBatchesNode{})

	// Error handling nodes
	r.Register(&TryCatchNode{})
	r.Register(&RetryNode{})
	r.Register(&ThrowErrorNode{})
	r.Register(&ContinueOnFailNode{})
	r.Register(&TimeoutNode{})
	r.Register(&FallbackNode{})
}

func (f *NodeFactory) registerActionNodes(r *Registry) {
	r.Register(&HTTPRequestNode{})
	r.Register(&CodeNode{})
	r.Register(&SetVariableNode{})
	r.Register(&RespondNode{})
	r.Register(&FunctionNode{})
	r.Register(&TransformNode{})

	// Sub-workflow nodes (require dependencies)
	if f.deps.QueueClient != nil && f.deps.RedisClient != nil {
		r.Register(&SubWorkflowNode{
			queueClient: f.deps.QueueClient,
			redisClient: f.deps.RedisClient,
		})
		r.Register(&ExecuteWorkflowNode{
			queueClient: f.deps.QueueClient,
			redisClient: f.deps.RedisClient,
		})
	}
}

func (f *NodeFactory) registerIntegrationNodes(r *Registry) {
	r.Register(&SlackNode{})
	r.Register(&EmailNode{})
	r.Register(&OpenAINode{})
	r.Register(&GitHubNode{})
	r.Register(&DiscordNode{})
	r.Register(&TelegramNode{})
	r.Register(&PostgreSQLNode{})
	r.Register(&NotionNode{})
	r.Register(&AirtableNode{})
	r.Register(&AnthropicNode{})
}
