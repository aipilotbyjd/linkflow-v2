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

	// Data transformation nodes
	core.Register(&DataFilterNode{}, core.NodeMeta{
		Name:        "Data Filter",
		Description: "Filter array items with advanced conditions",
		Category:    "logic",
		Icon:        "filter",
		Version:     "1.0.0",
	})

	core.Register(&DataSortNode{}, core.NodeMeta{
		Name:        "Data Sort",
		Description: "Sort array items by field",
		Category:    "logic",
		Icon:        "sort-asc",
		Version:     "1.0.0",
	})

	core.Register(&DataLimitNode{}, core.NodeMeta{
		Name:        "Data Limit",
		Description: "Limit number of items with offset",
		Category:    "logic",
		Icon:        "list",
		Version:     "1.0.0",
	})

	core.Register(&RemoveDuplicatesNode{}, core.NodeMeta{
		Name:        "Remove Duplicates",
		Description: "Remove duplicate items from array",
		Category:    "logic",
		Icon:        "copy-slash",
		Version:     "1.0.0",
	})

	core.Register(&AggregateNode{}, core.NodeMeta{
		Name:        "Aggregate",
		Description: "Perform aggregations on array data",
		Category:    "logic",
		Icon:        "calculator",
		Version:     "1.0.0",
	})

	// Date/Time node
	core.Register(&DateTimeNode{}, core.NodeMeta{
		Name:        "Date & Time",
		Description: "Date and time operations",
		Category:    "logic",
		Icon:        "calendar",
		Version:     "1.0.0",
	})

	// HTML Extract node
	core.Register(&HTMLExtractNode{}, core.NodeMeta{
		Name:        "HTML Extract",
		Description: "Extract data from HTML content",
		Category:    "logic",
		Icon:        "code",
		Version:     "1.0.0",
	})

	// Crypto node
	core.Register(&CryptoNode{}, core.NodeMeta{
		Name:        "Crypto",
		Description: "Cryptographic operations (hash, encrypt, decrypt)",
		Category:    "logic",
		Icon:        "lock",
		Version:     "1.0.0",
	})

	// XML node
	core.Register(&XMLNode{}, core.NodeMeta{
		Name:        "XML",
		Description: "Parse and generate XML",
		Category:    "logic",
		Icon:        "file-code",
		Version:     "1.0.0",
	})

	// JSON Transform node
	core.Register(&JSONTransformNode{}, core.NodeMeta{
		Name:        "JSON Transform",
		Description: "Transform JSON data",
		Category:    "logic",
		Icon:        "braces",
		Version:     "1.0.0",
	})

	core.Register(&SplitDataNode{}, core.NodeMeta{
		Name:        "Split Data",
		Description: "Split data into batches",
		Category:    "logic",
		Icon:        "split",
		Version:     "1.0.0",
	})

	core.Register(&MergeDataNode{}, core.NodeMeta{
		Name:        "Merge Data",
		Description: "Merge multiple data inputs",
		Category:    "logic",
		Icon:        "merge",
		Version:     "1.0.0",
	})

	// Expression and Math nodes
	core.Register(&ExpressionNode{}, core.NodeMeta{
		Name:        "Expression",
		Description: "Evaluate expressions and calculations",
		Category:    "logic",
		Icon:        "function",
		Version:     "1.0.0",
	})

	core.Register(&MathNode{}, core.NodeMeta{
		Name:        "Math",
		Description: "Mathematical operations",
		Category:    "logic",
		Icon:        "calculator",
		Version:     "1.0.0",
	})
}
