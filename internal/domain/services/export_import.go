package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type WorkflowExportService struct {
	exportRepo   *repositories.BaseRepository[models.WorkflowExport]
	importRepo   *repositories.BaseRepository[models.WorkflowImport]
	workflowRepo *repositories.WorkflowRepository
}

// NewWorkflowExportService creates a new WorkflowExportService for exporting and importing workflows.
func NewWorkflowExportService(
	exportRepo *repositories.BaseRepository[models.WorkflowExport],
	importRepo *repositories.BaseRepository[models.WorkflowImport],
	workflowRepo *repositories.WorkflowRepository,
) *WorkflowExportService {
	return &WorkflowExportService{
		exportRepo:   exportRepo,
		importRepo:   importRepo,
		workflowRepo: workflowRepo,
	}
}

type WorkflowExportData struct {
	Version     string                   `json:"version"`
	ExportedAt  time.Time                `json:"exported_at"`
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	Nodes       []map[string]interface{} `json:"nodes"`
	Connections []map[string]interface{} `json:"connections"`
	Settings    map[string]interface{}   `json:"settings,omitempty"`
	Variables   []map[string]interface{} `json:"variables,omitempty"`
}

func (s *WorkflowExportService) Export(ctx context.Context, workflowID uuid.UUID, exportedBy uuid.UUID, includeCredentials bool) (*WorkflowExportData, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	// Convert JSONArray to []map[string]interface{}
	nodes := make([]map[string]interface{}, 0)
	for _, n := range workflow.Nodes {
		if m, ok := n.(map[string]interface{}); ok {
			nodes = append(nodes, m)
		}
	}

	connections := make([]map[string]interface{}, 0)
	for _, c := range workflow.Connections {
		if m, ok := c.(map[string]interface{}); ok {
			connections = append(connections, m)
		}
	}

	settings := map[string]interface{}(workflow.Settings)

	// Remove credential references if not including them
	if !includeCredentials {
		for i := range nodes {
			if params, ok := nodes[i]["parameters"].(map[string]interface{}); ok {
				delete(params, "credential_id")
			}
		}
	}

	data := &WorkflowExportData{
		Version:     "1.0",
		ExportedAt:  time.Now(),
		Name:        workflow.Name,
		Description: workflow.Description,
		Nodes:       nodes,
		Connections: connections,
		Settings:    settings,
	}

	// Log export
	exportLog := &models.WorkflowExport{
		WorkflowID:         workflowID,
		WorkspaceID:        workflow.WorkspaceID,
		ExportedBy:         exportedBy,
		Version:            workflow.Version,
		Format:             "json",
		IncludeCredentials: includeCredentials,
	}
	s.exportRepo.Create(ctx, exportLog)

	return data, nil
}

func (s *WorkflowExportService) Import(ctx context.Context, workspaceID, importedBy uuid.UUID, data *WorkflowExportData) (*models.Workflow, error) {
	// Convert []map[string]interface{} to JSONArray
	nodes := make(models.JSONArray, 0)
	for _, n := range data.Nodes {
		nodes = append(nodes, n)
	}

	connections := make(models.JSONArray, 0)
	for _, c := range data.Connections {
		connections = append(connections, c)
	}

	workflow := &models.Workflow{
		WorkspaceID: workspaceID,
		CreatedBy:   importedBy,
		Name:        data.Name + " (Imported)",
		Description: data.Description,
		Status:      "draft",
		Version:     1,
		Nodes:       nodes,
		Connections: connections,
		Settings:    data.Settings,
	}

	if err := s.workflowRepo.Create(ctx, workflow); err != nil {
		return nil, err
	}

	// Log import
	importLog := &models.WorkflowImport{
		WorkflowID:  &workflow.ID,
		WorkspaceID: workspaceID,
		ImportedBy:  importedBy,
		SourceName:  &data.Name,
		SourceType:  "file",
		Status:      "completed",
	}
	now := time.Now()
	importLog.CompletedAt = &now
	s.importRepo.Create(ctx, importLog)

	return workflow, nil
}
