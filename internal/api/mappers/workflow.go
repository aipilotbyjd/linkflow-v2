package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// WorkflowToResponse converts a Workflow model to WorkflowResponse DTO
func WorkflowToResponse(wf *models.Workflow) dto.WorkflowResponse {
	var lastExecutedAt *int64
	if wf.LastExecutedAt != nil {
		ts := wf.LastExecutedAt.Unix()
		lastExecutedAt = &ts
	}

	return dto.WorkflowResponse{
		ID:             wf.ID.String(),
		Name:           wf.Name,
		Description:    wf.Description,
		Status:         wf.Status,
		Version:        wf.Version,
		Nodes:          wf.Nodes,
		Connections:    wf.Connections,
		Settings:       wf.Settings,
		Tags:           wf.Tags,
		Color:          wf.Color,
		Icon:           wf.Icon,
		Category:       wf.Category,
		IsFavorite:     wf.IsFavorite,
		ExecutionCount: wf.ExecutionCount,
		LastExecutedAt: lastExecutedAt,
		CreatedAt:      wf.CreatedAt.Unix(),
		UpdatedAt:      wf.UpdatedAt.Unix(),
	}
}

// WorkflowToListResponse converts a Workflow model to a list response (without nodes/connections)
func WorkflowToListResponse(wf *models.Workflow) dto.WorkflowResponse {
	var lastExecutedAt *int64
	if wf.LastExecutedAt != nil {
		ts := wf.LastExecutedAt.Unix()
		lastExecutedAt = &ts
	}

	return dto.WorkflowResponse{
		ID:             wf.ID.String(),
		Name:           wf.Name,
		Description:    wf.Description,
		Status:         wf.Status,
		Version:        wf.Version,
		Tags:           wf.Tags,
		Color:          wf.Color,
		Icon:           wf.Icon,
		Category:       wf.Category,
		IsFavorite:     wf.IsFavorite,
		ExecutionCount: wf.ExecutionCount,
		LastExecutedAt: lastExecutedAt,
		CreatedAt:      wf.CreatedAt.Unix(),
		UpdatedAt:      wf.UpdatedAt.Unix(),
	}
}

// WorkflowsToResponse converts a slice of Workflow models to WorkflowResponse DTOs
func WorkflowsToResponse(workflows []models.Workflow) []dto.WorkflowResponse {
	result := make([]dto.WorkflowResponse, len(workflows))
	for i := range workflows {
		result[i] = WorkflowToListResponse(&workflows[i])
	}
	return result
}

// WorkflowVersionToResponse converts a WorkflowVersion model to WorkflowVersionResponse DTO
func WorkflowVersionToResponse(v *models.WorkflowVersion) dto.WorkflowVersionResponse {
	return dto.WorkflowVersionResponse{
		ID:            v.ID.String(),
		Version:       v.Version,
		Nodes:         v.Nodes,
		Connections:   v.Connections,
		Settings:      v.Settings,
		ChangeMessage: v.ChangeMessage,
		CreatedAt:     v.CreatedAt.Unix(),
	}
}

// WorkflowVersionsToResponse converts a slice of WorkflowVersion models to DTOs
func WorkflowVersionsToResponse(versions []models.WorkflowVersion) []dto.WorkflowVersionResponse {
	result := make([]dto.WorkflowVersionResponse, len(versions))
	for i := range versions {
		result[i] = WorkflowVersionToResponse(&versions[i])
	}
	return result
}
