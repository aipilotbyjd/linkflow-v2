package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type ExecutionReplayHandler struct {
	replaySvc *services.ExecutionReplayService
}

func NewExecutionReplayHandler(replaySvc *services.ExecutionReplayService) *ExecutionReplayHandler {
	return &ExecutionReplayHandler{replaySvc: replaySvc}
}

func (h *ExecutionReplayHandler) Replay(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	execID, ok := middleware.ParseUUID(w, r, "executionID")
	if !ok {
		return
	}

	execution, err := h.replaySvc.Replay(r.Context(), execID, &claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to replay execution")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":      execution.ID,
		"status":  execution.Status,
		"message": "execution replay started",
	})
}

func (h *ExecutionReplayHandler) ReplayFromNode(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	execID, ok := middleware.ParseUUID(w, r, "executionID")
	if !ok {
		return
	}

	var req struct {
		NodeID string `json:"node_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	execution, err := h.replaySvc.ReplayFromNode(r.Context(), execID, req.NodeID, &claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to replay execution from node")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":        execution.ID,
		"status":    execution.Status,
		"from_node": req.NodeID,
		"message":   "partial execution replay started",
	})
}
