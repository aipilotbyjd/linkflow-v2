package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// ExecutionToResponse converts an Execution model to ExecutionResponse DTO
func ExecutionToResponse(e *models.Execution) dto.ExecutionResponse {
	var startedAt, completedAt *int64
	if e.StartedAt != nil {
		ts := e.StartedAt.Unix()
		startedAt = &ts
	}
	if e.CompletedAt != nil {
		ts := e.CompletedAt.Unix()
		completedAt = &ts
	}

	return dto.ExecutionResponse{
		ID:              e.ID.String(),
		WorkflowID:      e.WorkflowID.String(),
		WorkflowVersion: e.WorkflowVersion,
		Status:          e.Status,
		TriggerType:     e.TriggerType,
		InputData:       e.InputData,
		OutputData:      e.OutputData,
		ErrorMessage:    e.ErrorMessage,
		ErrorNodeID:     e.ErrorNodeID,
		NodesTotal:      e.NodesTotal,
		NodesCompleted:  e.NodesCompleted,
		QueuedAt:        e.QueuedAt.Unix(),
		StartedAt:       startedAt,
		CompletedAt:     completedAt,
	}
}

// ExecutionsToResponse converts a slice of Execution models to ExecutionResponse DTOs
func ExecutionsToResponse(executions []models.Execution) []dto.ExecutionResponse {
	result := make([]dto.ExecutionResponse, len(executions))
	for i := range executions {
		result[i] = ExecutionToResponse(&executions[i])
	}
	return result
}

// NodeExecutionToResponse converts a NodeExecution model to NodeExecutionResponse DTO
func NodeExecutionToResponse(n *models.NodeExecution) dto.NodeExecutionResponse {
	var startedAt, completedAt *int64
	var durationMs *int
	if n.StartedAt != nil {
		ts := n.StartedAt.Unix()
		startedAt = &ts
	}
	if n.CompletedAt != nil {
		ts := n.CompletedAt.Unix()
		completedAt = &ts
	}
	if n.DurationMs != nil {
		durationMs = n.DurationMs
	}

	return dto.NodeExecutionResponse{
		ID:           n.ID.String(),
		NodeID:       n.NodeID,
		NodeType:     n.NodeType,
		NodeName:     n.NodeName,
		Status:       n.Status,
		InputData:    n.InputData,
		OutputData:   n.OutputData,
		ErrorMessage: n.ErrorMessage,
		DurationMs:   durationMs,
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
	}
}

// NodeExecutionsToResponse converts a slice of NodeExecution models to DTOs
func NodeExecutionsToResponse(nodes []models.NodeExecution) []dto.NodeExecutionResponse {
	result := make([]dto.NodeExecutionResponse, len(nodes))
	for i := range nodes {
		result[i] = NodeExecutionToResponse(&nodes[i])
	}
	return result
}
