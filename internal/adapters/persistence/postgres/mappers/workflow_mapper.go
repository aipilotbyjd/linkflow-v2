package mappers

import (
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

func WorkflowToModel(wf *workflow.Workflow) *models.Workflow {
	return &models.Workflow{
		ID:             wf.ID,
		WorkspaceID:    wf.WorkspaceID,
		CreatedBy:      wf.CreatedBy,
		Name:           wf.Name,
		Description:    wf.Description,
		Status:         string(wf.Status),
		Version:        wf.Version,
		Nodes:          wf.Nodes,
		Connections:    wf.Connections,
		Settings:       wf.Settings,
		Tags:           wf.Tags,
		ExecutionCount: wf.ExecutionCount,
		LastExecutedAt: wf.LastExecutedAt,
		ActivatedAt:    wf.ActivatedAt,
		CreatedAt:      wf.CreatedAt,
		UpdatedAt:      wf.UpdatedAt,
	}
}

func WorkflowToDomain(m *models.Workflow) *workflow.Workflow {
	return &workflow.Workflow{
		ID:             m.ID,
		WorkspaceID:    m.WorkspaceID,
		CreatedBy:      m.CreatedBy,
		Name:           m.Name,
		Description:    m.Description,
		Status:         workflow.Status(m.Status),
		Version:        m.Version,
		Nodes:          m.Nodes,
		Connections:    m.Connections,
		Settings:       m.Settings,
		Tags:           m.Tags,
		ExecutionCount: m.ExecutionCount,
		LastExecutedAt: m.LastExecutedAt,
		ActivatedAt:    m.ActivatedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func WorkflowVersionToModel(v *workflow.Version) *models.WorkflowVersion {
	var createdBy uuid.UUID
	if v.CreatedBy != nil {
		createdBy = *v.CreatedBy
	}
	return &models.WorkflowVersion{
		ID:          v.ID,
		WorkflowID:  v.WorkflowID,
		Version:     v.Version,
		Nodes:       v.Nodes,
		Connections: v.Connections,
		Settings:    v.Settings,
		ChangeNote:  v.ChangeMessage,
		CreatedBy:   createdBy,
		CreatedAt:   v.CreatedAt,
	}
}

func WorkflowVersionToDomain(m *models.WorkflowVersion) *workflow.Version {
	return &workflow.Version{
		ID:            m.ID,
		WorkflowID:    m.WorkflowID,
		Version:       m.Version,
		Nodes:         m.Nodes,
		Connections:   m.Connections,
		Settings:      m.Settings,
		ChangeMessage: m.ChangeNote,
		CreatedBy:     &m.CreatedBy,
		CreatedAt:     m.CreatedAt,
	}
}
