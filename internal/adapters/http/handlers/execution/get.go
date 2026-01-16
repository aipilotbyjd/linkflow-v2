package execution

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	executionQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ExecutionResponse represents execution in responses
type ExecutionResponse struct {
	ID           string     `json:"id"`
	WorkflowID   string     `json:"workflow_id"`
	WorkspaceID  string     `json:"workspace_id"`
	WorkflowName string     `json:"workflow_name,omitempty"`
	Status       string     `json:"status"`
	TriggerType  string     `json:"trigger_type"`
	TriggeredBy  *string    `json:"triggered_by,omitempty"`
	InputData    types.JSON `json:"input_data,omitempty"`
	OutputData   types.JSON `json:"output_data,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	StartedAt    *string    `json:"started_at,omitempty"`
	CompletedAt  *string    `json:"completed_at,omitempty"`
	DurationMs   *int64     `json:"duration_ms,omitempty"`
	CreatedAt    string     `json:"created_at"`
}

// GetHandler handles getting a single execution
type GetHandler struct {
	handler *executionQuery.GetExecutionHandler
}

// NewGetHandler creates a new handler
func NewGetHandler(handler *executionQuery.GetExecutionHandler) *GetHandler {
	return &GetHandler{handler: handler}
}

// Handle handles the get execution request
func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
		return
	}

	exec, err := h.handler.Handle(r.Context(), executionQuery.GetExecutionQuery{
		ExecutionID: executionID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toExecutionResponse(exec))
}

func toExecutionResponse(e *execution.Execution) ExecutionResponse {
	resp := ExecutionResponse{
		ID:          e.ID.String(),
		WorkflowID:  e.WorkflowID.String(),
		WorkspaceID: e.WorkspaceID.String(),
		Status:      string(e.Status),
		TriggerType: e.TriggerType,
		InputData:   e.InputData,
		OutputData:  e.OutputData,
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if e.TriggeredBy != nil {
		s := e.TriggeredBy.String()
		resp.TriggeredBy = &s
	}
	if e.ErrorMessage != nil {
		resp.ErrorMessage = e.ErrorMessage
	}
	if e.StartedAt != nil {
		s := e.StartedAt.Format("2006-01-02T15:04:05Z")
		resp.StartedAt = &s
	}
	if e.CompletedAt != nil {
		s := e.CompletedAt.Format("2006-01-02T15:04:05Z")
		resp.CompletedAt = &s
	}
	durationMs := e.DurationMs()
	if durationMs > 0 {
		resp.DurationMs = &durationMs
	}

	return resp
}
