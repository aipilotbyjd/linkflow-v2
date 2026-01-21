package folder

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/folder"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type GetFolderTreeHandler struct {
	folderRepo   folder.Repository
	workflowRepo workflow.Repository
}

func NewGetFolderTreeHandler(folderRepo folder.Repository, workflowRepo workflow.Repository) *GetFolderTreeHandler {
	return &GetFolderTreeHandler{
		folderRepo:   folderRepo,
		workflowRepo: workflowRepo,
	}
}

type FolderTreeNode struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Color       *string           `json:"color,omitempty"`
	Children    []*FolderTreeNode `json:"children,omitempty"`
	Workflows   []WorkflowSummary `json:"workflows,omitempty"`
}

type WorkflowSummary struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
}

func (h *GetFolderTreeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	folders, _, err := h.folderRepo.FindByWorkspace(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	workflows, _, err := h.workflowRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Build folder map
	folderMap := make(map[uuid.UUID]*FolderTreeNode)
	for _, f := range folders {
		folderMap[f.ID] = &FolderTreeNode{
			ID:          f.ID,
			Name:        f.Name,
			Description: f.Description,
			Color:       f.Color,
			Children:    []*FolderTreeNode{},
			Workflows:   []WorkflowSummary{},
		}
	}

	// Assign workflows to folders
	workflowsByFolder := make(map[uuid.UUID][]WorkflowSummary)
	unfolderedWorkflows := []WorkflowSummary{}
	for _, wf := range workflows {
		summary := WorkflowSummary{
			ID:     wf.ID,
			Name:   wf.Name,
			Status: string(wf.Status),
		}
		if wf.FolderID != nil {
			workflowsByFolder[*wf.FolderID] = append(workflowsByFolder[*wf.FolderID], summary)
		} else {
			unfolderedWorkflows = append(unfolderedWorkflows, summary)
		}
	}

	for folderID, wfs := range workflowsByFolder {
		if node, ok := folderMap[folderID]; ok {
			node.Workflows = wfs
		}
	}

	// Build tree structure
	var rootNodes []*FolderTreeNode
	for _, f := range folders {
		node := folderMap[f.ID]
		if f.ParentID != nil {
			if parent, ok := folderMap[*f.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		} else {
			rootNodes = append(rootNodes, node)
		}
	}

	response := map[string]interface{}{
		"folders":   rootNodes,
		"workflows": unfolderedWorkflows,
	}

	common.Success(w, response)
}
