package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

func ExecutionToModel(e *execution.Execution) *models.Execution {
	return &models.Execution{
		ID:                e.ID,
		WorkflowID:        e.WorkflowID,
		WorkspaceID:       e.WorkspaceID,
		TriggeredBy:       e.TriggeredBy,
		WorkflowVersion:   e.WorkflowVersion,
		Status:            string(e.Status),
		TriggerType:       e.TriggerType,
		TriggerData:       e.TriggerData,
		InputData:         e.InputData,
		OutputData:        e.OutputData,
		ErrorMessage:      e.ErrorMessage,
		ErrorNodeID:       e.ErrorNodeID,
		QueuedAt:          e.QueuedAt,
		StartedAt:         e.StartedAt,
		CompletedAt:       e.CompletedAt,
		PausedAt:          e.PausedAt,
		ResumedAt:         e.ResumedAt,
		NodesTotal:        e.NodesTotal,
		NodesCompleted:    e.NodesCompleted,
		RetryCount:        e.RetryCount,
		MaxRetries:        e.MaxRetries,
		Priority:          e.Priority,
		TimeoutSeconds:    e.TimeoutSeconds,
		ParentExecutionID: e.ParentExecutionID,
		BatchID:           e.BatchID,
		CreatedAt:         e.CreatedAt,
	}
}

func ExecutionToDomain(m *models.Execution) *execution.Execution {
	return &execution.Execution{
		ID:                m.ID,
		WorkflowID:        m.WorkflowID,
		WorkspaceID:       m.WorkspaceID,
		TriggeredBy:       m.TriggeredBy,
		WorkflowVersion:   m.WorkflowVersion,
		Status:            execution.Status(m.Status),
		TriggerType:       m.TriggerType,
		TriggerData:       m.TriggerData,
		InputData:         m.InputData,
		OutputData:        m.OutputData,
		ErrorMessage:      m.ErrorMessage,
		ErrorNodeID:       m.ErrorNodeID,
		QueuedAt:          m.QueuedAt,
		StartedAt:         m.StartedAt,
		CompletedAt:       m.CompletedAt,
		PausedAt:          m.PausedAt,
		ResumedAt:         m.ResumedAt,
		NodesTotal:        m.NodesTotal,
		NodesCompleted:    m.NodesCompleted,
		RetryCount:        m.RetryCount,
		MaxRetries:        m.MaxRetries,
		Priority:          m.Priority,
		TimeoutSeconds:    m.TimeoutSeconds,
		ParentExecutionID: m.ParentExecutionID,
		BatchID:           m.BatchID,
		CreatedAt:         m.CreatedAt,
	}
}

func NodeExecutionToModel(n *execution.NodeExecution) *models.NodeExecution {
	var durationMs *int
	if n.StartedAt != nil && n.CompletedAt != nil {
		ms := int(n.CompletedAt.Sub(*n.StartedAt).Milliseconds())
		durationMs = &ms
	}

	return &models.NodeExecution{
		ID:           n.ID,
		ExecutionID:  n.ExecutionID,
		NodeID:       n.NodeID,
		NodeType:     n.NodeType,
		NodeName:     n.NodeName,
		Status:       string(n.Status),
		InputData:    n.InputData,
		OutputData:   n.OutputData,
		ErrorMessage: n.ErrorMessage,
		StartedAt:    n.StartedAt,
		CompletedAt:  n.CompletedAt,
		DurationMs:   durationMs,
		RetryCount:   n.RetryCount,
		CreatedAt:    n.CreatedAt,
	}
}

func NodeExecutionToDomain(m *models.NodeExecution) *execution.NodeExecution {
	return &execution.NodeExecution{
		ID:           m.ID,
		ExecutionID:  m.ExecutionID,
		NodeID:       m.NodeID,
		NodeType:     m.NodeType,
		NodeName:     m.NodeName,
		Status:       execution.NodeStatus(m.Status),
		InputData:    m.InputData,
		OutputData:   m.OutputData,
		ErrorMessage: m.ErrorMessage,
		StartedAt:    m.StartedAt,
		CompletedAt:  m.CompletedAt,
		RetryCount:   m.RetryCount,
		CreatedAt:    m.CreatedAt,
	}
}
