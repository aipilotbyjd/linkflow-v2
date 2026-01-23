package replay

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/replay"
)

// CreateHandler handles creating a replay session
type CreateHandler struct{}

func NewCreateHandler() *CreateHandler {
	return &CreateHandler{}
}

func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)
	userID := middleware.GetUserID(ctx)

	var req CreateReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	execID, err := uuid.Parse(req.ExecutionID)
	if err != nil {
		common.BadRequest(w, "Invalid execution ID")
		return
	}

	// Parse mode
	var mode replay.ReplayMode
	switch req.Mode {
	case "full":
		mode = replay.ReplayModeFullReplay
	case "from_node":
		mode = replay.ReplayModeFromNode
	case "step":
		mode = replay.ReplayModeStepByStep
	case "breakpoint":
		mode = replay.ReplayModeToBreakpoint
	default:
		common.BadRequest(w, "Invalid replay mode")
		return
	}

	opts := replay.ReplayOptions{
		StartFromNodeID:  req.StartFromNodeID,
		EndAtNodeID:      req.EndAtNodeID,
		SkipNodes:        req.SkipNodes,
		ModifyInputs:     req.ModifyInputs,
		ModifyNodeParams: req.ModifyNodeParams,
		CaptureSnapshots: true,
	}

	// In production, fetch workflow ID from execution
	workflowID := uuid.New() // Placeholder

	session := replay.NewReplaySession(workspaceID, execID, workflowID, userID, mode, opts)

	for _, bp := range req.Breakpoints {
		session.AddBreakpoint(bp)
	}

	// In production, save to repository and start replay
	session.Start()

	common.Created(w, ToReplaySessionResponse(session))
}
