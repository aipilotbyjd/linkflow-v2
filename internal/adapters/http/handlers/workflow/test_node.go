package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type TestNodeHandler struct {
	nodeRegistry *nodes.Registry
	logger       logger.Logger
}

func NewTestNodeHandler(nodeRegistry *nodes.Registry, log logger.Logger) *TestNodeHandler {
	return &TestNodeHandler{
		nodeRegistry: nodeRegistry,
		logger:       log,
	}
}

func (h *TestNodeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	var req TestNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	// Get node from registry
	node, err := h.nodeRegistry.Get(req.NodeType)
	if err != nil {
		common.BadRequest(w, "unknown node type: "+req.NodeType)
		return
	}

	// Build node definition
	nodeDef := map[string]interface{}{
		"id":         "test_node",
		"type":       req.NodeType,
		"name":       "Test Node",
		"parameters": req.Parameters,
	}

	// Create minimal runtime for testing
	inputData := req.Input
	if inputData == nil {
		inputData = make(map[string]interface{})
	}

	// Create a mock execution and workflow for the runtime
	mockExec := &execution.Execution{
		ID:          uuid.New(),
		WorkspaceID: wsCtx.WorkspaceID,
		WorkflowID:  uuid.New(),
		InputData:   types.JSON(inputData),
	}
	mockWf := &workflow.Workflow{
		ID:          uuid.New(),
		WorkspaceID: wsCtx.WorkspaceID,
		Nodes:       types.JSONArray{nodeDef},
		Connections: types.JSONArray{},
	}

	runtime := executor.NewRuntime(mockExec, mockWf, h.logger)

	// Execute node with timeout
	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	output, execErr := node.Execute(ctx, runtime, nodeDef)
	duration := time.Since(start).Milliseconds()

	response := TestNodeResponse{
		DurationMs: duration,
	}

	if execErr != nil {
		response.Success = false
		response.Error = execErr.Error()
	} else {
		response.Success = true
		if output != nil {
			response.Output = output
		}
	}

	common.Success(w, response)
}
