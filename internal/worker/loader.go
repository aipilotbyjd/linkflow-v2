package worker

// Import all node packages to trigger their init() functions
// This must be done here (not in nodes package) to avoid import cycles

import (
	_ "github.com/linkflow-ai/linkflow/internal/worker/nodes/actions"
	_ "github.com/linkflow-ai/linkflow/internal/worker/nodes/integrations"
	_ "github.com/linkflow-ai/linkflow/internal/worker/nodes/logic"
	_ "github.com/linkflow-ai/linkflow/internal/worker/nodes/triggers"
)
