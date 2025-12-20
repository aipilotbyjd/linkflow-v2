package logic

import "github.com/linkflow-ai/linkflow/internal/worker/core"

func init() {
	// Register all logic nodes
	core.Register(&ConditionNode{}, core.NodeMeta{
		Name:        "IF",
		Description: "Route based on conditions",
		Category:    "logic",
		Icon:        "git-branch",
		Version:     "1.0.0",
	})

	core.Register(&SwitchNode{}, core.NodeMeta{
		Name:        "Switch",
		Description: "Route to multiple outputs based on value",
		Category:    "logic",
		Icon:        "shuffle",
		Version:     "1.0.0",
	})

	core.Register(&LoopNode{}, core.NodeMeta{
		Name:        "Loop",
		Description: "Iterate over items",
		Category:    "logic",
		Icon:        "repeat",
		Version:     "1.0.0",
	})

	core.Register(&MergeNode{}, core.NodeMeta{
		Name:        "Merge",
		Description: "Merge multiple inputs",
		Category:    "logic",
		Icon:        "git-merge",
		Version:     "1.0.0",
	})

	core.Register(&WaitNode{}, core.NodeMeta{
		Name:        "Wait",
		Description: "Pause execution for specified time",
		Category:    "logic",
		Icon:        "clock",
		Version:     "1.0.0",
	})

	// Error handling nodes
	core.Register(&TryCatchNode{}, core.NodeMeta{
		Name:        "Try/Catch",
		Description: "Handle errors in workflow",
		Category:    "logic",
		Icon:        "shield",
		Version:     "1.0.0",
	})

	core.Register(&RetryNode{}, core.NodeMeta{
		Name:        "Retry",
		Description: "Retry failed operations with backoff",
		Category:    "logic",
		Icon:        "refresh-cw",
		Version:     "1.0.0",
	})

	core.Register(&ThrowErrorNode{}, core.NodeMeta{
		Name:        "Throw Error",
		Description: "Throw a custom error",
		Category:    "logic",
		Icon:        "alert-triangle",
		Version:     "1.0.0",
	})

	core.Register(&ContinueOnFailNode{}, core.NodeMeta{
		Name:        "Continue On Fail",
		Description: "Continue workflow even if node fails",
		Category:    "logic",
		Icon:        "skip-forward",
		Version:     "1.0.0",
	})

	core.Register(&TimeoutNode{}, core.NodeMeta{
		Name:        "Timeout",
		Description: "Add timeout to operations",
		Category:    "logic",
		Icon:        "timer",
		Version:     "1.0.0",
	})

	core.Register(&FallbackNode{}, core.NodeMeta{
		Name:        "Fallback",
		Description: "Provide fallback value on error",
		Category:    "logic",
		Icon:        "life-buoy",
		Version:     "1.0.0",
	})
}
